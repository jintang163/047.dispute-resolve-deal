package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/ai"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type satisfactionServiceImpl struct {
	db *gorm.DB
}

var satisfactionServiceInstance *satisfactionServiceImpl

func InitSatisfactionService() {
	satisfactionServiceInstance = &satisfactionServiceImpl{
		db: database.GetDB(),
	}
	logger.Info("Satisfaction service initialized")
}

func GetSatisfactionService() *satisfactionServiceImpl {
	if satisfactionServiceInstance == nil {
		InitSatisfactionService()
	}
	return satisfactionServiceInstance
}

func (s *satisfactionServiceImpl) AnalyzeSatisfaction(ctx context.Context, caseID int64) (*model.ImprovementOrder, error) {
	var caseData model.DisputeCase
	if err := s.db.Where("id = ? AND deleted_at IS NULL", caseID).First(&caseData).Error; err != nil {
		return nil, fmt.Errorf("case not found: %w", err)
	}

	if caseData.SatisfactionRemark == "" {
		return nil, fmt.Errorf("no satisfaction comment to analyze")
	}

	if caseData.SentimentAnalyzedAt != nil {
		return nil, fmt.Errorf("satisfaction already analyzed")
	}

	caseInfo := map[string]interface{}{
		"title":            caseData.Title,
		"mediatorName":     caseData.MediatorName,
		"mediationResult":  caseData.MediationResult,
		"satisfactionScore": caseData.SatisfactionScore,
	}

	analyzer := ai.GetSentimentAnalyzer()
	result, err := analyzer.AnalyzeSatisfactionComment(caseData.SatisfactionRemark, caseInfo)
	if err != nil {
		return nil, fmt.Errorf("sentiment analysis failed: %w", err)
	}

	now := time.Now()
	keywordsJSON, _ := json.Marshal(map[string][]string{
		"positive": result.PositiveKeywords,
		"negative": result.NegativeKeywords,
	})

	if err := s.db.Model(&caseData).Updates(map[string]interface{}{
		"sentiment_emotion":     result.Emotion,
		"sentiment_score":       result.SentimentScore,
		"sentiment_confidence":  result.Confidence,
		"sentiment_keywords":    string(keywordsJSON),
		"sentiment_summary":     result.Summary,
		"sentiment_analyzed_at": now,
	}).Error; err != nil {
		return nil, fmt.Errorf("update case sentiment failed: %w", err)
	}

	logger.Info("Satisfaction sentiment analyzed",
		zap.Int64("caseId", caseID),
		zap.String("emotion", result.Emotion),
		zap.Float64("sentimentScore", result.SentimentScore),
	)

	var improvementOrder *model.ImprovementOrder
	if result.Emotion == "negative" {
		improvementOrder, err = s.createImprovementOrder(&caseData, result)
		if err != nil {
			logger.Error("Failed to create improvement order",
				zap.Int64("caseId", caseID),
				logger.Error(err),
			)
		} else {
			s.notifyMediator(&caseData, improvementOrder)
		}
	}

	return improvementOrder, nil
}

func (s *satisfactionServiceImpl) createImprovementOrder(caseData *model.DisputeCase, sentiment *ai.SatisfactionSentimentResult) (*model.ImprovementOrder, error) {
	now := time.Now()
	deadline := now.AddDate(0, 0, 7)
	if sentiment.PrioritySuggestion == 1 {
		deadline = now.AddDate(0, 0, 3)
	} else if sentiment.PrioritySuggestion == 3 {
		deadline = now.AddDate(0, 0, 14)
	}

	orderNo := fmt.Sprintf("IM%s%04d", now.Format("20060102150405"), caseData.ID%10000)

	deductionScore := s.calculateDeductionScore(sentiment)

	order := &model.ImprovementOrder{
		ID:                    utils.GenerateID(),
		OrderNo:              orderNo,
		CaseID:               caseData.ID,
		CaseNo:               caseData.CaseNo,
		CaseTitle:            caseData.Title,
		ApplicantID:          caseData.ApplicantID,
		ApplicantName:        caseData.ApplicantName,
		MediatorID:           caseData.MediatorID,
		MediatorName:         caseData.MediatorName,
		OrgID:                caseData.OrganizationID,
		OrgName:              caseData.OrganizationName,
		SatisfactionScore:    caseData.SatisfactionScore,
		SatisfactionComment:  caseData.SatisfactionRemark,
		SentimentEmotion:     sentiment.Emotion,
		SentimentScore:       sentiment.SentimentScore,
		SentimentSummary:     sentiment.Summary,
		IssueType:            sentiment.IssueType,
		IssueDescription:     sentiment.IssueDescription,
		ImprovementSuggestion: sentiment.ImprovementSuggestion,
		Status:               model.ImprovementStatusPending,
		Priority:             sentiment.PrioritySuggestion,
		Deadline:             &deadline,
		AssignedAt:           &now,
		DeductionScore:       deductionScore,
		DeductionReason:      fmt.Sprintf("满意度评价负面(%s): %s", sentiment.IssueType, sentiment.Summary),
	}

	if err := s.db.Create(order).Error; err != nil {
		return nil, fmt.Errorf("create improvement order failed: %w", err)
	}

	logger.Info("Improvement order created",
		zap.String("orderNo", orderNo),
		zap.Int64("caseId", caseData.ID),
		zap.Int64("mediatorId", caseData.MediatorID),
		zap.String("issueType", sentiment.IssueType),
		zap.Float64("deductionScore", deductionScore),
	)

	return order, nil
}

