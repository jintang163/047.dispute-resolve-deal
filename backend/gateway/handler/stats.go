package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type HeatmapQueryRequest struct {
	StartTime      string `form:"startTime"`
	EndTime        string `form:"endTime"`
	TypeID         int64  `form:"typeId"`
	OrganizationID int64  `form:"organizationId"`
}

type DashboardStatsRequest struct {
	Period         string `form:"period"`
	OrganizationID int64  `form:"organizationId"`
}

func GetDashboardStats(ctx context.Context, c *app.RequestContext) {
	var req DashboardStatsRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := req.OrganizationID
	if orgID == 0 {
		orgID = userInfo.OrganizationID
	}

	if req.Period == "" {
		req.Period = "month"
	}

	cacheKey := fmt.Sprintf("dashboard:stats:%d:%s", orgID, req.Period)
	cachedData, err := cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data map[string]interface{}
		json.Unmarshal([]byte(cachedData), &data)
		c.JSON(http.StatusOK, response.Success(data))
		return
	}

	now := time.Now()
	var startDate time.Time

	switch req.Period {
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "quarter":
		startDate = now.AddDate(0, -3, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0)
	}

	db := database.GetDB().Table("dispute_case").
		Where("deleted_at IS NULL").
		Where("created_at >= ?", startDate)

	if orgID > 0 {
		var childOrgs []int64
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", orgID, orgID).
			Pluck("id", &childOrgs)
		db = db.Where("organization_id IN ?", childOrgs)
	}

	var totalCases int64
	db.Count(&totalCases)

	var pendingCases int64
	db.Where("status = ?", constants.CaseStatusPending).Count(&pendingCases)

	var mediatingCases int64
	db.Where("status = ?", constants.CaseStatusMediating).Count(&mediatingCases)

	var closedCases int64
	db.Where("status = ?", constants.CaseStatusClosed).Count(&closedCases)

	var successCount int64
	db.Where("status = ? AND mediation_result = ?", constants.CaseStatusClosed, constants.MediationResultSuccess).
		Count(&successCount)

	successRate := 0.0
	if closedCases > 0 {
		successRate = float64(successCount) / float64(closedCases) * 100
	}

	var avgDays float64
	database.GetDB().Table("dispute_case").
		Where("status = ? AND deleted_at IS NULL", constants.CaseStatusClosed).
		Where("created_at >= ?", startDate).
		Select("AVG(TIMESTAMPDIFF(DAY, created_at, closed_time))").
		Scan(&avgDays)

	var todayNew int64
	todayStart := time.Now().Format("2006-01-02")
	database.GetDB().Table("dispute_case").
		Where("DATE(created_at) = ?", todayStart).
		Where("deleted_at IS NULL").
		Count(&todayNew)

	var typeStats []map[string]interface{}
	database.GetDB().Table("dispute_case dc").
		Select("dt.type_name, COUNT(*) as count").
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Where("dc.deleted_at IS NULL").
		Where("dc.created_at >= ?", startDate).
		Group("dt.id, dt.type_name").
		Order("count DESC").
		Limit(10).
		Find(&typeStats)

	var trendData []map[string]interface{}
	dateFormat := "%Y-%m-%d"
	if req.Period == "year" {
		dateFormat = "%Y-%m"
	}
	database.GetDB().Table("dispute_case").
		Select(fmt.Sprintf("DATE_FORMAT(created_at, '%s') as date, COUNT(*) as count", dateFormat)).
		Where("deleted_at IS NULL").
		Where("created_at >= ?", startDate).
		Group("date").
		Order("date ASC").
		Find(&trendData)

	var sourceStats []map[string]interface{}
	sourceMap := map[int]string{
		constants.CaseSourceKiosk:   "自助终端",
		constants.CaseSourceMiniApp: "小程序",
		constants.CaseSourcePhone:   "电话",
		constants.CaseSourceWindow:  "窗口",
		constants.CaseSourceTransfer: "转送",
	}
	database.GetDB().Table("dispute_case").
		Select("case_source, COUNT(*) as count").
		Where("deleted_at IS NULL").
		Where("created_at >= ?", startDate).
		Group("case_source").
		Find(&sourceStats)

	for _, item := range sourceStats {
		if source, ok := item["case_source"].(int); ok {
			item["source_name"] = sourceMap[source]
		}
	}

	result := map[string]interface{}{
		"totalCases":     totalCases,
		"pendingCases":   pendingCases,
		"mediatingCases": mediatingCases,
		"closedCases":    closedCases,
		"successRate":    fmt.Sprintf("%.1f%%", successRate),
		"avgDays":        fmt.Sprintf("%.1f", avgDays),
		"todayNew":       todayNew,
		"typeStats":      typeStats,
		"trendData":      trendData,
		"sourceStats":    sourceStats,
	}

	jsonData, _ := json.Marshal(result)
	cache.Set(ctx, cacheKey, string(jsonData), 300)

	c.JSON(http.StatusOK, response.Success(result))
}

