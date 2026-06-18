package service

import (
	"context"
)

type StatsService interface {
	GetDashboardStats(ctx context.Context, orgID int64, userID int64, role int32) (map[string]interface{}, error)
	GetHeatmapData(ctx context.Context, orgID int64, startTime, endTime string) ([]map[string]interface{}, error)
	GetOrganizationStats(ctx context.Context, orgID int64, startTime, endTime string) ([]map[string]interface{}, error)
	GetYearlyComparison(ctx context.Context, orgID int64, year int) (map[string]interface{}, error)
	GetMediatorRanking(ctx context.Context, orgID int64, startTime, endTime string, topN int) ([]map[string]interface{}, error)
	RefreshStatsCache(ctx context.Context, orgID int64) error
}
