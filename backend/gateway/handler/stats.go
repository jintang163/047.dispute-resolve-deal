package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type KeywordStatItem struct {
	Keyword    string  `json:"keyword"`
	Count      int     `json:"count"`
	Ratio      float64 `json:"ratio"`
	Category   string  `json:"category"`
	SampleSize int     `json:"sampleSize"`
}

type HeatmapQueryRequest struct {
	StartTime      string `form:"startTime"`
	EndTime        string `form:"endTime"`
	TypeID         int64  `form:"typeId"`
	OrganizationID int64  `form:"organizationId"`
	UseSpatial     bool   `form:"useSpatial"`
}

type BBoxDrilldownRequest struct {
	WestLng        float64 `form:"westLng" vd:"gte:-180"`
	SouthLat       float64 `form:"southLat" vd:"gte:-90"`
	EastLng        float64 `form:"eastLng" vd:"lte:180"`
	NorthLat       float64 `form:"northLat" vd:"lte:90"`
	CenterLng      float64 `form:"centerLng"`
	CenterLat      float64 `form:"centerLat"`
	RadiusMeters   float64 `form:"radiusMeters"`
	StartTime      string  `form:"startTime"`
	EndTime        string  `form:"endTime"`
	TypeID         int64   `form:"typeId"`
	OrganizationID int64   `form:"organizationId"`
	GridKey        string  `form:"gridKey"`
	ClusterID      string  `form:"clusterId"`
	Page           int     `form:"page"`
	PageSize       int     `form:"pageSize"`
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

func applyOrgFilter(query interface{ Where(query interface{}, args ...interface{}) interface{} }, orgID int64) interface{} {
	if orgID <= 0 {
		return query
	}
	var childOrgs []int64
	database.GetDB().Table("sys_organization").
		Select("id").
		Where("parent_id = ? OR id = ?", orgID, orgID).
		Pluck("id", &childOrgs)
	if len(childOrgs) > 0 {
		return query.Where("dc.organization_id IN ?", childOrgs)
	}
	return query
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

	endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
	if err != nil {
		endTime = time.Now()
	}
	startTime, _ := time.Parse("2006-01-02 15:04:05", req.StartTime)
	daysDiff := int(endTime.Sub(startTime).Hours()/24) + 1
	if daysDiff <= 0 || daysDiff > 90 {
		daysDiff = 7
		if endTime.Before(startTime) {
			startTime = endTime.AddDate(0, 0, -6)
		}
	}

	cacheKey := fmt.Sprintf("heatmap:timeline:%d:%s:%s:%d:%t", orgID, req.StartTime, req.EndTime, req.TypeID, req.UseSpatial)
	cachedData, err := cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data []map[string]interface{}
		json.Unmarshal([]byte(cachedData), &data)
		c.JSON(http.StatusOK, response.Success(data))
		return
	}

	spatialConfig := config.GetConfig().Spatial
	useSpatial := req.UseSpatial && spatialConfig.UseSpatialIndex
	var childOrgs []int64
	if orgID > 0 {
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", orgID, orgID).
			Pluck("id", &childOrgs)
	}

	timeline := make([]map[string]interface{}, 0, daysDiff)

	for i := 0; i < daysDiff; i++ {
		dayStart := startTime.AddDate(0, 0, i).Format("2006-01-02") + " 00:00:00"
		dayEnd := startTime.AddDate(0, 0, i).Format("2006-01-02") + " 23:59:59"
		dayLabel := startTime.AddDate(0, 0, i).Format("2006-01-02")

		db := database.GetDB().Table("dispute_case dc").
			Select("dc.latitude, dc.longitude, dc.id, dc.case_no, dc.title, dt.type_name, so.org_name, dc.status, dc.created_at").
			Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
			Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
			Where("dc.deleted_at IS NULL").
			Where("dc.created_at >= ? AND dc.created_at <= ?", dayStart, dayEnd)

		if useSpatial {
			db = db.Where("dc.geom IS NOT NULL")
		} else {
			db = db.Where("dc.latitude IS NOT NULL AND dc.longitude IS NOT NULL").
				Where("dc.latitude != 0 AND dc.longitude != 0")
		}

		if len(childOrgs) > 0 {
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
			} else if status, ok := item["status"].(int64); ok {
				item["status_name"] = constants.CaseStatusMap[int(status)]
			} else if status, ok := item["status"].(int); ok {
				item["status_name"] = constants.CaseStatusMap[status]
			}
		}

		timeline = append(timeline, map[string]interface{}{
			"date":  dayLabel,
			"count": len(dayData),
			"items": dayData,
		})
	}

	jsonData, _ := json.Marshal(timeline)
	cache.Set(ctx, cacheKey, string(jsonData), 300)

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

	cacheKey := fmt.Sprintf("heatmap:topcluster:%d:%s:%s:%d:%d", orgID, req.StartTime, req.EndTime, req.TypeID, limit)
	cachedData, err := cache.Get(ctx, cacheKey)
	if err == nil && cachedData != "" {
		var data []map[string]interface{}
		json.Unmarshal([]byte(cachedData), &data)
		c.JSON(http.StatusOK, response.Success(data))
		return
	}

	spatialConfig := config.GetConfig().Spatial
	geoPrefix := spatialConfig.UseGeohashPrefix
	if geoPrefix <= 2 || geoPrefix > 8 {
		geoPrefix = 5
	}

	var childOrgs []int64
	if orgID > 0 {
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", orgID, orgID).
			Pluck("id", &childOrgs)
	}

	db := database.GetDB().Table("dispute_case dc").
		Select(fmt.Sprintf(
			"CAST(AVG(dc.longitude) AS DECIMAL(12,8)) as longitude, "+
				"CAST(AVG(dc.latitude) AS DECIMAL(12,8)) as latitude, "+
				"COUNT(*) as case_count, "+
				"LEFT(ST_GeoHash(dc.geom, 8), %d) as geo_key, "+
				"MIN(dc.longitude) as min_lng, "+
				"MIN(dc.latitude) as min_lat, "+
				"MAX(dc.longitude) as max_lng, "+
				"MAX(dc.latitude) as max_lat",
			geoPrefix)).
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Where("dc.deleted_at IS NULL").
		Where("dc.geom IS NOT NULL").
		Where("dc.longitude IS NOT NULL AND dc.latitude IS NOT NULL").
		Where("dc.longitude != 0 AND dc.latitude != 0")

	if len(childOrgs) > 0 {
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

	var geoGroups []map[string]interface{}
	db.Group("geo_key").
		Order("case_count DESC").
		Limit(limit * 2).
		Find(&geoGroups)

	clusterRadius := spatialConfig.ClusterRadiusMeters
	if clusterRadius <= 0 {
		clusterRadius = 500.0
	}

	type cluster struct {
		lngSum, latSum float64
		count          int64
		caseIDs        []int64
		minLng, minLat float64
		maxLng, maxLat float64
	}

	finalClusters := make([]*cluster, 0)

	for _, g := range geoGroups {
		lng, _ := strconv.ParseFloat(fmt.Sprintf("%v", g["longitude"]), 64)
		lat, _ := strconv.ParseFloat(fmt.Sprintf("%v", g["latitude"]), 64)
		count, _ := g["case_count"].(int64)
		minLng, _ := strconv.ParseFloat(fmt.Sprintf("%v", g["min_lng"]), 64)
		minLat, _ := strconv.ParseFloat(fmt.Sprintf("%v", g["min_lat"]), 64)
		maxLng, _ := strconv.ParseFloat(fmt.Sprintf("%v", g["max_lng"]), 64)
		maxLat, _ := strconv.ParseFloat(fmt.Sprintf("%v", g["max_lat"]), 64)

		merged := false
		for i, fc := range finalClusters {
			centerLng := fc.lngSum / float64(fc.count)
			centerLat := fc.latSum / float64(fc.count)
			dist := haversineDistance(lng, lat, centerLng, centerLat)
			if dist <= clusterRadius {
				fc.lngSum += lng * float64(count)
				fc.latSum += lat * float64(count)
				fc.count += count
				fc.minLng = math.Min(fc.minLng, minLng)
				fc.minLat = math.Min(fc.minLat, minLat)
				fc.maxLng = math.Max(fc.maxLng, maxLng)
				fc.maxLat = math.Max(fc.maxLat, maxLat)
				merged = true
				_ = i
				break
			}
		}
		if !merged {
			finalClusters = append(finalClusters, &cluster{
				lngSum:  lng * float64(count),
				latSum:  lat * float64(count),
				count:   count,
				minLng:  minLng,
				minLat:  minLat,
				maxLng:  maxLng,
				maxLat:  maxLat,
				caseIDs: []int64{},
			})
		}
	}

	resultList := make([]map[string]interface{}, 0, len(finalClusters))
	for i, fc := range finalClusters {
		if i >= limit {
			break
		}
		centerLng := fc.lngSum / float64(fc.count)
		centerLat := fc.latSum / float64(fc.count)

		clusterName := fmt.Sprintf("热点区域%d", i+1)
		var orgNameList []string
		database.GetDB().Table("dispute_case dc").
			Select("so.org_name").
			Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
			Where("ST_Contains(ST_MakeEnvelope(POINT(?, ?), POINT(?, ?)), dc.geom)",
				fc.minLng, fc.minLat, fc.maxLng, fc.maxLat).
			Group("so.org_name").
			Order("COUNT(*) DESC").
			Limit(2).
			Pluck("org_name", &orgNameList)
		if len(orgNameList) > 0 {
			clusterName = strings.Join(orgNameList, "-") + "片区"
		}

		resultList = append(resultList, map[string]interface{}{
			"cluster_id": fmt.Sprintf("CLS_%d", i+1),
			"rank":       i + 1,
			"cluster_name": clusterName,
			"longitude":   fmt.Sprintf("%.8f", centerLng),
			"latitude":    fmt.Sprintf("%.8f", centerLat),
			"case_count":  fc.count,
			"bbox": map[string]float64{
				"west":  fc.minLng,
				"south": fc.minLat,
				"east":  fc.maxLng,
				"north": fc.maxLat,
			},
			"radius_meters": int(clusterRadius),
		})
	}

	if len(resultList) == 0 && len(childOrgs) > 0 {
		fallbackList := make([]map[string]interface{}, 0)
		database.GetDB().Table("dispute_case dc").
			Select("so.id as org_id, so.org_name, so.longitude, so.latitude, COUNT(*) as case_count").
			Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
			Where("dc.deleted_at IS NULL").
			Where("so.longitude IS NOT NULL AND so.latitude IS NOT NULL").
			Where("dc.organization_id IN ?", childOrgs).
			Group("so.id, so.org_name, so.longitude, so.latitude").
			Order("case_count DESC").
			Limit(limit).
			Find(&fallbackList)
		for i, item := range fallbackList {
			item["rank"] = i + 1
			item["cluster_id"] = fmt.Sprintf("ORG_%v", item["org_id"])
			item["cluster_name"] = item["org_name"]
			lng, _ := strconv.ParseFloat(fmt.Sprintf("%v", item["longitude"]), 64)
			lat, _ := strconv.ParseFloat(fmt.Sprintf("%v", item["latitude"]), 64)
			eps := 0.0045
			item["bbox"] = map[string]float64{
				"west":  lng - eps,
				"south": lat - eps,
				"east":  lng + eps,
				"north": lat + eps,
			}
			item["radius_meters"] = 500
		}
		resultList = fallbackList
	}

	jsonData, _ := json.Marshal(resultList)
	cache.Set(ctx, cacheKey, string(jsonData), 300)

	c.JSON(http.StatusOK, response.Success(resultList))
}

