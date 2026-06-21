package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/common/vector"
	"github.com/dispute-resolve/gateway/service"

	"go.uber.org/zap"
)

type CaseLibraryServiceImpl struct{}

func NewCaseLibraryService() service.CaseLibraryService {
	return &CaseLibraryServiceImpl{}
}

func (s *CaseLibraryServiceImpl) CreateCase(ctx context.Context, caseLib *model.CaseLibrary) error {
	if caseLib.CaseNo == "" {
		caseLib.CaseNo = utils.GenerateConsultNo()
	}

	err := database.GetDB().Create(caseLib).Error
	if err != nil {
		logger.Error("Create case library failed", logger.Error(err))
		return fmt.Errorf("创建案例失败: %w", err)
	}

	go func() {
		if e := s.VectorizeCase(context.Background(), caseLib.ID); e != nil {
			logger.Warn("Auto vectorize case failed",
				zap.Int64("caseId", caseLib.ID),
				logger.Error(e),
			)
		}
	}()

	return nil
}

func (s *CaseLibraryServiceImpl) UpdateCase(ctx context.Context, caseLib *model.CaseLibrary) error {
	err := database.GetDB().Model(caseLib).Omit("created_at", "vector_id", "vector_status", "reference_count", "avg_score", "score_count", "total_score").Updates(caseLib).Error
	if err != nil {
		logger.Error("Update case library failed", logger.Error(err))
		return fmt.Errorf("更新案例失败: %w", err)
	}

	needRevectorize := caseLib.Title != "" || caseLib.Description != "" ||
		caseLib.MediationTactics != "" || caseLib.KeyPoints != "" || caseLib.Keywords != ""

	if needRevectorize {
		go func() {
			if e := s.VectorizeCase(context.Background(), caseLib.ID); e != nil {
				logger.Warn("Re-vectorize case failed",
					zap.Int64("caseId", caseLib.ID),
					logger.Error(e),
				)
			}
		}()
	}

	return nil
}

