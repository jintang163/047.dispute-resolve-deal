package impl

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/gateway/service"
)

type PerformanceServiceImpl struct{}

func NewPerformanceService() service.PerformanceService {
	return &PerformanceServiceImpl{}
}

type PerformanceIndicator struct {
	Code     string  `json:"code"`
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	MaxScore float64 `json:"maxScore"`
}

var defaultIndicators = []PerformanceIndicator{
	{Code: "case_count", Name: "案件数量", Weight: 0.2, MaxScore: 100},
	{Code: "close_rate", Name: "办结率", Weight: 0.25, MaxScore: 100},
	{Code: "success_rate", Name: "调解成功率", Weight: 0.25, MaxScore: 100},
	{Code: "avg_days", Name: "平均办理天数", Weight: 0.15, MaxScore: 100},
	{Code: "satisfaction", Name: "满意度", Weight: 0.15, MaxScore: 100},
}

func (s *PerformanceServiceImpl) GetPerformanceScoreList(ctx context.Context, orgID int64, page, pageSize int, period string, year int, month int, quarter int) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("user u").
		Select("u.id, u.username, u.real_name, u.role, o.org_name").
		Joins("LEFT JOIN organization o ON u.org_id = o.id").
		Where("u.role IN ? AND u.deleted_at IS NULL", []int{constants.RoleMediator, constants.RoleLeader, constants.RoleDirector})

	if orgID > 0 {
		db = db.Where("u.org_id = ?", orgID)
	}

	var total int64
	db.Count(&total)

	var users []map[string]interface{}
	offset := (page - 1) * pageSize
	db.Order("u.id ASC").Offset(offset).Limit(pageSize).Find(&users)

	result := make([]map[string]interface{}, 0, len(users))
	for _, user := range users {
		userID := user["id"].(int64)
		score, err := s.calculatePerformanceInternal(ctx, userID, period, year, month, quarter)
		if err != nil {
			logger.Error("Calculate performance error", logger.Error(err))
			continue
		}
		score["userInfo"] = user
		result = append(result, score)
	}

	return result, total, nil
}

func (s *PerformanceServiceImpl) GetMyPerformance(ctx context.Context, userID int64, period string, year int, month int, quarter int) (map[string]interface{}, error) {
	return s.calculatePerformanceInternal(ctx, userID, period, year, month, quarter)
}

func (s *PerformanceServiceImpl) GetPerformanceDetail(ctx context.Context, userID int64, period string, year int, month int, quarter int) (map[string]interface{}, error) {
	result, err := s.calculatePerformanceInternal(ctx, userID, period, year, month, quarter)
	if err != nil {
		return nil, err
	}

	caseList, err := s.getMediatorCases(ctx, userID, period, year, month, quarter)
	if err != nil {
		logger.Error("Get mediator cases error", logger.Error(err))
	} else {
		result["caseList"] = caseList
	}

	return result, nil
}

func (s *PerformanceServiceImpl) CalculatePerformanceScore(ctx context.Context, userID int64, period string, year int, month int, quarter int) (map[string]interface{}, error) {
	result, err := s.calculatePerformanceInternal(ctx, userID, period, year, month, quarter)
	if err != nil {
		return nil, err
	}

	record := &model.PerformanceScore{
		UserID:       userID,
		Period:       getPeriodCode(period),
		Year:         year,
		Month:        month,
		Quarter:      quarter,
		TotalScore:   result["totalScore"].(float64),
		CaseCount:    result["caseCount"].(int),
		CloseRate:    result["closeRate"].(float64),
		SuccessRate:  result["successRate"].(float64),
		AvgDays:      result["avgDays"].(float64),
		Satisfaction: result["satisfaction"].(float64),
		Grade:        result["grade"].(string),
		CalculatedAt: time.Now(),
	}

	if err := database.GetDB().Create(record).Error; err != nil {
		logger.Error("Save performance score error", logger.Error(err))
	}

	return result, nil
}

