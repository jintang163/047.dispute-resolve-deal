package vector

import (
	"context"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"go.uber.org/zap"
)

const (
	DefaultShardsNum    = 2
	DefaultIndexType    = "IVF_FLAT"
	DefaultMetricType   = "L2"
	DefaultNlist        = 1024
	DefaultNprobe       = 10
	DefaultTimeout      = 60
)

var (
	milvusClient client.Client
	collectionName string
	dim             int
)

func InitMilvus() error {
	cfg := config.GetConfig()
	collectionName = cfg.Milvus.Collection
	dim = cfg.Milvus.Dimension
	SetVectorDimension(dim)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	var err error
	milvusClient, err = client.NewGrpcClient(ctx, cfg.Milvus.Address,
		client.WithUsername(cfg.Milvus.Username),
		client.WithPassword(cfg.Milvus.Password),
	)
	if err != nil {
		logger.Error("Connect to milvus failed",
			zap.String("address", cfg.Milvus.Address),
			logger.Error(err),
		)
		return fmt.Errorf("连接Milvus失败: %w", err)
	}

	logger.Info("Connected to milvus",
		zap.String("address", cfg.Milvus.Address),
		zap.String("collection", collectionName),
		zap.Int("dimension", dim),
	)

	if err := CreateCollectionIfNotExists(); err != nil {
		logger.Warn("Create collection failed", logger.Error(err))
		return err
	}

	return nil
}

func GetMilvusClient() client.Client {
	return milvusClient
}

func CreateCollectionIfNotExists() error {
	if milvusClient == nil {
		return fmt.Errorf("milvus client not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	exists, err := milvusClient.HasCollection(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("check collection exists failed: %w", err)
	}

	if exists {
		logger.Info("Collection already exists", zap.String("collection", collectionName))
		return loadCollectionIfNeeded()
	}

	schema := &entity.Schema{
		CollectionName: collectionName,
		Description:  "Law article embeddings collection",
		AutoID:       false,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeInt64,
				PrimaryKey: true,
				AutoID:     false,
				Description: "Primary key ID",
			},
			{
				Name:       "vector",
				DataType:   entity.FieldTypeFloatVector,
				TypeParams: map[string]string{entity.TypeParamDim: fmt.Sprintf("%d", dim)},
				Description: "Vector embedding",
			},
			{
				Name:       "law_id",
				DataType:   entity.FieldTypeInt64,
				Description: "Law article ID",
			},
			{
				Name:       "content",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{entity.TypeParamMaxLength: "5000"},
				Description: "Law article content",
			},
			{
				Name:       "keywords",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{entity.TypeParamMaxLength: "500"},
				Description: "Keywords",
			},
		},
	}

	err = milvusClient.CreateCollection(ctx, schema, DefaultShardsNum)
	if err != nil {
		logger.Error("Create collection failed",
			zap.String("collection", collectionName),
			logger.Error(err),
		)
		return fmt.Errorf("创建集合失败: %w", err)
	}

	err = createIndex()
	if err != nil {
		logger.Error("Create index failed", logger.Error(err))
		return fmt.Errorf("创建索引失败: %w", err)
	}

	logger.Info("Collection created successfully", zap.String("collection", collectionName))

	time.Sleep(2 * time.Second)

	return loadCollectionIfNeeded()
}

func createIndex() error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	idx, err := entity.NewIndexIvfFlat(entity.L2, DefaultNlist)
	if err != nil {
		return fmt.Errorf("create index entity failed: %w", err)
	}

	err = milvusClient.CreateIndex(ctx, collectionName, "vector", idx, false)
	if err != nil {
		return fmt.Errorf("create index on collection failed: %w", err)
	}

	logger.Info("Index created",
		zap.String("collection", collectionName),
		zap.String("indexType", DefaultIndexType),
	)

	return nil
}

func loadCollectionIfNeeded() error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	err := milvusClient.LoadCollection(ctx, collectionName, false)
	if err != nil {
		logger.Warn("Load collection failed", zap.String("collection", collectionName), logger.Error(err))
		return nil
	}

	logger.Info("Collection loaded", zap.String("collection", collectionName))
	return nil
}

