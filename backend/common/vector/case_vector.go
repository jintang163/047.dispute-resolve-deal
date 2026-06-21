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

type CaseSearchResult struct {
	ID              int64   `json:"id"`
	Score           float32 `json:"score"`
	CaseID          int64   `json:"caseId"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	DisputeType     string  `json:"disputeType"`
	MediationTactics string  `json:"mediationTactics"`
	KeyPoints       string  `json:"keyPoints"`
	Keywords        string  `json:"keywords"`
	DifficultyLevel int     `json:"difficultyLevel"`
	IsSuccess       int32   `json:"isSuccess"`
	Distance        float32 `json:"distance"`
}

var (
	caseCollectionName string
	caseDim            int
)

func InitCaseCollection() error {
	cfg := config.GetConfig()
	caseCollectionName = cfg.Milvus.CaseCollection
	caseDim = cfg.Milvus.CaseDimension
	if caseDim <= 0 {
		caseDim = cfg.Milvus.Dimension
	}

	if caseCollectionName == "" {
		caseCollectionName = "case_library"
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	milvusCl := GetMilvusClient()
	if milvusCl == nil {
		return fmt.Errorf("milvus client not initialized")
	}

	exists, err := milvusCl.HasCollection(ctx, caseCollectionName)
	if err != nil {
		return fmt.Errorf("check case collection exists failed: %w", err)
	}

	if exists {
		logger.Info("Case collection already exists", zap.String("collection", caseCollectionName))
		return loadCaseCollectionIfNeeded(milvusCl)
	}

	schema := &entity.Schema{
		CollectionName: caseCollectionName,
		Description:    "Case library embeddings collection",
		AutoID:         false,
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
				TypeParams: map[string]string{entity.TypeParamDim: fmt.Sprintf("%d", caseDim)},
				Description: "Vector embedding",
			},
			{
				Name:       "case_id",
				DataType:   entity.FieldTypeInt64,
				Description: "Case library ID",
			},
			{
				Name:       "title",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{entity.TypeParamMaxLength: "500"},
				Description: "Case title",
			},
			{
				Name:       "description",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{entity.TypeParamMaxLength: "5000"},
				Description: "Case description",
			},
			{
				Name:       "dispute_type",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{entity.TypeParamMaxLength: "200"},
				Description: "Dispute type",
			},
			{
				Name:       "mediation_tactics",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{entity.TypeParamMaxLength: "5000"},
				Description: "Mediation tactics",
			},
			{
				Name:       "key_points",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{entity.TypeParamMaxLength: "5000"},
				Description: "Key points",
			},
			{
				Name:       "keywords",
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{entity.TypeParamMaxLength: "500"},
				Description: "Keywords",
			},
			{
				Name:       "difficulty_level",
				DataType:   entity.FieldTypeInt64,
				Description: "Difficulty level",
			},
			{
				Name:       "is_success",
				DataType:   entity.FieldTypeInt64,
				Description: "Is success",
			},
		},
	}

	err = milvusCl.CreateCollection(ctx, schema, DefaultShardsNum)
	if err != nil {
		logger.Error("Create case collection failed",
			zap.String("collection", caseCollectionName),
			logger.Error(err),
		)
		return fmt.Errorf("创建案例集合失败: %w", err)
	}

	idx, err := entity.NewIndexIvfFlat(entity.L2, DefaultNlist)
	if err != nil {
		return fmt.Errorf("create case index entity failed: %w", err)
	}

	err = milvusCl.CreateIndex(ctx, caseCollectionName, "vector", idx, false)
	if err != nil {
		return fmt.Errorf("创建案例索引失败: %w", err)
	}

	logger.Info("Case collection created successfully",
		zap.String("collection", caseCollectionName),
		zap.Int("dimension", caseDim),
	)

	time.Sleep(2 * time.Second)

	return loadCaseCollectionIfNeeded(milvusCl)
}

func loadCaseCollectionIfNeeded(milvusCl client.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	err := milvusCl.LoadCollection(ctx, caseCollectionName, false)
	if err != nil {
		logger.Warn("Load case collection failed",
			zap.String("collection", caseCollectionName),
			logger.Error(err),
		)
		return nil
	}

	logger.Info("Case collection loaded", zap.String("collection", caseCollectionName))
	return nil
}

func InsertCaseVectors(ids []int64, vectors [][]float32, metadata []map[string]interface{}) error {
	milvusCl := GetMilvusClient()
	if milvusCl == nil {
		return fmt.Errorf("milvus client not initialized")
	}

	if len(ids) != len(vectors) || len(ids) != len(metadata) {
		return fmt.Errorf("ids, vectors and metadata length mismatch")
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	d := caseDim
	if d <= 0 {
		d = GetVectorDimension()
	}

	idColumn := entity.NewColumnInt64("id", ids)
	vectorColumn := entity.NewColumnFloatVector("vector", d, vectors)

	caseIDs := make([]int64, len(metadata))
	titles := make([]string, len(metadata))
	descriptions := make([]string, len(metadata))
	disputeTypes := make([]string, len(metadata))
	mediationTactics := make([]string, len(metadata))
	keyPoints := make([]string, len(metadata))
	keywords := make([]string, len(metadata))
	difficultyLevels := make([]int64, len(metadata))
	isSuccesses := make([]int64, len(metadata))

	for i, meta := range metadata {
		if v, ok := meta["case_id"].(int64); ok {
			caseIDs[i] = v
		}
		if v, ok := meta["caseId"].(int64); ok {
			caseIDs[i] = v
		}
		if v, ok := meta["title"].(string); ok {
			titles[i] = v
		}
		if v, ok := meta["description"].(string); ok {
			descriptions[i] = v
		}
		if v, ok := meta["dispute_type"].(string); ok {
			disputeTypes[i] = v
		}
		if v, ok := meta["disputeType"].(string); ok {
			disputeTypes[i] = v
		}
		if v, ok := meta["mediation_tactics"].(string); ok {
			mediationTactics[i] = v
		}
		if v, ok := meta["mediationTactics"].(string); ok {
			mediationTactics[i] = v
		}
		if v, ok := meta["key_points"].(string); ok {
			keyPoints[i] = v
		}
		if v, ok := meta["keyPoints"].(string); ok {
			keyPoints[i] = v
		}
		if v, ok := meta["keywords"].(string); ok {
			keywords[i] = v
		}
		if v, ok := meta["difficulty_level"].(int); ok {
			difficultyLevels[i] = int64(v)
		}
		if v, ok := meta["difficultyLevel"].(int); ok {
			difficultyLevels[i] = int64(v)
		}
		if v, ok := meta["difficulty_level"].(int64); ok {
			difficultyLevels[i] = v
		}
		if v, ok := meta["is_success"].(int); ok {
			isSuccesses[i] = int64(v)
		}
		if v, ok := meta["isSuccess"].(int); ok {
			isSuccesses[i] = int64(v)
		}
		if v, ok := meta["is_success"].(int64); ok {
			isSuccesses[i] = v
		}
	}

	columns := []entity.Column{
		idColumn,
		vectorColumn,
		entity.NewColumnInt64("case_id", caseIDs),
		entity.NewColumnVarChar("title", titles),
		entity.NewColumnVarChar("description", descriptions),
		entity.NewColumnVarChar("dispute_type", disputeTypes),
		entity.NewColumnVarChar("mediation_tactics", mediationTactics),
		entity.NewColumnVarChar("key_points", keyPoints),
		entity.NewColumnVarChar("keywords", keywords),
		entity.NewColumnInt64("difficulty_level", difficultyLevels),
		entity.NewColumnInt64("is_success", isSuccesses),
	}

	resultIDs, err := milvusCl.Insert(ctx, caseCollectionName, "", columns...)
	if err != nil {
		logger.Error("Insert case vectors failed",
			zap.Int("count", len(ids)),
			logger.Error(err),
		)
		return fmt.Errorf("插入案例向量失败: %w", err)
	}

	logger.Info("Case vectors inserted",
		zap.Int("count", len(ids)),
		zap.Any("resultIDs", resultIDs),
	)

	return nil
}

func SearchCaseVectors(queryVector []float32, topK int, filter string) ([]*CaseSearchResult, error) {
	milvusCl := GetMilvusClient()
	if milvusCl == nil {
		return nil, fmt.Errorf("milvus client not initialized")
	}

	if len(queryVector) == 0 {
		return nil, fmt.Errorf("查询向量不能为空")
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	if topK <= 0 {
		topK = 10
	}

	outputFields := []string{"id", "case_id", "title", "description", "dispute_type",
		"mediation_tactics", "key_points", "keywords", "difficulty_level", "is_success"}

	vectors := [][]float32{queryVector}

	sp, _ := entity.NewIndexIvfFlatSearchParam(DefaultNprobe)

	results, err := milvusCl.Search(
		ctx,
		caseCollectionName,
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
		logger.Error("Search case vectors failed",
			zap.Int("topK", topK),
			logger.Error(err),
		)
		return nil, fmt.Errorf("搜索案例向量失败: %w", err)
	}

	if len(results) == 0 {
		return []*CaseSearchResult{}, nil
	}

	result := results[0]
	searchResults := make([]*CaseSearchResult, result.ResultCount)

	idData := result.IDs
	distances := result.Scores

	fieldData := make(map[string]entity.Column)
	for _, field := range outputFields {
		for _, col := range result.Fields {
			if col.Name() == field {
				fieldData[field] = col
			}
		}
	}

	for i := 0; i < int(result.ResultCount); i++ {
		sr := &CaseSearchResult{
			Score:    convertScore(distances[i]),
			Distance: distances[i],
		}

		if idCol, ok := idData.(*entity.ColumnInt64); ok {
			if idCol.Len() > i {
				sr.ID = idCol.Data()[i]
			}
		}

		if col, ok := fieldData["case_id"].(*entity.ColumnInt64); ok && col.Len() > i {
			sr.CaseID = col.Data()[i]
		}
		if col, ok := fieldData["title"].(*entity.ColumnVarChar); ok && col.Len() > i {
			sr.Title = col.Data()[i]
		}
		if col, ok := fieldData["description"].(*entity.ColumnVarChar); ok && col.Len() > i {
			sr.Description = col.Data()[i]
		}
		if col, ok := fieldData["dispute_type"].(*entity.ColumnVarChar); ok && col.Len() > i {
			sr.DisputeType = col.Data()[i]
		}
		if col, ok := fieldData["mediation_tactics"].(*entity.ColumnVarChar); ok && col.Len() > i {
			sr.MediationTactics = col.Data()[i]
		}
		if col, ok := fieldData["key_points"].(*entity.ColumnVarChar); ok && col.Len() > i {
			sr.KeyPoints = col.Data()[i]
		}
		if col, ok := fieldData["keywords"].(*entity.ColumnVarChar); ok && col.Len() > i {
			sr.Keywords = col.Data()[i]
		}
		if col, ok := fieldData["difficulty_level"].(*entity.ColumnInt64); ok && col.Len() > i {
			sr.DifficultyLevel = int(col.Data()[i])
		}
		if col, ok := fieldData["is_success"].(*entity.ColumnInt64); ok && col.Len() > i {
			sr.IsSuccess = int32(col.Data()[i])
		}

		searchResults[i] = sr
	}

	logger.Debug("Case vector search completed",
		zap.Int("resultCount", len(searchResults)),
	)

	return searchResults, nil
}

func DeleteCaseVectors(ids []int64) error {
	milvusCl := GetMilvusClient()
	if milvusCl == nil {
		return fmt.Errorf("milvus client not initialized")
	}

	if len(ids) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	expr := fmt.Sprintf("id in [%s]", joinInt64List(ids))

	err := milvusCl.Delete(ctx, caseCollectionName, "", expr)
	if err != nil {
		logger.Error("Delete case vectors failed",
			zap.Int("count", len(ids)),
			logger.Error(err),
		)
		return fmt.Errorf("删除案例向量失败: %w", err)
	}

	logger.Info("Case vectors deleted", zap.Int("count", len(ids)))
	return nil
}

func DeleteCaseByCaseID(caseID int64) error {
	milvusCl := GetMilvusClient()
	if milvusCl == nil {
		return fmt.Errorf("milvus client not initialized")
	}

	expr := fmt.Sprintf("case_id == %d", caseID)

	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout*time.Second)
	defer cancel()

	err := milvusCl.Delete(ctx, caseCollectionName, "", expr)
	if err != nil {
		logger.Error("Delete case by case_id failed",
			zap.Int64("caseID", caseID),
			logger.Error(err),
		)
		return fmt.Errorf("按案例ID删除向量失败: %w", err)
	}

	logger.Info("Case vectors deleted by case_id", zap.Int64("caseID", caseID))
	return nil
}