func (s *PerformanceServiceImpl) GetPerformanceTrend(ctx context.Context, userID int64, orgID int64, period string, startDate, endDate string) ([]map[string]interface{}, error) {
	result := make([]map[string]interface{}, 0)

	start, _ := time.Parse("2006-01-02", startDate)
	end, _ := time.Parse("2006-01-02", endDate)

	for d := start; !d.After(end); {
		year := d.Year()
		month := int(d.Month())

		var userIDList []int64
		if userID > 0 {
			userIDList = []int64{userID}
		} else {
			db := database.GetDB().Table("user").Select("id").Where("role = ? AND deleted_at IS NULL", constants.RoleMediator)
			if orgID > 0 {
				db = db.Where("org_id = ?", orgID)
			}
			db.Pluck("id", &userIDList)
		}

		monthData := map[string]interface{}{
			"year":  year,
			"month": month,
			"date":  fmt.Sprintf("%d-%02d", year, month),
		}

		totalCaseCount := 0
		totalCloseRate := 0.0
		totalSuccessRate := 0.0
		totalAvgDays := 0.0
		totalSatisfaction := 0.0
		count := 0

		for _, uid := range userIDList {
			perf, _ := s.calculatePerformanceInternal(ctx, uid, "month", year, month, 0)
			if perf != nil {
				totalCaseCount += perf["caseCount"].(int)
				totalCloseRate += perf["closeRate"].(float64)
				totalSuccessRate += perf["successRate"].(float64)
				totalAvgDays += perf["avgDays"].(float64)
				totalSatisfaction += perf["satisfaction"].(float64)
				count++
			}
		}

		if count > 0 {
			monthData["caseCount"] = totalCaseCount
			monthData["closeRate"] = roundFloat(totalCloseRate/float64(count), 2)
			monthData["successRate"] = roundFloat(totalSuccessRate/float64(count), 2)
			monthData["avgDays"] = roundFloat(totalAvgDays/float64(count), 2)
			monthData["satisfaction"] = roundFloat(totalSatisfaction/float64(count), 2)
		} else {
			monthData["caseCount"] = 0
			monthData["closeRate"] = 0.0
			monthData["successRate"] = 0.0
			monthData["avgDays"] = 0.0
			monthData["satisfaction"] = 0.0
		}

		result = append(result, monthData)

		d = d.AddDate(0, 1, 0)
	}

	return result, nil
}

func (s *PerformanceServiceImpl) GetPerformanceRanking(ctx context.Context, orgID int64, period string, year int, month int, quarter int, topN int) ([]map[string]interface{}, error) {
	db := database.GetDB().Table("user u").
		Select("u.id, u.username, u.real_name").
		Where("u.role = ? AND u.deleted_at IS NULL", constants.RoleMediator)

	if orgID > 0 {
		db = db.Where("u.org_id = ?", orgID)
	}

	var users []map[string]interface{}
	db.Find(&users)

	type rankItem struct {
		UserID     int64
		RealName   string
		TotalScore float64
	}

	rankList := make([]rankItem, 0, len(users))
	for _, user := range users {
		userID := user["id"].(int64)
		realName, _ := user["real_name"].(string)
		perf, err := s.calculatePerformanceInternal(ctx, userID, period, year, month, quarter)
		if err != nil {
			continue
		}
		rankList = append(rankList, rankItem{
			UserID:     userID,
			RealName:   realName,
			TotalScore: perf["totalScore"].(float64),
		})
	}

	for i := 0; i < len(rankList)-1; i++ {
		for j := i + 1; j < len(rankList); j++ {
			if rankList[j].TotalScore > rankList[i].TotalScore {
				rankList[i], rankList[j] = rankList[j], rankList[i]
			}
		}
	}

	if topN > 0 && topN < len(rankList) {
		rankList = rankList[:topN]
	}

	result := make([]map[string]interface{}, 0, len(rankList))
	for i, item := range rankList {
		result = append(result, map[string]interface{}{
			"rank":       i + 1,
			"userId":     item.UserID,
			"realName":   item.RealName,
			"totalScore": item.TotalScore,
		})
	}

	return result, nil
}