func InsertVectors(ids []int64, vectors [][]float32, metadata []map[string]interface{}) error {
	if milvusClient == nil {
		if err := InitMilvus(); err != nil {
			return err
		}
	}

	if len(ids) != len(vectors) || len(ids) != len(metadata) {
		return fmt.Errorf("ids, vectors and metadata length mismatch")
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	idColumn := entity.NewColumnInt64("id", ids)

	vectorColumn := entity.NewColumnFloatVector("vector", dim, vectors)

	lawIDs := make([]int64, len(metadata))
	contents := make([]string, len(metadata))
	keywords := make([]string, len(metadata))

	for i, meta := range metadata {
		if v, ok := meta["law_id"].(int64); ok {
			lawIDs[i] = v
		}
		if v, ok := meta["lawId"].(int64); ok {
			lawIDs[i] = v
		}
		if v, ok := meta["content"].(string); ok {
			contents[i] = v
		}
		if v, ok := meta["keywords"].(string); ok {
			keywords[i] = v
		}
	}

	lawIDColumn := entity.NewColumnInt64("law_id", lawIDs)
	contentColumn := entity.NewColumnVarChar("content", contents)
	keywordColumn := entity.NewColumnVarChar("keywords", keywords)

	columns := []entity.Column{
		idColumn,
		vectorColumn,
		lawIDColumn,
		contentColumn,
		keywordColumn,
	}

	resultIDs, err := milvusClient.Insert(ctx, collectionName, "", columns...)
	if err != nil {
		logger.Error("Insert vectors failed",
			zap.Int("count", len(ids)),
			logger.Error(err),
		)
		return fmt.Errorf("插入向量失败: %w", err)
	}

	logger.Info("Vectors inserted",
		zap.Int("count", len(ids)),
		zap.Any("resultIDs", resultIDs),
	)

	return nil
}

func SearchVectors(queryVector []float32, topK int, filter string) ([]*SearchResult, error) {
	if milvusClient == nil {
		if err := InitMilvus(); err != nil {
			return nil, err
		}
	}

	if len(queryVector) == 0 {
		return nil, fmt.Errorf("查询向量不能为空")
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	if topK <= 0 {
		topK = 10
	}

	outputFields := []string{"id", "law_id", "content", "keywords"}

	vectors := [][]float32{queryVector}

	sp, _ := entity.NewIndexIvfFlatSearchParam(DefaultNprobe)

	results, err := milvusClient.Search(
		ctx,
		collectionName,
		nil,
		filter,
		outputFields,
		vectors,
		"vector",
		entity.L2,
		topK,
		sp,
	)
	if err != nil {
		logger.Error("Search vectors failed",
			zap.Int("topK", topK),
			logger.Error(err),
		)
		return fmt.Errorf("搜索向量失败: %w", err)
	}

	if len(results) == 0 {
		return []*SearchResult{}, nil
	}

	result := results[0]
	searchResults := make([]*SearchResult, result.ResultCount)

	idData := result.IDs
	distances := result.Scores

	var lawIDData, contentData, keywordData entity.Column
	for _, field := range outputFields {
		for _, col := range result.Fields {
			if col.Name() == field {
				switch field {
				case "law_id":
					lawIDData = col
				case "content":
					contentData = col
				case "keywords":
					keywordData = col
				}
			}
		}
	}

	for i := 0; i < int(result.ResultCount); i++ {
		sr := &SearchResult{
			Score:    convertScore(distances[i]),
			Distance: distances[i],
		}

		if idCol, ok := idData.(*entity.ColumnInt64); ok {
			if idCol.Len() > i {
				sr.ID = idCol.Data()[i]
			}
		}

		if lawIDData != nil {
			if col, ok := lawIDData.(*entity.ColumnInt64); ok {
				if col.Len() > i {
					sr.LawID = col.Data()[i]
				}
			}
		}

		if contentData != nil {
			if col, ok := contentData.(*entity.ColumnVarChar); ok {
				if col.Len() > i {
					sr.Content = col.Data()[i]
				}
			}
		}

		if keywordData != nil {
			if col, ok := keywordData.(*entity.ColumnVarChar); ok {
				if col.Len() > i {
					sr.Keywords = col.Data()[i]
				}
			}
		}

		searchResults[i] = sr
	}

	searchResults = SortByScore(searchResults)

	logger.Debug("Vector search completed",
		zap.Int("queryDimension", len(queryVector)),
		zap.Int("resultCount", len(searchResults)),
	)

	return searchResults, nil
}

func DeleteVectors(ids []int64) error {
	if milvusClient == nil {
		if err := InitMilvus(); err != nil {
			return err
		}
	}

	if len(ids) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	idStr := make([]string, len(ids))
	for i, id := range ids {
		idStr[i] = fmt.Sprintf("%d", id)
	}

	expr := fmt.Sprintf("id in [%s]", joinInt64List(ids))

	err := milvusClient.Delete(ctx, collectionName, "", expr)
	if err != nil {
		logger.Error("Delete vectors failed",
			zap.Int("count", len(ids)),
			logger.Error(err),
		)
		return fmt.Errorf("删除向量失败: %w", err)
	}

	logger.Info("Vectors deleted", zap.Int("count", len(ids)))

	return nil
}

func DeleteByLawID(lawID int64) error {
	expr := fmt.Sprintf("law_id == %d", lawID)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	err := milvusClient.Delete(ctx, collectionName, "", expr)
	if err != nil {
		logger.Error("Delete by law_id failed",
			zap.Int64("lawID", lawID),
			logger.Error(err),
		)
		return fmt.Errorf("按法条ID删除向量失败: %w", err)
	}

	logger.Info("Vectors deleted by law_id", zap.Int64("lawID", lawID))

	return nil
}

func convertScore(distance float32) float32 {
	return float32(1.0 / (1.0 + distance))
}

func joinInt64List(ids []int64) string {
	if len(ids) == 0 {
		return ""
	}
	result := fmt.Sprintf("%d", ids[0])
	for i := 1; i < len(ids); i++ {
		result += fmt.Sprintf(",%d", ids[i])
	}
	return result
}
