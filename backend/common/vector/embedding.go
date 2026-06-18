package vector

import (
	"fmt"
	"math"

	"github.com/dispute-resolve/common/ai"
	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

var vectorDimension int

func SetVectorDimension(dim int) {
	vectorDimension = dim
}

func GetVectorDimension() int {
	return vectorDimension
}

func GetEmbedding(text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("输入文本不能为空")
	}

	client := ai.GetDeepSeekClient()

	embedding, err := client.GetEmbedding(text)
	if err != nil {
		logger.Error("Get embedding failed",
			zap.String("text", truncateText(text, 50)),
			logger.Error(err),
		)
		return nil, fmt.Errorf("获取文本向量失败: %w", err)
	}

	if len(embedding) == 0 {
		return nil, fmt.Errorf("返回的向量为空")
	}

	if vectorDimension > 0 && len(embedding) != vectorDimension {
		logger.Warn("Embedding dimension mismatch",
			zap.Int("expected", vectorDimension),
			zap.Int("actual", len(embedding)),
		)
	}

	logger.Debug("Get embedding success",
		zap.Int("dimension", len(embedding)),
		zap.String("text", truncateText(text, 30)),
	)

	return embedding, nil
}

func GetBatchEmbeddings(texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("输入文本列表不能为空")
	}

	client := ai.GetDeepSeekClient()

	embeddings, err := client.GetBatchEmbeddings(texts)
	if err != nil {
		logger.Error("Get batch embeddings failed",
			zap.Int("textCount", len(texts)),
			logger.Error(err),
		)
		return nil, fmt.Errorf("批量获取向量失败: %w", err)
	}

	if len(embeddings) != len(texts) {
		logger.Warn("Embedding count mismatch",
			zap.Int("expected", len(texts)),
			zap.Int("actual", len(embeddings)),
		)
	}

	validCount := 0
	for _, e := range embeddings {
		if len(e) > 0 {
			validCount++
		}
	}

	logger.Info("Batch embeddings completed",
		zap.Int("total", len(texts)),
		zap.Int("valid", validCount),
	)

	return embeddings, nil
}

func NormalizeVector(vec []float32) []float32 {
	if len(vec) == 0 {
		return vec
	}

	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	if norm == 0 {
		return vec
	}

	norm = sqrtFloat32(norm)
	normalized := make([]float32, len(vec))
	for i, v := range vec {
		normalized[i] = v / norm
	}
	return normalized
}

func truncateText(text string, maxLen int) string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return text
	}
	return string(runes[:maxLen]) + "..."
}

func sqrtFloat32(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}
