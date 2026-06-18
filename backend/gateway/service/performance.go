package service

import (
	"context"
)

type PerformanceService interface {
	GetPerformanceScoreList(ctx context.Context, orgID int64, page, pageSize int, period string, year int, month int, quarter int) ([]map[string]interface{}, int64, error)
	GetMyPerformance(ctx context.Context, userID int64, period string, year int, month int, quarter int) (map[string]interface{}, error)
	GetPerformanceDetail(ctx context.Context, userID int64, period string, year int, month int, quarter int) (map[string]interface{}, error)
	CalculatePerformanceScore(ctx context.Context, userID int64, period string, year int, month int, quarter int) (map[string]interface{}, error)
	GetPerformanceTrend(ctx context.Context, userID int64, orgID int64, period string, startDate, endDate string) ([]map[string]interface{}, error)
	GetPerformanceRanking(ctx context.Context, orgID int64, period string, year int, month int, quarter int, topN int) ([]map[string]interface{}, error)
	GetPerformanceIndicatorConfig(ctx context.Context) (map[string]interface{}, error)
}
