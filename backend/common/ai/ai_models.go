package ai

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type LawReference struct {
	LawName    string  `json:"lawName"`
	ArticleNo  string  `json:"articleNo"`
	Content    string  `json:"content"`
	Similarity float64 `json:"similarity"`
}

type AIAnswer struct {
	Answer          string         `json:"answer"`
	RelatedArticles []LawReference `json:"relatedArticles"`
	References      []string       `json:"references"`
	Keywords        []string       `json:"keywords"`
}

type RiskResult struct {
	RiskLevel    string   `json:"riskLevel"`
	RiskFactors  []string `json:"riskFactors"`
	Suggestions  []string `json:"suggestions"`
	RiskScore    float64  `json:"riskScore"`
}

type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type ChatCompletionStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int         `json:"index"`
		Delta        ChatMessage `json:"delta"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

type EmbeddingRequest struct {
	Model string          `json:"model"`
	Input []string        `json:"input"`
}

type EmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}