func (s *satisfactionServiceImpl) calculateDeductionScore(sentiment *ai.SatisfactionSentimentResult) float64 {
	baseScore := 2.0

	switch sentiment.PrioritySuggestion {
	case 1:
		baseScore = 5.0
	case 2:
		baseScore = 3.0
	case 3:
		baseScore = 1.0
	}

	confidenceMultiplier := 0.5 + sentiment.Confidence*0.5

	return baseScore * confidenceMultiplier
}

func (s *satisfactionServiceImpl) notifyMediator(caseData *model.DisputeCase, order *model.ImprovementOrder) {
	priorityText := "中"
	switch order.Priority {
	case 1:
		priorityText = "高"
	case 3:
		priorityText = "低"
	}

	deadlineStr := ""
	if order.Deadline != nil {
		deadlineStr = order.Deadline.Format("2006-01-02")
	}

	title := fmt.Sprintf("【整改通知】满意度负面评价整改工单(%s)", order.OrderNo)
	content := fmt.Sprintf(
		"尊敬的%s调解员：\n\n"+
			"您处理的案件（编号：%s，标题：%s）收到群众负面满意度评价。\n\n"+
			"问题类型：%s\n"+
			"问题描述：%s\n"+
			"改进建议：%s\n"+
			"优先级：%s\n"+
			"整改截止时间：%s\n\n"+
			"请务必在截止时间前完成整改并提交整改报告，逾期未整改将纳入绩效考核扣分。\n\n"+
			"工单编号：%s",
		caseData.MediatorName,
		caseData.CaseNo,
		caseData.Title,
		s.getIssueTypeLabel(order.IssueType),
		order.IssueDescription,
		order.ImprovementSuggestion,
		priorityText,
		deadlineStr,
		order.OrderNo,
	)

	params := map[string]interface{}{
		"orderNo":      order.OrderNo,
		"caseNo":       caseData.CaseNo,
		"issueType":    order.IssueType,
		"priority":     order.Priority,
		"deadline":     deadlineStr,
		"deductionScore": order.DeductionScore,
	}
	paramsJSON, _ := json.Marshal(params)

	channels := []string{"站内信", "短信"}
	for _, ch := range channels {
		insertSQL := `INSERT INTO notification_record 
			(id, msg_no, template_code, channel_type, receiver_id, receiver_type, title, content, params, biz_type, biz_id, case_id, case_no, send_status, is_read, created_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		msgNo := fmt.Sprintf("NTF%d", time.Now().UnixNano())
		if err := s.db.Exec(insertSQL,
			utils.GenerateID(),
			msgNo,
			"improvement_order",
			ch,
			caseData.MediatorID,
			2,
			title,
			content,
			string(paramsJSON),
			"improvement",
			order.OrderNo,
			caseData.ID,
			caseData.CaseNo,
			1,
			0,
			now(),
		).Error; err != nil {
			logger.Warn("Send improvement notification failed",
				zap.String("channel", ch),
				zap.Int64("mediatorId", caseData.MediatorID),
				logger.Error(err),
			)
		}
	}

	logger.Info("Improvement order notification sent",
		zap.String("orderNo", order.OrderNo),
		zap.Int64("mediatorId", caseData.MediatorID),
	)
}

func (s *satisfactionServiceImpl) getIssueTypeLabel(issueType string) string {
	labels := map[string]string{
		"attitude":     "态度问题",
		"efficiency":   "效率问题",
		"professional": "专业性问题",
		"result":       "结果不满意",
		"process":      "流程问题",
		"other":        "其他",
	}
	if label, ok := labels[issueType]; ok {
		return label
	}
	return issueType
}

func (s *satisfactionServiceImpl) SubmitRectification(ctx context.Context, orderID int64, content string, result string) error {
	var order model.ImprovementOrder
	if err := s.db.Where("id = ? AND deleted_at IS NULL", orderID).First(&order).Error; err != nil {
		return fmt.Errorf("improvement order not found: %w", err)
	}

	if order.Status != model.ImprovementStatusPending && order.Status != model.ImprovementStatusProcessing {
		return fmt.Errorf("improvement order is not in rectifiable status")
	}

	now := time.Now()
	order.Status = model.ImprovementStatusRectified
	order.RectifyContent = content
	order.RectifyResult = result
	order.RectifiedAt = &now

	if err := s.db.Save(&order).Error; err != nil {
		return fmt.Errorf("submit rectification failed: %w", err)
	}

	logger.Info("Rectification submitted",
		zap.String("orderNo", order.OrderNo),
		zap.Int64("orderID", orderID),
	)

	return nil
}

func (s *satisfactionServiceImpl) ReviewRectification(ctx context.Context, orderID int64, reviewerID int64, reviewerName string, opinion string, approved bool) error {
	var order model.ImprovementOrder
	if err := s.db.Where("id = ? AND deleted_at IS NULL", orderID).First(&order).Error; err != nil {
		return fmt.Errorf("improvement order not found: %w", err)
	}

	if order.Status != model.ImprovementStatusRectified {
		return fmt.Errorf("improvement order is not in reviewable status")
	}

	now := time.Now()
	order.ReviewedBy = reviewerID
	order.ReviewedByName = reviewerName
	order.ReviewOpinion = opinion
	order.ReviewedAt = &now

	if approved {
		order.Status = model.ImprovementStatusReviewed
		s.applyDeduction(&order)
	} else {
		order.Status = model.ImprovementStatusProcessing
		newDeadline := now.AddDate(0, 0, 3)
		order.Deadline = &newDeadline
	}

	if err := s.db.Save(&order).Error; err != nil {
		return fmt.Errorf("review rectification failed: %w", err)
	}

	logger.Info("Rectification reviewed",
		zap.String("orderNo", order.OrderNo),
		zap.Bool("approved", approved),
		zap.Float64("deductionScore", order.DeductionScore),
	)

	return nil
}

func (s *satisfactionServiceImpl) applyDeduction(order *model.ImprovementOrder) {
	now := time.Now()
	order.IsDeductionApplied = 1
	order.DeductionAppliedAt = &now

	go func() {
		s.recalculatePerformanceScore(order.MediatorID)
	}()

	logger.Info("Deduction applied",
		zap.String("orderNo", order.OrderNo),
		zap.Int64("mediatorId", order.MediatorID),
		zap.Float64("deductionScore", order.DeductionScore),
	)
}

func (s *satisfactionServiceImpl) recalculatePerformanceScore(mediatorID int64) {
	now := time.Now()
	year := now.Year()
	month := int(now.Month())

	var stat model.PerformanceStat
	err := s.db.Where("user_id = ? AND period_type = 2 AND YEAR(CAST(period_date AS DATE)) = ? AND MONTH(CAST(period_date AS DATE)) = ?",
		mediatorID, year, month).First(&stat).Error
	if err != nil {
		logger.Warn("No performance stat found for deduction update",
			zap.Int64("mediatorId", mediatorID),
		)
		return
	}

	var totalDeduction float64
	var deductionCount int
	s.db.Model(&model.ImprovementOrder{}).
		Where("mediator_id = ? AND is_deduction_applied = 1 AND deleted_at IS NULL", mediatorID).
		Select("SUM(deduction_score), COUNT(*)").
		Row().Scan(&totalDeduction, &deductionCount)

	stat.DeductionTotal = totalDeduction
	stat.DeductionCount = deductionCount
	stat.FinalScore = stat.Score - totalDeduction
	if stat.FinalScore < 0 {
		stat.FinalScore = 0
	}

	if stat.FinalScore >= 90 {
		stat.Grade = "S"
	} else if stat.FinalScore >= 80 {
		stat.Grade = "A"
	} else if stat.FinalScore >= 70 {
		stat.Grade = "B"
	} else if stat.FinalScore >= 60 {
		stat.Grade = "C"
	} else {
		stat.Grade = "D"
	}

	s.db.Save(&stat)

	logger.Info("Performance score recalculated with deduction",
		zap.Int64("mediatorId", mediatorID),
		zap.Float64("originalScore", stat.Score),
		zap.Float64("deductionTotal", totalDeduction),
		zap.Float64("finalScore", stat.FinalScore),
		zap.String("grade", stat.Grade),
	)
}

func (s *satisfactionServiceImpl) GetImprovementOrderList(ctx context.Context, mediatorID int64, status int, page int, pageSize int) ([]model.ImprovementOrder, int64, error) {
	query := s.db.Model(&model.ImprovementOrder{}).Where("deleted_at IS NULL")

	if mediatorID > 0 {
		query = query.Where("mediator_id = ?", mediatorID)
	}
	if status > 0 {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var orders []model.ImprovementOrder
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&orders).Error; err != nil {
		return nil, 0, fmt.Errorf("query improvement orders failed: %w", err)
	}

	return orders, total, nil
}

func (s *satisfactionServiceImpl) GetImprovementOrderDetail(ctx context.Context, orderID int64) (*model.ImprovementOrder, error) {
	var order model.ImprovementOrder
	if err := s.db.Where("id = ? AND deleted_at IS NULL", orderID).First(&order).Error; err != nil {
		return nil, fmt.Errorf("improvement order not found: %w", err)
	}
	return &order, nil
}

func (s *satisfactionServiceImpl) CloseImprovementOrder(ctx context.Context, orderID int64, remark string) error {
	var order model.ImprovementOrder
	if err := s.db.Where("id = ? AND deleted_at IS NULL", orderID).First(&order).Error; err != nil {
		return fmt.Errorf("improvement order not found: %w", err)
	}

	if order.Status != model.ImprovementStatusReviewed {
		return fmt.Errorf("only reviewed orders can be closed")
	}

	order.Status = model.ImprovementStatusClosed
	order.Remark = remark

	if err := s.db.Save(&order).Error; err != nil {
		return fmt.Errorf("close improvement order failed: %w", err)
	}

	logger.Info("Improvement order closed",
		zap.String("orderNo", order.OrderNo),
	)

	return nil
}

func (s *satisfactionServiceImpl) GetSatisfactionSentimentStats(ctx context.Context, orgID int64, startDate, endDate string) (map[string]interface{}, error) {
	query := s.db.Model(&model.DisputeCase{}).
		Where("satisfaction_remark != '' AND satisfaction_remark IS NOT NULL AND sentiment_analyzed_at IS NOT NULL AND deleted_at IS NULL")

	if orgID > 0 {
		query = query.Where("org_id = ?", orgID)
	}
	if startDate != "" {
		query = query.Where("sentiment_analyzed_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("sentiment_analyzed_at <= ?", endDate)
	}

	var totalCount int64
	var positiveCount int64
	var neutralCount int64
	var negativeCount int64
	var avgScore float64

	query.Count(&totalCount)

	s.db.Model(&model.DisputeCase{}).
		Where("satisfaction_remark != '' AND sentiment_analyzed_at IS NOT NULL AND sentiment_emotion = 'positive' AND deleted_at IS NULL").
		Count(&positiveCount)

	s.db.Model(&model.DisputeCase{}).
		Where("satisfaction_remark != '' AND sentiment_analyzed_at IS NOT NULL AND sentiment_emotion = 'neutral' AND deleted_at IS NULL").
		Count(&neutralCount)

	s.db.Model(&model.DisputeCase{}).
		Where("satisfaction_remark != '' AND sentiment_analyzed_at IS NOT NULL AND sentiment_emotion = 'negative' AND deleted_at IS NULL").
		Count(&negativeCount)

	s.db.Model(&model.DisputeCase{}).
		Where("satisfaction_remark != '' AND sentiment_analyzed_at IS NOT NULL AND deleted_at IS NULL").
		Select("AVG(sentiment_score)").Scan(&avgScore)

	var issueTypeStats []map[string]interface{}
	s.db.Model(&model.ImprovementOrder{}).
		Where("deleted_at IS NULL").
		Select("issue_type, COUNT(*) as count").
		Group("issue_type").
		Find(&issueTypeStats)

	return map[string]interface{}{
		"totalAnalyzed":   totalCount,
		"positiveCount":   positiveCount,
		"neutralCount":    neutralCount,
		"negativeCount":   negativeCount,
		"positiveRate":    calcRate(positiveCount, totalCount),
		"neutralRate":     calcRate(neutralCount, totalCount),
		"negativeRate":    calcRate(negativeCount, totalCount),
		"avgSentimentScore": avgScore,
		"issueTypeStats":  issueTypeStats,
	}, nil
}

func calcRate(count, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total) * 100
}

func now() time.Time {
	return time.Now()
}

func ProcessPendingSatisfactionAnalysis() {
	svc := GetSatisfactionService()
	db := svc.db
	if db == nil {
		return
	}

	var cases []model.DisputeCase
	if err := db.Where(
		"satisfaction_remark != '' AND satisfaction_remark IS NOT NULL AND sentiment_analyzed_at IS NULL AND status = ? AND deleted_at IS NULL",
		constants.CaseStatusClosed,
	).Limit(50).Find(&cases).Error; err != nil {
		logger.Error("Query pending satisfaction analysis cases failed", logger.Error(err))
		return
	}

	if len(cases) == 0 {
		return
	}

	logger.Info("Found pending satisfaction analysis cases", zap.Int("count", len(cases)))

	for _, c := range cases {
		ctx := context.Background()
		if _, err := svc.AnalyzeSatisfaction(ctx, c.ID); err != nil {
			logger.Warn("Analyze satisfaction failed",
				zap.Int64("caseId", c.ID),
				logger.Error(err),
			)
		}
	}
}