func haversineDistance(lng1, lat1, lng2, lat2 float64) float64 {
	const earthRadius = 6371000.0
	toRad := math.Pi / 180.0
	dLat := (lat2 - lat1) * toRad
	dLng := (lng2 - lng1) * toRad
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*toRad)*math.Cos(lat2*toRad)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadius * c
}

func GetHeatmapDrilldown(ctx context.Context, c *app.RequestContext) {
	var req BBoxDrilldownRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := req.OrganizationID
	if orgID == 0 {
		orgID = userInfo.OrganizationID
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	useSpatial := config.GetConfig().Spatial.UseSpatialIndex

	db := database.GetDB().Table("dispute_case dc").
		Select("dc.id, dc.case_no, dc.title, dc.applicant_name, dc.respondent_name, "+
			"dc.event_address, dc.latitude, dc.longitude, dc.status, dc.created_at, "+
			"dt.type_name, so.org_name").
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
		Where("dc.deleted_at IS NULL").
		Where("dc.latitude IS NOT NULL AND dc.longitude IS NOT NULL").
		Where("dc.latitude != 0 AND dc.longitude != 0")

	hasBBox := req.WestLng != 0 || req.SouthLat != 0 || req.EastLng != 0 || req.NorthLat != 0
	if hasBBox && req.EastLng > req.WestLng && req.NorthLat > req.SouthLat {
		if useSpatial {
			db = db.Where("ST_Contains(ST_MakeEnvelope(POINT(?, ?), POINT(?, ?)), dc.geom)",
				req.WestLng, req.SouthLat, req.EastLng, req.NorthLat)
		} else {
			db = db.Where("dc.longitude >= ? AND dc.longitude <= ? AND dc.latitude >= ? AND dc.latitude <= ?",
				req.WestLng, req.EastLng, req.SouthLat, req.NorthLat)
		}
	} else if req.CenterLng != 0 && req.CenterLat != 0 && req.RadiusMeters > 0 {
		r := req.RadiusMeters
		epsLat := r / 111320.0
		epsLng := r / (111320.0 * math.Cos(req.CenterLat*math.Pi/180.0))
		if useSpatial {
			db = db.Where("ST_Distance_Sphere(dc.geom, ST_SRID(POINT(?, ?), 4326)) <= ?",
				req.CenterLng, req.CenterLat, r)
		}
		db = db.Where("dc.longitude >= ? AND dc.longitude <= ? AND dc.latitude >= ? AND dc.latitude <= ?",
			req.CenterLng-epsLng, req.CenterLng+epsLng,
			req.CenterLat-epsLat, req.CenterLat+epsLat)
	}

	var childOrgs []int64
	if orgID > 0 {
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", orgID, orgID).
			Pluck("id", &childOrgs)
		if len(childOrgs) > 0 {
			db = db.Where("dc.organization_id IN ?", childOrgs)
		}
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

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	offset := (req.Page - 1) * req.PageSize
	db.Order("dc.created_at DESC").
		Limit(req.PageSize).
		Offset(offset).
		Find(&list)

	for _, item := range list {
		if status, ok := item["status"].(int32); ok {
			item["status_name"] = constants.CaseStatusMap[int(status)]
		} else if status, ok := item["status"].(int64); ok {
			item["status_name"] = constants.CaseStatusMap[int(status)]
		} else if status, ok := item["status"].(int); ok {
			item["status_name"] = constants.CaseStatusMap[status]
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
		"list":     list,
	}))
}

func GetAmapConfig(ctx context.Context, c *app.RequestContext) {
	amapCfg := config.GetConfig().Amap
	spatialCfg := config.GetConfig().Spatial
	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"web_key":        amapCfg.WebKey,
		"security_code":  amapCfg.SecurityCode,
		"default_city":   amapCfg.DefaultCity,
		"default_lng":    amapCfg.DefaultLng,
		"default_lat":    amapCfg.DefaultLat,
		"default_zoom":   amapCfg.DefaultZoom,
		"cluster_radius": spatialCfg.ClusterRadiusMeters,
		"grid_level":     spatialCfg.GridLevel,
		"use_spatial":    spatialCfg.UseSpatialIndex,
	}))
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

