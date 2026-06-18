package impl

import (
	"context"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type StatsServiceImpl struct{}

func NewStatsService() service.StatsService {
	return &StatsServiceImpl{}
}

func (s *StatsServiceImpl) GetDashboardStats(ctx context.Context, orgID int64, userID int64, role int32) (map[string]interface{}, error) {
	cacheKey := constants.RedisPrefixStatsDashboard + utils.Int64ToString(orgID) + ":" + utils.Int64ToString(userID)
	var result map[string]interface{}

	if cache.Get(ctx, cacheKey, &result) == nil && result != nil {
		return result, nil
	}

	db := database.GetDB().Table("dispute_case").Where("deleted_at IS NULL")
	if role == constants.RoleMediator {
		db = db.Where("mediator_id = ?", userID)
	} else if role >= constants.RoleLeader && role <= constants.RoleDirector {
		db = db.Where("organization_id = ?", orgID)
	}

	var total int64
	db.Count(&total)

	var pending int64
	db.Where("status = ?", constants.CaseStatusPending).Count(&pending)

	var mediating int64
	db.Where("status = ?", constants.CaseStatusMediating).Count(&mediating)

	var approving int64
	db.Where("status = ?", constants.CaseStatusApproving).Count(&approving)

	var closed int64
	db.Where("status = ?", constants.CaseStatusClosed).Count(&closed)

	var success int64
	db.Where("status = ? AND mediation_result = ?", constants.CaseStatusClosed, 1).Count(&success)

	var avgDays float64
	database.GetDB().Table("dispute_case").
		Select("AVG(DATEDIFF(closed_at, created_at))").
		Where("status = ? AND deleted_at IS NULL", constants.CaseStatusClosed).
		Scan(&avgDays)

	successRate := 0.0
	if closed > 0 {
		successRate = float64(success) / float64(closed) * 100
	}

	var todayNew int64
	today := time.Now().Format("2006-01-02")
	db.Where("DATE(created_at) = ?", today).Count(&todayNew)

	var todayClosed int64
	db.Where("DATE(closed_at) = ?", today).Count(&todayClosed)

	result = map[string]interface{}{
		"total":        total,
		"pending":      pending,
		"mediating":    mediating,
		"approving":    approving,
		"closed":       closed,
		"successRate":  successRate,
		"avgDays":      avgDays,
		"todayNew":     todayNew,
		"todayClosed":  todayClosed,
		"updatedAt":    time.Now().Format("2006-01-02 15:04:05"),
	}

	cache.Set(ctx, cacheKey, result, 5*time.Minute)
	return result, nil
}

func (s *StatsServiceImpl) GetHeatmapData(ctx context.Context, orgID int64, startTime, endTime string) ([]map[string]interface{}, error) {
	cacheKey := constants.RedisPrefixStatsHeatmap + utils.Int64ToString(orgID)
	var result []map[string]interface{}

	if cache.Get(ctx, cacheKey, &result) == nil && len(result) > 0 {
		return result, nil
	}

	db := database.GetDB().Table("dispute_case").
		Select("longitude, latitude, COUNT(*) as count, organization_id").
		Where("longitude IS NOT NULL AND latitude IS NOT NULL AND deleted_at IS NULL")

	if orgID > 0 {
		db = db.Where("organization_id = ?", orgID)
	}
	if startTime != "" {
		db = db.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		db = db.Where("created_at <= ?", endTime)
	}

	db.Group("longitude, latitude, organization_id").
		Order("count DESC").
		Limit(1000).
		Find(&result)

	cache.Set(ctx, cacheKey, result, 30*time.Minute)
	return result, nil
}

func (s *StatsServiceImpl) GetOrganizationStats(ctx context.Context, orgID int64, startTime, endTime string) ([]map[string]interface{}, error) {
	var result []map[string]interface{}

	db := database.GetDB().Table("dispute_case dc").
		Select("o.id, o.name, o.level, COUNT(dc.id) as total, " +
			"SUM(CASE WHEN dc.status = 1 THEN 1 ELSE 0 END) as pending, " +
			"SUM(CASE WHEN dc.status = 3 THEN 1 ELSE 0 END) as mediating, " +
			"SUM(CASE WHEN dc.status = 5 THEN 1 ELSE 0 END) as closed, " +
			"SUM(CASE WHEN dc.status = 5 AND dc.mediation_result = 1 THEN 1 ELSE 0 END) as success").
		Joins("LEFT JOIN organization o ON dc.organization_id = o.id").
		Where("dc.deleted_at IS NULL")

	if orgID > 0 {
		db = db.Where("o.parent_id = ? OR o.id = ?", orgID, orgID)
	}
	if startTime != "" {
		db = db.Where("dc.created_at >= ?", startTime)
	}
	if endTime != "" {
		db = db.Where("dc.created_at <= ?", endTime)
	}

	db.Group("o.id, o.name, o.level").
		Order("total DESC").
		Find(&result)

	for _, r := range result {
		total, _ := r["total"].(int64)
		closed, _ := r["closed"].(int64)
		success, _ := r["success"].(int64)
		if closed > 0 {
			r["successRate"] = float64(success) / float64(closed) * 100
		} else {
			r["successRate"] = 0
		}
		r["closeRate"] = 0.0
		if total > 0 {
			r["closeRate"] = float64(closed) / float64(total) * 100
		}
	}

	return result, nil
}

func (s *StatsServiceImpl) GetYearlyComparison(ctx context.Context, orgID int64, year int) (map[string]interface{}, error) {
	currentYear := year
	if currentYear == 0 {
		currentYear = time.Now().Year()
	}
	lastYear := currentYear - 1

	type MonthStats struct {
		Month   int   `json:"month"`
		Current int64 `json:"current"`
		Last    int64 `json:"last"`
	}

	stats := make([]MonthStats, 12)
	for i := 0; i < 12; i++ {
		stats[i].Month = i + 1
	}

	db := database.GetDB().Table("dispute_case").Where("deleted_at IS NULL")
	if orgID > 0 {
		db = db.Where("organization_id = ?", orgID)
	}

	var currentData []struct {
		Month int   `json:"month"`
		Count int64 `json:"count"`
	}
	db.Where("YEAR(created_at) = ?", currentYear).
		Select("MONTH(created_at) as month, COUNT(*) as count").
		Group("MONTH(created_at)").
		Find(&currentData)

	for _, d := range currentData {
		if d.Month >= 1 && d.Month <= 12 {
			stats[d.Month-1].Current = d.Count
		}
	}

	var lastData []struct {
		Month int   `json:"month"`
		Count int64 `json:"count"`
	}
	db.Where("YEAR(created_at) = ?", lastYear).
		Select("MONTH(created_at) as month, COUNT(*) as count").
		Group("MONTH(created_at)").
		Find(&lastData)

	for _, d := range lastData {
		if d.Month >= 1 && d.Month <= 12 {
			stats[d.Month-1].Last = d.Count
		}
	}

	totalCurrent := int64(0)
	totalLast := int64(0)
	for _, s := range stats {
		totalCurrent += s.Current
		totalLast += s.Last
	}

	growthRate := 0.0
	if totalLast > 0 {
		growthRate = float64(totalCurrent-totalLast) / float64(totalLast) * 100
	}

	return map[string]interface{}{
		"year":       currentYear,
		"lastYear":   lastYear,
		"totalCurrent": totalCurrent,
		"totalLast":    totalLast,
		"growthRate":   growthRate,
		"monthlyData":  stats,
	}, nil
}

func (s *StatsServiceImpl) GetMediatorRanking(ctx context.Context, orgID int64, startTime, endTime string, topN int) ([]map[string]interface{}, error) {
	if topN <= 0 || topN > 100 {
		topN = 10
	}

	db := database.GetDB().Table("dispute_case dc").
		Select("u.id, u.real_name, u.avatar, o.name as org_name, " +
			"COUNT(dc.id) as total_cases, " +
			"SUM(CASE WHEN dc.status = 5 THEN 1 ELSE 0 END) as closed_cases, " +
			"SUM(CASE WHEN dc.status = 5 AND dc.mediation_result = 1 THEN 1 ELSE 0 END) as success_cases, " +
			"AVG(CASE WHEN dc.status = 5 THEN DATEDIFF(dc.closed_at, dc.created_at) ELSE NULL END) as avg_days, " +
			"AVG(CASE WHEN dc.satisfaction_score > 0 THEN dc.satisfaction_score ELSE NULL END) as avg_satisfaction").
		Joins("LEFT JOIN user u ON dc.mediator_id = u.id").
		Joins("LEFT JOIN organization o ON u.organization_id = o.id").
		Where("dc.mediator_id IS NOT NULL AND dc.deleted_at IS NULL")

	if orgID > 0 {
		db = db.Where("u.organization_id = ?", orgID)
	}
	if startTime != "" {
		db = db.Where("dc.created_at >= ?", startTime)
	}
	if endTime != "" {
		db = db.Where("dc.created_at <= ?", endTime)
	}

	var result []map[string]interface{}
	db.Group("u.id, u.real_name, u.avatar, o.name").
		Order("total_cases DESC").
		Limit(topN).
		Find(&result)

	for i, r := range result {
		closed, _ := r["closed_cases"].(int64)
		success, _ := r["success_cases"].(int64)
		total, _ := r["total_cases"].(int64)

		successRate := 0.0
		if closed > 0 {
			successRate = float64(success) / float64(closed) * 100
		}

		closeRate := 0.0
		if total > 0 {
			closeRate = float64(closed) / float64(total) * 100
		}

		r["rank"] = i + 1
		r["successRate"] = successRate
		r["closeRate"] = closeRate
	}

	return result, nil
}

func (s *StatsServiceImpl) RefreshStatsCache(ctx context.Context, orgID int64) error {
	cache.DelByPrefix(ctx, constants.RedisPrefixStats)
	return nil
}