func (s *CaseLibraryServiceImpl) DeleteCase(ctx context.Context, id int64) error {
	var caseLib model.CaseLibrary
	if err := database.GetDB().Where("id = ?", id).First(&caseLib).Error; err != nil {
		return fmt.Errorf("案例不存在: %w", err)
	}

	if caseLib.VectorID != "" {
		go func() {
			if err := vector.DeleteCaseByCaseID(id); err != nil {
				logger.Warn("Delete case vectors from milvus failed",
					zap.Int64("caseId", id),
					logger.Error(err),
				)
			}
		}()
	}

	err := database.GetDB().Model(&model.CaseLibrary{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
	if err != nil {
		return fmt.Errorf("删除案例失败: %w", err)
	}

	return nil
}

func (s *CaseLibraryServiceImpl) GetCase(ctx context.Context, id int64) (*model.CaseLibrary, error) {
	var caseLib model.CaseLibrary
	err := database.GetDB().Where("id = ? AND deleted_at IS NULL", id).First(&caseLib).Error
	if err != nil {
		return nil, fmt.Errorf("案例不存在: %w", err)
	}
	return &caseLib, nil
}

func (s *CaseLibraryServiceImpl) ListCases(ctx context.Context, page, pageSize int, keyword, disputeType string, difficultyLevel, status int) ([]*model.CaseLibrary, int64, error) {
	var cases []*model.CaseLibrary
	var total int64

	db := database.GetDB().Model(&model.CaseLibrary{}).Where("deleted_at IS NULL")

	if status >= 0 {
		db = db.Where("status = ?", status)
	} else {
		db = db.Where("status = ?", constants.CaseLibraryStatusActive)
	}

	if keyword != "" {
		db = db.Where("title LIKE ? OR description LIKE ? OR keywords LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if disputeType != "" {
		db = db.Where("dispute_type = ?", disputeType)
	}
	if difficultyLevel > 0 {
		db = db.Where("difficulty_level = ?", difficultyLevel)
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Offset(offset).Limit(pageSize).Order("avg_score DESC, reference_count DESC, created_at DESC").Find(&cases)

	return cases, total, nil
}

func (s *CaseLibraryServiceImpl) SearchSimilarCases(ctx context.Context, query string, topK int) ([]*vector.CaseSearchResult, error) {
	if topK <= 0 {
		topK = 5
	}

	embedding, err := vector.GetEmbedding(query)
	if err != nil {
		logger.Error("Get embedding for case search failed", logger.Error(err))
		return nil, fmt.Errorf("获取查询向量失败: %w", err)
	}

	filter := ""
	results, err := vector.SearchCaseVectors(embedding, topK, filter)
	if err != nil {
		logger.Error("Search case vectors failed", logger.Error(err))
		return nil, fmt.Errorf("搜索相似案例失败: %w", err)
	}

	return results, nil
}

func (s *CaseLibraryServiceImpl) ScoreCase(ctx context.Context, score *model.CaseLibraryScore) error {
	if score.Score < 1 || score.Score > 5 {
		return fmt.Errorf("评分必须在1-5之间")
	}

	var caseLib model.CaseLibrary
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", score.CaseID).First(&caseLib).Error; err != nil {
		return fmt.Errorf("案例不存在")
	}

	score.CaseNo = caseLib.CaseNo

	var existing model.CaseLibraryScore
	err := database.GetDB().Where("case_id = ? AND user_id = ? AND source_case_id = ?",
		score.CaseID, score.UserID, score.SourceCaseID).First(&existing).Error

	if err == nil {
		oldScore := existing.Score
		err = database.GetDB().Model(&existing).Update("score", score.Score).Error
		if err != nil {
			return fmt.Errorf("更新评分失败: %w", err)
		}
		diff := score.Score - oldScore
		updateScoreStats(score.CaseID, float64(diff), 0)
	} else {
		err = database.GetDB().Create(score).Error
		if err != nil {
			return fmt.Errorf("创建评分失败: %w", err)
		}
		updateScoreStats(score.CaseID, float64(score.Score), 1)
	}

	return nil
}

func updateScoreStats(caseID int64, scoreDelta float64, countDelta int) {
	db := database.GetDB()
	var caseLib model.CaseLibrary
	if err := db.Where("id = ?", caseID).First(&caseLib).Error; err != nil {
		return
	}

	newTotal := caseLib.TotalScore + scoreDelta
	newCount := caseLib.ScoreCount + countDelta
	var newAvg float64
	if newCount > 0 {
		newAvg = newTotal / float64(newCount)
	}

	db.Model(&model.CaseLibrary{}).Where("id = ?", caseID).Updates(map[string]interface{}{
		"total_score": newTotal,
		"score_count": newCount,
		"avg_score":   newAvg,
	})
}

func (s *CaseLibraryServiceImpl) QuoteCase(ctx context.Context, quote *model.CaseLibraryQuote) error {
	var caseLib model.CaseLibrary
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL AND status = ?",
		quote.LibraryCaseID, constants.CaseLibraryStatusActive).First(&caseLib).Error; err != nil {
		return fmt.Errorf("案例不存在或已归档")
	}

	quote.LibraryCaseNo = caseLib.CaseNo

	if quote.QuoteType == constants.CaseLibraryQuoteTypeTactics && caseLib.MediationTactics != "" {
		if quote.QuoteContent == "" {
			quote.QuoteContent = caseLib.MediationTactics
		}
	} else if quote.QuoteType == constants.CaseLibraryQuoteTypeStrategy && caseLib.KeyPoints != "" {
		if quote.QuoteContent == "" {
			quote.QuoteContent = caseLib.KeyPoints
		}
	} else if quote.QuoteType == constants.CaseLibraryQuoteTypeFull {
		if quote.QuoteContent == "" {
			quote.QuoteContent = fmt.Sprintf("标题：%s\n描述：%s\n调解话术：%s\n调解要点：%s\n结果：%s",
				caseLib.Title, caseLib.Description, caseLib.MediationTactics, caseLib.KeyPoints, caseLib.ResultSummary)
		}
	}

	err := database.GetDB().Create(quote).Error
	if err != nil {
		return fmt.Errorf("创建引用记录失败: %w", err)
	}

	now := time.Now()
	database.GetDB().Model(&model.CaseLibrary{}).Where("id = ?", quote.LibraryCaseID).Updates(map[string]interface{}{
		"reference_count": caseLib.ReferenceCount + 1,
		"last_used_at":    now,
	})

	return nil
}

func (s *CaseLibraryServiceImpl) GetQuoteList(ctx context.Context, sourceCaseID int64) ([]*model.CaseLibraryQuote, error) {
	var quotes []*model.CaseLibraryQuote
	database.GetDB().Where("source_case_id = ?", sourceCaseID).
		Order("created_at DESC").
		Find(&quotes)
	return quotes, nil
}

func (s *CaseLibraryServiceImpl) ArchiveCase(ctx context.Context, id int64, archivedBy int64, reason int32) error {
	var caseLib model.CaseLibrary
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", id).First(&caseLib).Error; err != nil {
		return fmt.Errorf("案例不存在")
	}

	if caseLib.Status == constants.CaseLibraryStatusArchived {
		return fmt.Errorf("案例已归档")
	}

	caseData, _ := json.Marshal(caseLib)
	archive := &model.CaseLibraryArchive{
		OriginalID:     caseLib.ID,
		CaseNo:         caseLib.CaseNo,
		Title:          caseLib.Title,
		ArchiveReason:  reason,
		AvgScore:       caseLib.AvgScore,
		ReferenceCount: caseLib.ReferenceCount,
		LastUsedAt:     caseLib.LastUsedAt,
		ArchivedBy:     archivedBy,
		CaseData:       string(caseData),
	}

	if err := database.GetDB().Create(archive).Error; err != nil {
		return fmt.Errorf("创建归档记录失败: %w", err)
	}

	now := time.Now()
	if err := database.GetDB().Model(&caseLib).Updates(map[string]interface{}{
		"status":      constants.CaseLibraryStatusArchived,
		"archived_at": now,
	}).Error; err != nil {
		return fmt.Errorf("更新案例状态失败: %w", err)
	}

	go func() {
		if err := vector.DeleteCaseByCaseID(caseLib.ID); err != nil {
			logger.Warn("Delete archived case vectors failed",
				zap.Int64("caseId", caseLib.ID),
				logger.Error(err),
			)
		}
	}()

	return nil
}

func (s *CaseLibraryServiceImpl) ArchiveUnusedCases(ctx context.Context) (int, error) {
	db := database.GetDB()
	if db == nil {
		return 0, fmt.Errorf("数据库未初始化")
	}

	now := time.Now()
	threshold := now.AddDate(0, -constants.CaseLibraryArchiveMonths, 0)

	var unusedCases []model.CaseLibrary
	err := db.Where("status = ? AND deleted_at IS NULL AND (last_used_at IS NULL OR last_used_at < ?) AND created_at < ?",
		constants.CaseLibraryStatusActive, threshold, threshold).
		Find(&unusedCases).Error
	if err != nil {
		logger.Error("Query unused cases for archive failed", logger.Error(err))
		return 0, err
	}

	archivedCount := 0
	for _, c := range unusedCases {
		err := s.ArchiveCase(ctx, c.ID, 0, constants.CaseLibraryArchiveReasonUnused)
		if err != nil {
			logger.Warn("Archive unused case failed",
				zap.Int64("caseId", c.ID),
				zap.String("caseNo", c.CaseNo),
				logger.Error(err),
			)
			continue
		}
		archivedCount++
		logger.Info("Archived unused case",
			zap.Int64("caseId", c.ID),
			zap.String("caseNo", c.CaseNo),
		)
	}

	logger.Info("Archive unused cases completed",
		zap.Int("archivedCount", archivedCount),
		zap.Int("totalUnused", len(unusedCases)),
	)

	return archivedCount, nil
}

func (s *CaseLibraryServiceImpl) VectorizeCase(ctx context.Context, id int64) error {
	var caseLib model.CaseLibrary
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", id).First(&caseLib).Error; err != nil {
		return fmt.Errorf("案例不存在")
	}

	database.GetDB().Model(&caseLib).Update("vector_status", constants.CaseLibraryVectorProcessing)

	text := caseLib.Title
	if caseLib.Description != "" {
		text += "\n" + caseLib.Description
	}
	if caseLib.MediationTactics != "" {
		text += "\n" + caseLib.MediationTactics
	}
	if caseLib.KeyPoints != "" {
		text += "\n" + caseLib.KeyPoints
	}
	if caseLib.Keywords != "" {
		text += "\n" + caseLib.Keywords
	}

	embedding, err := vector.GetEmbedding(text)
	if err != nil {
		database.GetDB().Model(&caseLib).Update("vector_status", constants.CaseLibraryVectorFailed)
		return fmt.Errorf("获取向量失败: %w", err)
	}

	metadata := map[string]interface{}{
		"case_id":           caseLib.ID,
		"title":             caseLib.Title,
		"description":       caseLib.Description,
		"dispute_type":      caseLib.DisputeType,
		"mediation_tactics": caseLib.MediationTactics,
		"key_points":        caseLib.KeyPoints,
		"keywords":          caseLib.Keywords,
		"difficulty_level":  caseLib.DifficultyLevel,
		"is_success":        caseLib.IsSuccess,
	}

	if caseLib.VectorID != "" {
		vector.DeleteCaseVectors([]int64{caseLib.ID})
	}

	vectorID := utils.GenerateIDStr()
	err = vector.InsertCaseVectors([]int64{caseLib.ID}, [][]float32{embedding}, []map[string]interface{}{metadata})
	if err != nil {
		database.GetDB().Model(&caseLib).Update("vector_status", constants.CaseLibraryVectorFailed)
		return fmt.Errorf("插入向量失败: %w", err)
	}

	database.GetDB().Model(&caseLib).Updates(map[string]interface{}{
		"vector_id":     vectorID,
		"vector_status": constants.CaseLibraryVectorDone,
	})

	logger.Info("Case vectorized successfully",
		zap.Int64("caseId", caseLib.ID),
		zap.String("vectorId", vectorID),
	)

	return nil
}