func GetHeatmapData(ctx context.Context, c *app.RequestContext) {
	var req HeatmapQueryRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := req.OrganizationID
	if orgID == 0 {
		orgID = userInfo.OrganizationID
	}

	cacheKey := fmt.Sprintf("heatmap:%d:%s:%s:%d", orgID, req.StartTime, req.EndTime, req.TypeID)
	cachedData, err := cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data []map[string]interface{}
		json.Unmarshal([]byte(cachedData), &data)
		c.JSON(http.StatusOK, response.Success(data))
		return
	}

	db := database.GetDB().Table("dispute_case dc").
		Select("dc.latitude, dc.longitude, dc.id, dc.case_no, dc.title, dt.type_name, so.org_name, dc.status, dc.created_at").
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
		Where("dc.deleted_at IS NULL").
		Where("dc.latitude IS NOT NULL AND dc.longitude IS NOT NULL").
		Where("dc.latitude != 0 AND dc.longitude != 0")

	if orgID > 0 {
		var childOrgs []int64
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", orgID, orgID).
			Pluck("id", &childOrgs)
		db = db.Where("dc.organization_id IN ?", childOrgs)
	}

	if req.StartTime != "" {
		db = db.Where("dc.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("dc.created_at <= ?", req.EndTime)
	}
	if req.TypeID > 0 {
		db = db.Where("dc.type_id = ?", req.TypeID)
	}

	var data []map[string]interface{}
	db.Order("dc.created_at DESC").Find(&data)

	for _, item := range data {
		if status, ok := item["status"].(int32); ok {
			item["status_name"] = constants.CaseStatusMap[int(status)]
		}
	}

	jsonData, _ := json.Marshal(data)
	cache.Set(ctx, cacheKey, string(jsonData), 600)

	c.JSON(http.StatusOK, response.Success(data))
}

func GetOrganizationStats(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	cacheKey := fmt.Sprintf("org:stats:%d", userInfo.OrganizationID)
	cachedData, err := cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data []map[string]interface{}
		json.Unmarshal([]byte(cachedData), &data)
		c.JSON(http.StatusOK, response.Success(data))
		return
	}

	var childOrgs []int64
	database.GetDB().Table("sys_organization").
		Select("id").
		Where("parent_id = ? OR id = ?", userInfo.OrganizationID, userInfo.OrganizationID).
		Pluck("id", &childOrgs)

	var stats []map[string]interface{}
	database.GetDB().Table("sys_organization so").
		Select("so.id, so.org_name, so.org_code, "+
			"COUNT(DISTINCT dc.id) as total_cases, "+
			"SUM(CASE WHEN dc.status = 10 THEN 1 ELSE 0 END) as pending_cases, "+
			"SUM(CASE WHEN dc.status = 20 THEN 1 ELSE 0 END) as mediating_cases, "+
			"SUM(CASE WHEN dc.status = 50 THEN 1 ELSE 0 END) as closed_cases, "+
			"SUM(CASE WHEN dc.status = 50 AND dc.mediation_result = 1 THEN 1 ELSE 0 END) as success_cases").
		Joins("LEFT JOIN dispute_case dc ON so.id = dc.organization_id AND dc.deleted_at IS NULL").
		Where("so.id IN ?", childOrgs).
		Group("so.id, so.org_name, so.org_code").
		Find(&stats)

	for _, item := range stats {
		closed := item["closed_cases"].(int64)
		success := item["success_cases"].(int64)
		rate := "0%"
		if closed > 0 {
			rate = fmt.Sprintf("%.1f%%", float64(success)/float64(closed)*100)
		}
		item["success_rate"] = rate
	}

	jsonData, _ := json.Marshal(stats)
	cache.Set(ctx, cacheKey, string(jsonData), 300)

	c.JSON(http.StatusOK, response.Success(stats))
}