func (s *PerformanceServiceImpl) GetPerformanceIndicatorConfig(ctx context.Context) (map[string]interface{}, error) {
	var configs []*model.PerformanceIndicatorConfig
	database.GetDB().Where("status = 1").Order("sort_order ASC").Find(&configs)

	indicators := make([]map[string]interface{}, 0, len(configs))
	totalWeight := 0.0
	for _, config := range configs {
		indicators = append(indicators, map[string]interface{}{
			"id":          config.ID,
			"code":        config.IndicatorCode,
			"name":        config.IndicatorName,
			"weight":      config.Weight,
			"description": config.Description,
			"sortOrder":   config.SortOrder,
		})
		totalWeight += config.Weight
	}

	if len(indicators) == 0 {
		for _, ind := range defaultIndicators {
			indicators = append(indicators, map[string]interface{}{
				"code":   ind.Code,
				"name":   ind.Name,
				"weight": ind.Weight,
			})
			totalWeight += ind.Weight
		}
	}

	return map[string]interface{}{
		"indicators":  indicators,
		"totalWeight": totalWeight,
	}, nil
}

func (s *PerformanceServiceImpl) calculatePerformanceInternal(ctx context.Context, userID int64, period string, year int, month int, quarter int) (map[string]interface{}, error) {
	caseQuery := s.buildCaseQuery(userID, period, year, month, quarter)

	var totalCases int64
	caseQuery.Count(&totalCases)

	var closedCases int64
	caseQuery.Where("dc.status = ?", constants.CaseStatusClosed).Count(&closedCases)

	var successfulCases int64
	caseQuery.Where("dc.status = ? AND dc.mediation_result = ?", constants.CaseStatusClosed, constants.MediationResultSuccess).Count(&successfulCases)

	closeRate := 0.0
	if totalCases > 0 {
		closeRate = roundFloat(float64(closedCases)/float64(totalCases)*100, 2)
	}

	successRate := 0.0
	if closedCases > 0 {
		successRate = roundFloat(float64(successfulCases)/float64(closedCases)*100, 2)
	}

	avgDays := s.calculateAvgDays(caseQuery)

	satisfaction := s.calculateSatisfaction(userID, period, year, month, quarter)

	indicatorScores := make(map[string]float64)
	indicatorScores["case_count"] = s.calculateCaseCountScore(int(totalCases))
	indicatorScores["close_rate"] = s.calculateCloseRateScore(closeRate)
	indicatorScores["success_rate"] = s.calculateSuccessRateScore(successRate)
	indicatorScores["avg_days"] = s.calculateAvgDaysScore(avgDays)
	indicatorScores["satisfaction"] = s.calculateSatisfactionScore(satisfaction)

	totalScore := 0.0
	for _, ind := range defaultIndicators {
		totalScore += indicatorScores[ind.Code] * ind.Weight
	}
	totalScore = roundFloat(totalScore, 2)

	grade := s.getGrade(totalScore)

	return map[string]interface{}{
		"userId":          userID,
		"period":          period,
		"year":            year,
		"month":           month,
		"quarter":         quarter,
		"caseCount":       int(totalCases),
		"closedCount":     int(closedCases),
		"successfulCount": int(successfulCases),
		"closeRate":       closeRate,
		"successRate":     successRate,
		"avgDays":         avgDays,
		"satisfaction":    satisfaction,
		"indicatorScores": indicatorScores,
		"totalScore":      totalScore,
		"grade":           grade,
	}, nil
}

func (s *PerformanceServiceImpl) buildCaseQuery(userID int64, period string, year int, month int, quarter int) *database.DB {
	db := database.GetDB().Table("dispute_case dc").
		Where("dc.mediator_id = ? AND dc.deleted_at IS NULL", userID)

	if year > 0 {
		db = db.Where("YEAR(dc.created_at) = ?", year)
	}
	if month > 0 {
		db = db.Where("MONTH(dc.created_at) = ?", month)
	}
	if quarter > 0 {
		startMonth := (quarter - 1) * 3
		endMonth := startMonth + 3
		db = db.Where("MONTH(dc.created_at) >= ? AND MONTH(dc.created_at) < ?", startMonth+1, endMonth+1)
	}

	return db
}