func (s *CaseLibraryServiceImpl) VectorizeAllCases(ctx context.Context) (int, error) {
	var cases []model.CaseLibrary
	database.GetDB().Where("status = ? AND deleted_at IS NULL AND vector_status != ?",
		constants.CaseLibraryStatusActive, constants.CaseLibraryVectorDone).
		Find(&cases)

	processed := 0
	for _, c := range cases {
		if err := s.VectorizeCase(ctx, c.ID); err != nil {
			logger.Warn("Vectorize case failed",
				zap.Int64("caseId", c.ID),
				logger.Error(err),
			)
			continue
		}
		processed++
		time.Sleep(time.Millisecond * 100)
	}

	return processed, nil
}

func (s *CaseLibraryServiceImpl) RestoreFromArchive(ctx context.Context, id int64) error {
	var caseLib model.CaseLibrary
	if err := database.GetDB().Where("id = ? AND deleted_at IS NULL", id).First(&caseLib).Error; err != nil {
		return fmt.Errorf("案例不存在")
	}

	if caseLib.Status != constants.CaseLibraryStatusArchived {
		return fmt.Errorf("案例未归档，无需恢复")
	}

	if err := database.GetDB().Model(&caseLib).Updates(map[string]interface{}{
		"status":      constants.CaseLibraryStatusActive,
		"archived_at": nil,
	}).Error; err != nil {
		return fmt.Errorf("恢复案例失败: %w", err)
	}

	go func() {
		if err := s.VectorizeCase(context.Background(), id); err != nil {
			logger.Warn("Re-vectorize restored case failed",
				zap.Int64("caseId", id),
				logger.Error(err),
			)
		}
	}()

	return nil
}

func (s *CaseLibraryServiceImpl) GetArchiveList(ctx context.Context, page, pageSize int) ([]*model.CaseLibraryArchive, int64, error) {
	var archives []*model.CaseLibraryArchive
	var total int64

	db := database.GetDB().Model(&model.CaseLibraryArchive{})
	db.Count(&total)

	offset := (page - 1) * pageSize
	db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&archives)

	return archives, total, nil
}
