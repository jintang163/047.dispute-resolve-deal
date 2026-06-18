package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

type DeepSeekClient struct {
	apiKey      string
	baseURL     string
	model       string
	maxTokens   int
	temperature float64
	httpClient  *http.Client
}

var deepSeekClient *DeepSeekClient

func NewDeepSeekClient() *DeepSeekClient {
	cfg := config.GetConfig()
	return &DeepSeekClient{
		apiKey:      cfg.DeepSeek.APIKey,
		baseURL:     cfg.DeepSeek.BaseURL,
		model:       cfg.DeepSeek.Model,
		maxTokens:   cfg.DeepSeek.MaxTokens,
		temperature: cfg.DeepSeek.Temperature,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func InitDeepSeek() {
	deepSeekClient = NewDeepSeekClient()
	logger.Info("DeepSeek client initialized",
		zap.String("baseURL", deepSeekClient.baseURL),
		zap.String("model", deepSeekClient.model),
	)
}

func GetDeepSeekClient() *DeepSeekClient {
	if deepSeekClient == nil {
		InitDeepSeek()
	}
	return deepSeekClient
}

func (c *DeepSeekClient) ChatCompletion(messages []ChatMessage, systemPrompt string) (string, error) {
	return c.ChatCompletionWithModel(messages, systemPrompt, c.model)
}

func (c *DeepSeekClient) ChatCompletionWithModel(messages []ChatMessage, systemPrompt string, model string) (string, error) {
	fullMessages := make([]ChatMessage, 0)
	if systemPrompt != "" {
		fullMessages = append(fullMessages, ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	fullMessages = append(fullMessages, messages...)

	reqBody := ChatCompletionRequest{
		Model:       model,
		Messages:    fullMessages,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		Stream:      false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	logger.Debug("DeepSeek chat request",
		zap.String("model", model),
		zap.Int("messageCount", len(fullMessages)),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("deepseek api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Error("DeepSeek API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return "", fmt.Errorf("deepseek api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %w, body=%s", err, string(respBody))
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	result := chatResp.Choices[0].Message.Content

	logger.Debug("DeepSeek chat response",
		zap.Int("promptTokens", chatResp.Usage.PromptTokens),
		zap.Int("completionTokens", chatResp.Usage.CompletionTokens),
		zap.Int("totalTokens", chatResp.Usage.TotalTokens),
	)

	return result, nil
}

func (c *DeepSeekClient) ChatCompletionStream(messages []ChatMessage, systemPrompt string, callback func(string)) error {
	fullMessages := make([]ChatMessage, 0)
	if systemPrompt != "" {
		fullMessages = append(fullMessages, ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		})
	}
	fullMessages = append(fullMessages, messages...)

	reqBody := ChatCompletionRequest{
		Model:       c.model,
		Messages:    fullMessages,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
		Stream:      true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request failed: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	logger.Debug("DeepSeek stream chat request",
		zap.String("model", c.model),
		zap.Int("messageCount", len(fullMessages)),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("deepseek stream request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("deepseek stream api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				break
			}

			var streamResp ChatCompletionStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				logger.Warn("Unmarshal stream data failed",
					zap.String("data", data),
					logger.Error(err),
				)
				continue
			}

			if len(streamResp.Choices) > 0 {
				if callback != nil {
					callback(streamResp.Choices[0].Delta.Content)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stream failed: %w", err)
	}

	return nil
}

func (c *DeepSeekClient) GetEmbedding(text string) ([]float32, error) {
	embeddings, err := c.GetBatchEmbeddings([]string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return embeddings[0], nil
}

func (c *DeepSeekClient) GetBatchEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("empty input texts")
	}

	reqBody := EmbeddingRequest{
		Model: "deepseek-embed",
		Input: texts,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request failed: %w", err)
	}

	url := c.baseURL + "/embeddings"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("create embedding request failed: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	logger.Debug("DeepSeek embedding request",
		zap.Int("textCount", len(texts)),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embedding api request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read embedding response failed: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Error("DeepSeek Embedding API error",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(respBody)),
		)
		return nil, fmt.Errorf("embedding api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var embedResp EmbeddingResponse
	if err := json.Unmarshal(respBody, &embedResp); err != nil {
		return nil, fmt.Errorf("unmarshal embedding response failed: %w", err)
	}

	results := make([][]float32, len(texts))
	for _, d := range embedResp.Data {
		if d.Index >= 0 && d.Index < len(texts) {
			results[d.Index] = d.Embedding
		}
	}

	logger.Debug("DeepSeek embedding response",
		zap.Int("totalTokens", embedResp.Usage.TotalTokens),
		zap.Int("embeddingCount", len(embedResp.Data)),
	)

	return results, nil
}