func GetKeywordStats(ctx context.Context, c *app.RequestContext) {
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 30
	}
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	typeID, _ := strconv.ParseInt(c.DefaultQuery("typeId", "0"), 10, 64)
	orgID, _ := strconv.ParseInt(c.DefaultQuery("organizationId", "0"), 10, 64)

	startDate := time.Now().AddDate(0, 0, -days)

	baseDB := database.GetDB().Table("dispute_case dc").
		Where("dc.deleted_at IS NULL").
		Where("dc.created_at >= ?", startDate).
		Where("dc.keywords IS NOT NULL")

	if typeID > 0 {
		typePath := fmt.Sprintf("/%d/", typeID)
		baseDB = baseDB.Where("(dc.type_id = ? OR dc.type_path LIKE ?)", typeID, typePath+"%")
	}
	if orgID > 0 {
		baseDB = baseDB.Where("dc.organization_id = ?", orgID)
	}

	var rows []map[string]interface{}
	baseDB.Select("dc.id, dc.keywords").
		Order("dc.created_at DESC").
		Limit(5000).
		Scan(&rows)

	counter := make(map[string]int)
	caseIDs := make(map[string][]int64)
	totalCases := int64(0)

	for _, row := range rows {
		totalCases++
		caseID, _ := toInt64Safe(row["id"])
		raw := row["keywords"]
		var kws []string
		switch v := raw.(type) {
		case []byte:
			_ = json.Unmarshal(v, &kws)
		case string:
			_ = json.Unmarshal([]byte(v), &kws)
		}
		for _, kw := range kws {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}
			counter[kw]++
			caseIDs[kw] = append(caseIDs[kw], caseID)
		}
	}

	typeItem struct {
		Keyword    string  `json:"keyword"`
		Count      int     `json:"count"`
		Ratio      float64 `json:"ratio"`
		Category   string  `json:"category"`
		SampleSize int     `json:"sampleSize"`
	}
	_ = typeItem{}

	list := make([]*KeywordStatItem, 0, len(counter))
	for kw, cnt := range counter {
		item := &KeywordStatItem{
			Keyword:    kw,
			Count:      cnt,
			SampleSize: len(caseIDs[kw]),
		}
		if totalCases > 0 {
			item.Ratio = float64(cnt) / float64(totalCases)
		}
		list = append(list, item)
	}

	sort.Slice(list, func(i, j int) bool {
		return list[i].Count > list[j].Count
	})
	if len(list) > limit {
		list = list[:limit]
	}

	if len(list) > 0 {
		keywordsList := make([]string, 0, len(list))
		for _, it := range list {
			keywordsList = append(keywordsList, it.Keyword)
		}
		var dictItems []map[string]interface{}
		database.GetDB().Table("dispute_keyword_dict").
			Select("keyword, category").
			Where("keyword IN ?", keywordsList).
			Find(&dictItems)
		kwCat := make(map[string]string)
		for _, d := range dictItems {
			if k, ok := d["keyword"].(string); ok {
				if cat, ok2 := d["category"].(string); ok2 {
					kwCat[k] = cat
				}
			}
		}
		for _, it := range list {
			if cat, ok := kwCat[it.Keyword]; ok {
				it.Category = cat
			}
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"list":        list,
		"total_cases": totalCases,
		"unique_kws":  len(counter),
		"days":        days,
	}))
}

func toInt64Safe(v interface{}) (int64, bool) {
	if v == nil {
		return 0, false
	}
	switch val := v.(type) {
	case int64:
		return val, true
	case int32:
		return int64(val), true
	case int:
		return int64(val), true
	case float64:
		return int64(val), true
	case string:
		n, e := strconv.ParseInt(val, 10, 64)
		return n, e == nil
	}
	return 0, false
}