func GetYearlyComparison(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	year := c.Query("year")
	if year == "" {
		year = strconv.Itoa(time.Now().Year())
	}

	prevYear, _ := strconv.Atoi(year)
	prevYear--

	var currentYearData []map[string]interface{}
	var prevYearData []map[string]interface{}

	for i, y := range []string{year, strconv.Itoa(prevYear)} {
		var data []map[string]interface{}
		err := database.GetDB().Table("dispute_case").
			Select("DATE_FORMAT(created_at, '%m') as month, COUNT(*) as count").
			Where("YEAR(created_at) = ?", y).
			Where("deleted_at IS NULL").
			Where("organization_id = ?", userInfo.OrganizationID).
			Group("month").
			Order("month ASC").
			Find(&data).Error

		if err != nil {
			logger.Error("Get yearly comparison failed", logger.Error(err))
			continue
		}

		if i == 0 {
			currentYearData = data
		} else {
			prevYearData = data
		}
	}

	result := map[string]interface{}{
		"currentYear": year,
		"prevYear":    prevYear,
		"currentData": currentYearData,
		"prevData":    prevYearData,
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func GetMediatorRanking(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	limit := c.DefaultQuery("limit", "10")
	limitNum, _ := strconv.Atoi(limit)

	var rankings []map[string]interface{}
	database.GetDB().Table("sys_user su").
		Select("su.id, su.real_name, so.org_name, "+
			"COUNT(DISTINCT dc.id) as total_cases, "+
			"SUM(CASE WHEN dc.status = 50 THEN 1 ELSE 0 END) as closed_cases, "+
			"SUM(CASE WHEN dc.status = 50 AND dc.mediation_result = 1 THEN 1 ELSE 0 END) as success_cases, "+
			"AVG(CASE WHEN dc.status = 50 THEN dc.satisfaction_score ELSE NULL END) as avg_satisfaction").
		Joins("LEFT JOIN dispute_case dc ON su.id = dc.mediator_id AND dc.deleted_at IS NULL").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("su.role = ?", constants.RoleMediator).
		Where("su.status = 1").
		Where("su.organization_id = ? OR so.parent_id = ?", userInfo.OrganizationID, userInfo.OrganizationID).
		Group("su.id, su.real_name, so.org_name").
		Having("total_cases > 0").
		Order("success_cases DESC, total_cases DESC").
		Limit(limitNum).
		Find(&rankings)

	for i, item := range rankings {
		item["rank"] = i + 1
		closed := item["closed_cases"].(int64)
		success := item["success_cases"].(int64)
		rate := "0%"
		if closed > 0 {
			rate = fmt.Sprintf("%.1f%%", float64(success)/float64(closed)*100)
		}
		item["success_rate"] = rate
	}

	c.JSON(http.StatusOK, response.Success(rankings))
}

func GetHeatmapTimeline(ctx context.Context, c *app.RequestContext) {
	var req HeatmapQueryRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := req.OrganizationID
	if orgID == 0 {
		orgID = userInfo.OrganizationID
	}

	if req.EndTime == "" {
		req.EndTime = time.Now().Format("2006-01-02 15:04:05")
	}
	if req.StartTime == "" {
		t, _ := time.Parse("2006-01-02 15:04:05", req.EndTime)
		req.StartTime = t.AddDate(0, 0, -7).Format("2006-01-02 15:04:05")
	}

	days := 7
	endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
	if err != nil {
		endTime = time.Now()
	}
	startTime, _ := time.Parse("2006-01-02 15:04:05", req.StartTime)
	daysDiff := int(endTime.Sub(startTime).Hours()/24) + 1
	if daysDiff > 0 && daysDiff <= 90 {
		days = daysDiff
	}

	timeline := make([]map[string]interface{}, 0, days)

	for i := 0; i < days; i++ {
		dayStart := startTime.AddDate(0, 0, i).Format("2006-01-02") + " 00:00:00"
		dayEnd := startTime.AddDate(0, 0, i).Format("2006-01-02") + " 23:59:59"
		dayLabel := startTime.AddDate(0, 0, i).Format("2006-01-02")

		db := database.GetDB().Table("dispute_case dc").
			Select("dc.latitude, dc.longitude, dc.id, dc.case_no, dc.title, dt.type_name, so.org_name, dc.status, dc.created_at").
			Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
			Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
			Where("dc.deleted_at IS NULL").
			Where("dc.latitude IS NOT NULL AND dc.longitude IS NOT NULL").
			Where("dc.latitude != 0 AND dc.longitude != 0").
			Where("dc.created_at >= ? AND dc.created_at <= ?", dayStart, dayEnd)

		if orgID > 0 {
			var childOrgs []int64
			database.GetDB().Table("sys_organization").
				Select("id").
				Where("parent_id = ? OR id = ?", orgID, orgID).
				Pluck("id", &childOrgs)
			db = db.Where("dc.organization_id IN ?", childOrgs)
		}
		if req.TypeID > 0 {
			db = db.Where("dc.type_id = ?", req.TypeID)
		}

		var dayData []map[string]interface{}
		db.Order("dc.created_at ASC").Find(&dayData)

		for _, item := range dayData {
			if status, ok := item["status"].(int32); ok {
				item["status_name"] = constants.CaseStatusMap[int(status)]
			}
		}

		timeline = append(timeline, map[string]interface{}{
			"date":  dayLabel,
			"count": len(dayData),
			"items": dayData,
		})
	}

	c.JSON(http.StatusOK, response.Success(timeline))
}

func GetTopCommunities(ctx context.Context, c *app.RequestContext) {
	var req HeatmapQueryRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := req.OrganizationID
	if orgID == 0 {
		orgID = userInfo.OrganizationID
	}

	limitStr := c.DefaultQuery("limit", "5")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 20 {
		limit = 5
	}

	db := database.GetDB().Table("dispute_case dc").
		Select("so.id as org_id, so.org_name, so.longitude, so.latitude, COUNT(*) as case_count").
		Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
		Where("dc.deleted_at IS NULL").
		Where("so.longitude IS NOT NULL AND so.latitude IS NOT NULL")

	if orgID > 0 {
		var childOrgs []int64
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", orgID, orgID).
			Pluck("id", &childOrgs)
		db = db.Where("dc.organization_id IN ?", childOrgs)
	}
	if req.StartTime != "" {
		db = db.Where("dc.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("dc.created_at <= ?", req.EndTime)
	}
	if req.TypeID > 0 {
		db = db.Where("dc.type_id = ?", req.TypeID)
	}

	var topList []map[string]interface{}
	db.Group("so.id, so.org_name, so.longitude, so.latitude").
		Order("case_count DESC").
		Limit(limit).
		Find(&topList)

	for i, item := range topList {
		item["rank"] = i + 1
	}

	c.JSON(http.StatusOK, response.Success(topList))
}

func RefreshStatsCache(ctx context.Context, c *app.RequestContext) {
	keys := []string{"dashboard:", "heatmap:", "org:stats:"}
	deleted := 0
	for _, prefix := range keys {
		count, err := cache.DelByPrefix(ctx, prefix)
		if err != nil {
			logger.Error("Delete cache by prefix failed", logger.String("prefix", prefix), logger.Error(err))
			continue
		}
		deleted += count
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"deletedCount": deleted,
	}, "缓存刷新成功"))
}