func (s *PerformanceServiceImpl) calculateAvgDays(db *database.DB) float64 {
	type Result struct {
		CreatedAt time.Time
		ClosedAt  *time.Time
	}

	var results []Result
	db.Select("created_at, closed_at").Where("status = ?", constants.CaseStatusClosed).Scan(&results)

	if len(results) == 0 {
		return 0
	}

	totalDays := 0.0
	count := 0
	for _, r := range results {
		if r.ClosedAt != nil {
			days := r.ClosedAt.Sub(r.CreatedAt).Hours() / 24
			totalDays += days
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return roundFloat(totalDays/float64(count), 2)
}

func (s *PerformanceServiceImpl) calculateSatisfaction(userID int64, period string, year int, month int, quarter int) float64 {
	db := database.GetDB().Table("dispute_evaluation de").
		Joins("LEFT JOIN dispute_case dc ON de.case_id = dc.id").
		Where("dc.mediator_id = ? AND dc.deleted_at IS NULL", userID)

	if year > 0 {
		db = db.Where("YEAR(de.created_at) = ?", year)
	}
	if month > 0 {
		db = db.Where("MONTH(de.created_at) = ?", month)
	}

	var avg float64
	db.Select("AVG(de.score)").Row().Scan(&avg)

	return roundFloat(avg, 2)
}

func (s *PerformanceServiceImpl) calculateCaseCountScore(count int) float64 {
	switch {
	case count >= 30:
		return 100
	case count >= 20:
		return 90
	case count >= 15:
		return 80
	case count >= 10:
		return 70
	case count >= 5:
		return 60
	default:
		return float64(count * 10)
	}
}

func (s *PerformanceServiceImpl) calculateCloseRateScore(rate float64) float64 {
	switch {
	case rate >= 95:
		return 100
	case rate >= 90:
		return 90
	case rate >= 85:
		return 80
	case rate >= 80:
		return 70
	case rate >= 70:
		return 60
	default:
		return rate * 0.6
	}
}

func (s *PerformanceServiceImpl) calculateSuccessRateScore(rate float64) float64 {
	switch {
	case rate >= 90:
		return 100
	case rate >= 85:
		return 90
	case rate >= 80:
		return 80
	case rate >= 75:
		return 70
	case rate >= 70:
		return 60
	default:
		return rate * 0.6
	}
}

func (s *PerformanceServiceImpl) calculateAvgDaysScore(days float64) float64 {
	switch {
	case days <= 7:
		return 100
	case days <= 14:
		return 90
	case days <= 21:
		return 80
	case days <= 30:
		return 70
	case days <= 45:
		return 60
	default:
		return math.Max(0, 100-days)
	}
}

func (s *PerformanceServiceImpl) calculateSatisfactionScore(score float64) float64 {
	switch {
	case score >= 4.8:
		return 100
	case score >= 4.5:
		return 90
	case score >= 4.2:
		return 80
	case score >= 4.0:
		return 70
	case score >= 3.5:
		return 60
	default:
		return score * 15
	}
}

func (s *PerformanceServiceImpl) getGrade(score float64) string {
	switch {
	case score >= 90:
		return "S"
	case score >= 80:
		return "A"
	case score >= 70:
		return "B"
	case score >= 60:
		return "C"
	default:
		return "D"
	}
}

func (s *PerformanceServiceImpl) getMediatorCases(ctx context.Context, userID int64, period string, year int, month int, quarter int) ([]map[string]interface{}, error) {
	caseQuery := s.buildCaseQuery(userID, period, year, month, quarter)

	var cases []map[string]interface{}
	caseQuery.Select("dc.id, dc.case_no, dc.title, dc.status, dc.mediation_result, dc.created_at, dc.closed_at").
		Order("dc.created_at DESC").
		Find(&cases)

	for _, c := range cases {
		if status, ok := c["status"].(int); ok {
			c["statusText"] = constants.CaseStatusMap[status]
		}
		if result, ok := c["mediation_result"].(int); ok {
			resultMap := map[int]string{
				0: "待处理", 1: "调解成功", 2: "调解失败", 3: "部分达成",
			}
			c["mediationResultText"] = resultMap[result]
		}
	}

	return cases, nil
}

func getPeriodCode(period string) int {
	switch period {
	case "month":
		return 1
	case "quarter":
		return 2
	case "year":
		return 3
	default:
		return 1
	}
}

func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
