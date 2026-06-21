package handler

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	common "github.com/dispute-resolve/common/model"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type PerformanceScoreRequest struct {
	UserID      int64   `json:"userId" binding:"required"`
	Period      int32   `json:"period" binding:"required"`
	Year        int32   `json:"year" binding:"required"`
	Month       int32   `json:"month"`
	Quarter     int32   `json:"quarter"`
	CaseCount   float64 `json:"caseCount"`
	CloseRate   float64 `json:"closeRate"`
	SuccessRate float64 `json:"successRate"`
	AvgDays     float64 `json:"avgDays"`
	Satisfaction float64 `json:"satisfaction"`
	Remark      string  `json:"remark"`
}

func GetPerformanceScoreList(ctx context.Context, c *app.RequestContext) {
	var req struct {
		common.BaseQuery
		Period         int32 `form:"period"`
		Year           int32 `form:"year"`
		UserID         int64 `form:"userId"`
		OrganizationID int64 `form:"organizationId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	db := database.GetDB().Table("performance_score ps").
		Select("ps.*, su.real_name, so.org_name").
		Joins("LEFT JOIN sys_user su ON ps.user_id = su.id").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("ps.deleted_at IS NULL")

	if userInfo.Role == constants.RoleLeader {
		var childOrgs []int64
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", userInfo.OrganizationID, userInfo.OrganizationID).
			Pluck("id", &childOrgs)
		db = db.Where("su.organization_id IN ?", childOrgs)
	} else if userInfo.Role == constants.RoleMediator {
		db = db.Where("ps.user_id = ?", userInfo.UserID)
	}

	if req.Period > 0 {
		db = db.Where("ps.period = ?", req.Period)
	}
	if req.Year > 0 {
		db = db.Where("ps.year = ?", req.Year)
	}
	if req.UserID > 0 {
		db = db.Where("ps.user_id = ?", req.UserID)
	}
	if req.OrganizationID > 0 {
		db = db.Where("su.organization_id = ?", req.OrganizationID)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("ps.total_score DESC, ps.created_at DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	periodMap := map[int]string{
		constants.PerformancePeriodMonth:   "月度",
		constants.PerformancePeriodQuarter: "季度",
		constants.PerformancePeriodYear:    "年度",
	}

	for i, item := range list {
		if p, ok := item["period"].(int); ok {
			item["period_name"] = periodMap[p]
		}
		item["rank"] = i + 1
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetMyPerformance(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	period, _ := strconv.Atoi(c.DefaultQuery("period", "1"))

	var score map[string]interface{}
	db := database.GetDB().Table("performance_score ps").
		Select("ps.*").
		Where("ps.user_id = ?", userInfo.UserID).
		Where("ps.year = ?", year).
		Where("ps.period = ?", period)

	if period == constants.PerformancePeriodMonth {
		month := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))
		db = db.Where("ps.month = ?", month)
	} else if period == constants.PerformancePeriodQuarter {
		quarter := c.DefaultQuery("quarter", "1")
		db = db.Where("ps.quarter = ?", quarter)
	}

	db.Find(&score)

	var indicators []map[string]interface{}
	database.GetDB().Table("performance_indicator_config").
		Where("status = 1").
		Order("sort_order ASC").
		Find(&indicators)

	result := map[string]interface{}{
		"score":      score,
		"indicators": indicators,
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func GetPerformanceDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var score map[string]interface{}
	result := database.GetDB().Table("performance_score ps").
		Select("ps.*, su.real_name, so.org_name").
		Joins("LEFT JOIN sys_user su ON ps.user_id = su.id").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("ps.id = ?", id).
		Find(&score)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("考核记录不存在"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role == constants.RoleMediator && score["user_id"].(int64) != userInfo.UserID {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限查看他人考核记录"))
		return
	}

	periodMap := map[int]string{
		constants.PerformancePeriodMonth:   "月度",
		constants.PerformancePeriodQuarter: "季度",
		constants.PerformancePeriodYear:    "年度",
	}
	if p, ok := score["period"].(int); ok {
		score["period_name"] = periodMap[p]
	}

	c.JSON(http.StatusOK, response.Success(score))
}

func CalculatePerformanceScore(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限计算考核分数"))
		return
	}

	var req struct {
		UserID  int64 `json:"userId" binding:"required"`
		Period  int32 `json:"period" binding:"required"`
		Year    int32 `json:"year" binding:"required"`
		Month   int32 `json:"month"`
		Quarter int32 `json:"quarter"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var user struct {
		RealName       string `gorm:"column:real_name"`
		OrganizationID int64  `gorm:"column:organization_id"`
	}
	database.GetDB().Table("sys_user").
		Select("real_name, organization_id").
		Where("id = ?", req.UserID).
		First(&user)

	var startDate, endDate time.Time
	year := int(req.Year)

	switch req.Period {
	case constants.PerformancePeriodMonth:
		month := int(req.Month)
		if month == 0 {
			month = int(time.Now().Month())
		}
		startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 1, 0)
	case constants.PerformancePeriodQuarter:
		quarter := int(req.Quarter)
		if quarter == 0 {
			quarter = 1
		}
		startMonth := (quarter-1)*3 + 1
		startDate = time.Date(year, time.Month(startMonth), 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 3, 0)
	case constants.PerformancePeriodYear:
		startDate = time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(1, 0, 0)
	}

	var totalCases int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Count(&totalCases)

	var closedCases int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Count(&closedCases)

	var successCases int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("mediation_result = ?", constants.MediationResultSuccess).
		Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Count(&successCases)

	var avgDays float64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Select("AVG(TIMESTAMPDIFF(DAY, created_at, closed_time))").
		Scan(&avgDays)

	var avgSatisfaction float64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Where("satisfaction_score > 0").
		Select("AVG(satisfaction_score)").
		Scan(&avgSatisfaction)

	var urgeCount int64
	database.GetDB().Table("workflow_urge").
		Where("current_handler_id = ?", req.UserID).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Count(&urgeCount)

	var indicators []map[string]interface{}
	database.GetDB().Table("performance_indicator_config").
		Where("status = 1").
		Find(&indicators)

	indicatorScoreMap := make(map[string]float64)
	totalWeight := 0.0
	totalScore := 0.0

	for _, ind := range indicators {
		code, _ := ind["indicator_code"].(string)
		weight, _ := ind["weight"].(float64)
		if code == "" {
			continue
		}
		totalWeight += weight
		score := calculateIndicatorScore(code, totalCases, closedCases, successCases, avgDays, avgSatisfaction, urgeCount)
		indicatorScoreMap[code] = score
		totalScore += score * weight
	}

	if totalWeight > 0 {
		totalScore = totalScore / totalWeight
	}
	totalScore = math.Round(totalScore*100) / 100

	level := "C"
	if totalScore >= 90 {
		level = "S"
	} else if totalScore >= 80 {
		level = "A"
	} else if totalScore >= 70 {
		level = "B"
	} else if totalScore >= 60 {
		level = "C"
	} else {
		level = "D"
	}

	closeRate := 0.0
	if totalCases > 0 {
		closeRate = math.Round(float64(closedCases)/float64(totalCases)*10000) / 100
	}
	successRate := 0.0
	if closedCases > 0 {
		successRate = math.Round(float64(successCases)/float64(closedCases)*10000) / 100
	}

	scoreID := utils.GenerateID()
	scoreData := map[string]interface{}{
		"id":                scoreID,
		"user_id":           req.UserID,
		"user_name":         user.RealName,
		"period":            req.Period,
		"year":              req.Year,
		"month":             req.Month,
		"quarter":           req.Quarter,
		"start_date":        startDate,
		"end_date":          endDate,
		"case_count":        totalCases,
		"closed_count":      closedCases,
		"success_count":     successCases,
		"close_rate":        closeRate,
		"success_rate":      successRate,
		"avg_days":          math.Round(avgDays*100) / 100,
		"satisfaction":      math.Round(avgSatisfaction*100) / 100,
		"urge_count":        urgeCount,
		"total_score":       totalScore,
		"level":             level,
		"calculated_by":     userInfo.UserID,
		"calculated_by_name": userInfo.RealName,
		"organization_id":   user.OrganizationID,
	}

	for code, sc := range indicatorScoreMap {
		scoreData[code+"_score"] = math.Round(sc*100) / 100
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("performance_score").Create(scoreData).Error; err != nil {
		tx.Rollback()
		logger.Error("Create performance score failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("考核计算失败"))
		return
	}

	snapshotData := map[string]interface{}{
		"user_id":          req.UserID,
		"user_name":        user.RealName,
		"org_id":           user.OrganizationID,
		"year":             req.Year,
		"month":            req.Month,
		"case_count":       totalCases,
		"closed_count":     closedCases,
		"close_rate":       closeRate,
		"success_count":    successCases,
		"success_rate":     successRate,
		"avg_days":         math.Round(avgDays*100) / 100,
		"avg_satisfaction": math.Round(avgSatisfaction*100) / 100,
		"urge_count":       urgeCount,
		"total_score":      totalScore,
		"level":            level,
	}
	var orgName string
	database.GetDB().Table("sys_organization").Select("org_name").Where("id = ?", user.OrganizationID).Scan(&orgName)
	snapshotData["org_name"] = orgName
	tx.Table("performance_monthly_snapshot").Where("user_id = ? AND year = ? AND month = ?", req.UserID, req.Year, req.Month).Delete(nil)
	tx.Table("performance_monthly_snapshot").Create(snapshotData)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id":          scoreID,
		"totalScore":  totalScore,
		"level":       level,
		"caseCount":   totalCases,
		"closedCount": closedCases,
		"successCount": successCases,
		"closeRate":   closeRate,
		"successRate": successRate,
		"avgDays":     math.Round(avgDays*100) / 100,
		"satisfaction": math.Round(avgSatisfaction*100) / 100,
		"urgeCount":   urgeCount,
	}, "考核计算成功"))
}

func calculateIndicatorScore(code string, totalCases, closedCases, successCases int64, avgDays, avgSatisfaction float64, urgeCount int64) float64 {
	switch code {
	case "CASE_COUNT":
		target := 30.0
		if float64(totalCases) >= target {
			return 100
		}
		return float64(totalCases) / target * 100
	case "CLOSE_RATE":
		if totalCases == 0 {
			return 0
		}
		rate := float64(closedCases) / float64(totalCases) * 100
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
	case "SUCCESS_RATE":
		if closedCases == 0 {
			return 0
		}
		rate := float64(successCases) / float64(closedCases) * 100
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
	case "AVG_DAYS":
		if avgDays <= 0 {
			return 100
		}
		switch {
		case avgDays <= 7:
			return 100
		case avgDays <= 14:
			return 90
		case avgDays <= 21:
			return 80
		case avgDays <= 30:
			return 70
		case avgDays <= 45:
			return 60
		default:
			return math.Max(0, 100-avgDays)
		}
	case "SATISFACTION":
		switch {
		case avgSatisfaction >= 4.8:
			return 100
		case avgSatisfaction >= 4.5:
			return 90
		case avgSatisfaction >= 4.2:
			return 80
		case avgSatisfaction >= 4.0:
			return 70
		case avgSatisfaction >= 3.5:
			return 60
		default:
			return avgSatisfaction * 15
		}
	case "URGE_COUNT":
		switch {
		case urgeCount == 0:
			return 100
		case urgeCount <= 1:
			return 90
		case urgeCount <= 3:
			return 70
		case urgeCount <= 5:
			return 50
		default:
			return math.Max(0, 50-float64(urgeCount-5)*10)
		}
	default:
		return 0
	}
}

func GetPerformanceDashboard(ctx context.Context, c *app.RequestContext) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	orgID, _ := strconv.ParseInt(c.DefaultQuery("organizationId"), 10, 64)

	userInfo := middleware.GetUserInfo(c)
	if orgID == 0 && userInfo.Role == constants.RoleLeader {
		orgID = userInfo.OrganizationID
	}

	db := database.GetDB().Table("performance_monthly_snapshot pms").
		Where("pms.year = ? AND pms.month = ?", year, month)

	if orgID > 0 {
		db = db.Where("pms.org_id = ?", orgID)
	}
	if userInfo.Role == constants.RoleMediator {
		db = db.Where("pms.user_id = ?", userInfo.UserID)
	}

	var snapshots []map[string]interface{}
	db.Order("pms.total_score DESC").Find(&snapshots)

	useRealtime := len(snapshots) == 0
	if useRealtime {
		snapshots = aggregatePerformanceFromBusinessTables(year, month, orgID, userInfo)
	}

	totalCases := 0
	totalClosed := 0
	totalSuccess := 0
	totalUrge := 0
	totalAvgDays := 0.0
	totalSatisfaction := 0.0
	totalScore := 0.0
	count := len(snapshots)

	for _, s := range snapshots {
		if v, ok := s["case_count"].(int); ok {
			totalCases += v
		} else if v, ok := s["case_count"].(int64); ok {
			totalCases += int(v)
		}
		if v, ok := s["closed_count"].(int); ok {
			totalClosed += v
		} else if v, ok := s["closed_count"].(int64); ok {
			totalClosed += int(v)
		}
		if v, ok := s["success_count"].(int); ok {
			totalSuccess += v
		} else if v, ok := s["success_count"].(int64); ok {
			totalSuccess += int(v)
		}
		if v, ok := s["urge_count"].(int); ok {
			totalUrge += v
		} else if v, ok := s["urge_count"].(int64); ok {
			totalUrge += int(v)
		}
		if v, ok := s["avg_days"].(float64); ok {
			totalAvgDays += v
		}
		if v, ok := s["avg_satisfaction"].(float64); ok {
			totalSatisfaction += v
		}
		if v, ok := s["total_score"].(float64); ok {
			totalScore += v
		}
	}

	avgCloseRate := 0.0
	if totalCases > 0 {
		avgCloseRate = math.Round(float64(totalClosed)/float64(totalCases)*10000) / 100
	}
	avgSuccessRate := 0.0
	if totalClosed > 0 {
		avgSuccessRate = math.Round(float64(totalSuccess)/float64(totalClosed)*10000) / 100
	}
	avgDays := 0.0
	if count > 0 {
		avgDays = math.Round(totalAvgDays/float64(count)*100) / 100
	}
	avgSatisfaction := 0.0
	if count > 0 {
		avgSatisfaction = math.Round(totalSatisfaction/float64(count)*100) / 100
	}
	avgScore := 0.0
	if count > 0 {
		avgScore = math.Round(totalScore/float64(count)*100) / 100
	}

	summary := map[string]interface{}{
		"year":             year,
		"month":            month,
		"mediatorCount":    count,
		"totalCases":       totalCases,
		"totalClosed":      totalClosed,
		"totalSuccess":     totalSuccess,
		"totalUrge":        totalUrge,
		"avgCloseRate":     avgCloseRate,
		"avgSuccessRate":   avgSuccessRate,
		"avgDays":          avgDays,
		"avgSatisfaction":  avgSatisfaction,
		"avgScore":         avgScore,
		"dataSource":       map[bool]string{true: "realtime", false: "snapshot"}[useRealtime],
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"summary":    summary,
		"mediators":  snapshots,
	}))
}

func aggregatePerformanceFromBusinessTables(year, month int, orgID int64, userInfo *middleware.UserClaims) []map[string]interface{} {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0)

	var mediatorIDs []int64
	userDB := database.GetDB().Table("sys_user su").
		Select("su.id, su.real_name, su.organization_id, so.org_name").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("su.role_id = ?", constants.RoleMediator).
		Where("su.status = 1")

	if orgID > 0 {
		userDB = userDB.Where("su.organization_id = ?", orgID)
	}
	if userInfo.Role == constants.RoleMediator {
		userDB = userDB.Where("su.id = ?", userInfo.UserID)
	}
	var users []map[string]interface{}
	userDB.Find(&users)

	var indicators []map[string]interface{}
	database.GetDB().Table("performance_indicator_config").
		Where("status = 1").
		Find(&indicators)

	var results []map[string]interface{}
	for _, user := range users {
		userID, _ := user["id"].(int64)
		realName, _ := user["real_name"].(string)
		organizationID, _ := user["organization_id"].(int64)
		orgName, _ := user["org_name"].(string)

		var totalCases int64
		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("created_at >= ? AND created_at < ?", startDate, endDate).
			Where("deleted_at IS NULL").
			Count(&totalCases)

		var closedCases int64
		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("status = ?", constants.CaseStatusClosed).
			Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
			Where("deleted_at IS NULL").
			Count(&closedCases)

		var successCases int64
		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("status = ?", constants.CaseStatusClosed).
			Where("mediation_result = ?", constants.MediationResultSuccess).
			Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
			Where("deleted_at IS NULL").
			Count(&successCases)

		var avgDays float64
		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("status = ?", constants.CaseStatusClosed).
			Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
			Where("deleted_at IS NULL").
			Select("AVG(TIMESTAMPDIFF(DAY, created_at, closed_time))").
			Scan(&avgDays)

		var avgSatisfaction float64
		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("status = ?", constants.CaseStatusClosed).
			Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
			Where("deleted_at IS NULL").
			Where("satisfaction_score > 0").
			Select("AVG(satisfaction_score)").
			Scan(&avgSatisfaction)

		var urgeCount int64
		database.GetDB().Table("workflow_urge").
			Where("current_handler_id = ?", userID).
			Where("created_at >= ? AND created_at < ?", startDate, endDate).
			Count(&urgeCount)

		indicatorScoreMap := make(map[string]float64)
		totalWeight := 0.0
		totalScore := 0.0

		for _, ind := range indicators {
			code, _ := ind["indicator_code"].(string)
			weight, _ := ind["weight"].(float64)
			if code == "" {
				continue
			}
			totalWeight += weight
			score := calculateIndicatorScore(code, totalCases, closedCases, successCases, avgDays, avgSatisfaction, urgeCount)
			indicatorScoreMap[code] = score
			totalScore += score * weight
		}

		if totalWeight > 0 {
			totalScore = totalScore / totalWeight
		}
		totalScore = math.Round(totalScore*100) / 100

		level := "C"
		if totalScore >= 90 {
			level = "S"
		} else if totalScore >= 80 {
			level = "A"
		} else if totalScore >= 70 {
			level = "B"
		} else if totalScore >= 60 {
			level = "C"
		} else {
			level = "D"
		}

		closeRate := 0.0
		if totalCases > 0 {
			closeRate = math.Round(float64(closedCases)/float64(totalCases)*10000) / 100
		}
		successRate := 0.0
		if closedCases > 0 {
			successRate = math.Round(float64(successCases)/float64(closedCases)*10000) / 100
		}

		result := map[string]interface{}{
			"user_id":          userID,
			"user_name":        realName,
			"org_id":           organizationID,
			"org_name":         orgName,
			"year":             year,
			"month":            month,
			"case_count":       int(totalCases),
			"closed_count":     int(closedCases),
			"close_rate":       closeRate,
			"success_count":    int(successCases),
			"success_rate":     successRate,
			"avg_days":         math.Round(avgDays*100) / 100,
			"avg_satisfaction": math.Round(avgSatisfaction*100) / 100,
			"urge_count":       int(urgeCount),
			"total_score":      totalScore,
			"level":            level,
		}

		for code, sc := range indicatorScoreMap {
			result[code+"_score"] = math.Round(sc*100) / 100
		}

		results = append(results, result)
	}

	return results
}

func GetPerformanceMonthComparison(ctx context.Context, c *app.RequestContext) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	orgID, _ := strconv.ParseInt(c.DefaultQuery("organizationId"), 10, 64)
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)

	userInfo := middleware.GetUserInfo(c)
	if userID == 0 && userInfo.Role == constants.RoleMediator {
		userID = userInfo.UserID
	}

	prevMonth := month - 1
	prevYear := year
	if prevMonth == 0 {
		prevMonth = 12
		prevYear = year - 1
	}

	type monthAgg struct {
		CaseCount      int     `gorm:"column:case_count"`
		ClosedCount    int     `gorm:"column:closed_count"`
		CloseRate      float64 `gorm:"column:close_rate"`
		SuccessCount   int     `gorm:"column:success_count"`
		SuccessRate    float64 `gorm:"column:success_rate"`
		AvgDays        float64 `gorm:"column:avg_days"`
		AvgSatisfaction float64 `gorm:"column:avg_satisfaction"`
		UrgeCount      int     `gorm:"column:urge_count"`
		TotalScore     float64 `gorm:"column:total_score"`
	}

	fetchAgg := func(y, m int) map[string]interface{} {
		db := database.GetDB().Table("performance_monthly_snapshot").
			Where("year = ? AND month = ?", y, m)
		if userID > 0 {
			db = db.Where("user_id = ?", userID)
		}
		if orgID > 0 {
			db = db.Where("org_id = ?", orgID)
		}
		if userInfo.Role == constants.RoleLeader && orgID == 0 {
			db = db.Where("org_id = ?", userInfo.OrganizationID)
		}

		var agg monthAgg
		db.Select("SUM(case_count) as case_count, SUM(closed_count) as closed_count, "+
			"AVG(close_rate) as close_rate, SUM(success_count) as success_count, "+
			"AVG(success_rate) as success_rate, AVG(avg_days) as avg_days, "+
			"AVG(avg_satisfaction) as avg_satisfaction, SUM(urge_count) as urge_count, "+
			"AVG(total_score) as total_score").
			Scan(&agg)

		return map[string]interface{}{
			"year":             y,
			"month":            m,
			"caseCount":        agg.CaseCount,
			"closedCount":      agg.ClosedCount,
			"closeRate":        math.Round(agg.CloseRate*100) / 100,
			"successCount":     agg.SuccessCount,
			"successRate":      math.Round(agg.SuccessRate*100) / 100,
			"avgDays":          math.Round(agg.AvgDays*100) / 100,
			"avgSatisfaction":  math.Round(agg.AvgSatisfaction*100) / 100,
			"urgeCount":        agg.UrgeCount,
			"totalScore":       math.Round(agg.TotalScore*100) / 100,
		}
	}

	current := fetchAgg(year, month)
	previous := fetchAgg(prevYear, prevMonth)

	calcChange := func(curr, prev float64) float64 {
		if prev == 0 {
			return 0
		}
		return math.Round((curr-prev)/prev*10000) / 100
	}
	calcChangeInt := func(curr, prev int) float64 {
		if prev == 0 {
			return 0
		}
		return math.Round(float64(curr-prev)/float64(prev)*10000) / 100
	}

	comparison := map[string]interface{}{
		"caseCountChange":       calcChangeInt(current["caseCount"].(int), previous["caseCount"].(int)),
		"closeRateChange":       calcChange(current["closeRate"].(float64), previous["closeRate"].(float64)),
		"successRateChange":     calcChange(current["successRate"].(float64), previous["successRate"].(float64)),
		"avgDaysChange":         calcChange(current["avgDays"].(float64), previous["avgDays"].(float64)),
		"avgSatisfactionChange": calcChange(current["avgSatisfaction"].(float64), previous["avgSatisfaction"].(float64)),
		"urgeCountChange":       calcChangeInt(current["urgeCount"].(int), previous["urgeCount"].(int)),
		"totalScoreChange":      calcChange(current["totalScore"].(float64), previous["totalScore"].(float64)),
	}

	var trend []map[string]interface{}
	trendDB := database.GetDB().Table("performance_monthly_snapshot").
		Select("year, month, AVG(close_rate) as close_rate, AVG(success_rate) as success_rate, "+
			"AVG(avg_days) as avg_days, AVG(avg_satisfaction) as avg_satisfaction, "+
			"SUM(urge_count) as urge_count, AVG(total_score) as total_score").
		Where("year = ?", year)
	if userID > 0 {
		trendDB = trendDB.Where("user_id = ?", userID)
	}
	if orgID > 0 {
		trendDB = trendDB.Where("org_id = ?", orgID)
	}
	trendDB.Group("year, month").Order("month ASC").Find(&trend)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"current":    current,
		"previous":   previous,
		"comparison": comparison,
		"trend":      trend,
	}))
}

func GetPerformanceTrend(ctx context.Context, c *app.RequestContext) {
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))

	userInfo := middleware.GetUserInfo(c)
	if userID == 0 {
		userID = userInfo.UserID
	}

	var snapshots []map[string]interface{}
	database.GetDB().Table("performance_monthly_snapshot").
		Where("user_id = ? AND year = ?", userID, year).
		Order("month ASC").
		Find(&snapshots)

	monthData := make([]map[string]interface{}, 12)
	for i := 0; i < 12; i++ {
		monthData[i] = map[string]interface{}{
			"month":           i + 1,
			"caseCount":       0,
			"closedCount":     0,
			"closeRate":       0,
			"successRate":     0,
			"avgDays":         0,
			"avgSatisfaction": 0,
			"urgeCount":       0,
			"totalScore":      0,
			"level":           "-",
			"hasData":         false,
		}
	}

	for _, s := range snapshots {
		m, _ := s["month"].(int)
		if m < 1 || m > 12 {
			continue
		}
		idx := m - 1
		monthData[idx]["hasData"] = true
		for _, key := range []string{"caseCount", "closedCount", "closeRate", "successRate", "avgDays", "avgSatisfaction", "urgeCount", "totalScore"} {
			if v, ok := s[key]; ok {
				monthData[idx][key] = v
			}
		}
		if v, ok := s["level"]; ok {
			monthData[idx]["level"] = v
		}
	}

	c.JSON(http.StatusOK, response.Success(monthData))
}

func GetPerformanceRanking(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	period, _ := strconv.Atoi(c.DefaultQuery("period", "1"))

	db := database.GetDB().Table("performance_score ps").
		Select("ps.user_id, ps.user_name, so.org_name, ps.total_score, ps.level, "+
			"ps.case_count, ps.closed_count, ps.success_count").
		Joins("LEFT JOIN sys_user su ON ps.user_id = su.id").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("ps.year = ?", year).
		Where("ps.period = ?", period).
		Where("ps.deleted_at IS NULL")

	if userInfo.Role == constants.RoleLeader {
		var childOrgs []int64
		database.GetDB().Table("sys_organization").
			Select("id").
			Where("parent_id = ? OR id = ?", userInfo.OrganizationID, userInfo.OrganizationID).
			Pluck("id", &childOrgs)
		db = db.Where("su.organization_id IN ?", childOrgs)
	}

	if period == constants.PerformancePeriodMonth {
		month := c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month())))
		db = db.Where("ps.month = ?", month)
	} else if period == constants.PerformancePeriodQuarter {
		quarter := c.DefaultQuery("quarter", "1")
		db = db.Where("ps.quarter = ?", quarter)
	}

	var rankings []map[string]interface{}
	db.Order("ps.total_score DESC").
		Limit(20).
		Find(&rankings)

	for i, item := range rankings {
		item["rank"] = i + 1
	}

	c.JSON(http.StatusOK, response.Success(rankings))
}

func GetPerformanceIndicatorConfig(ctx context.Context, c *app.RequestContext) {
	var configs []map[string]interface{}
	database.GetDB().Table("performance_indicator_config").
		Where("status = 1").
		Order("sort_order ASC").
		Find(&configs)

	totalWeight := 0.0
	for _, cfg := range configs {
		if w, ok := cfg["weight"].(float64); ok {
			totalWeight += w
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"indicators":  configs,
		"totalWeight": totalWeight,
	}))
}

func UpdatePerformanceIndicatorConfig(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector {
		c.JSON(http.StatusForbidden, response.Forbidden("仅管理员或主任可调整考核权重"))
		return
	}

	var req struct {
		Indicators []struct {
			ID     int64   `json:"id" binding:"required"`
			Weight float64 `json:"weight" binding:"required"`
		} `json:"indicators" binding:"required"`
		AutoRecalculate bool `json:"autoRecalculate"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	totalWeight := 0.0
	for _, ind := range req.Indicators {
		totalWeight += ind.Weight
	}
	if math.Abs(totalWeight-1.0) > 0.01 {
		c.JSON(http.StatusBadRequest, response.BadRequest(fmt.Sprintf("权重总和必须为1.0，当前为%.2f", totalWeight)))
		return
	}

	tx := database.GetDB().Begin()
	for _, ind := range req.Indicators {
		if err := tx.Table("performance_indicator_config").
			Where("id = ?", ind.ID).
			Update("weight", ind.Weight).Error; err != nil {
			tx.Rollback()
			logger.Error("Update indicator weight failed", logger.Error(err))
			c.JSON(http.StatusInternalServerError, response.ServerError("更新权重失败"))
			return
		}
	}
	tx.Commit()

	recalculateCount := 0
	if req.AutoRecalculate {
		year := time.Now().Year()
		month := int(time.Now().Month())
		recalculateCount = batchRecalculateAllMediators(year, month, userInfo)
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"recalculateCount": recalculateCount,
	}, "权重更新成功"))
}

func BatchCalculatePerformanceScore(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限计算考核分数"))
		return
	}

	var req struct {
		Year  int   `json:"year" binding:"required"`
		Month int   `json:"month" binding:"required"`
		OrgID int64 `json:"organizationId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	count := batchRecalculateAllMediators(req.Year, req.Month, userInfo)

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"count": count,
	}, fmt.Sprintf("批量计算完成，共计算 %d 名调解员绩效", count)))
}

func batchRecalculateAllMediators(year, month int, userInfo *middleware.UserClaims) int {
	orgID := int64(0)
	if userInfo.Role == constants.RoleLeader {
		orgID = userInfo.OrganizationID
	}

	userDB := database.GetDB().Table("sys_user su").
		Select("su.id, su.real_name, su.organization_id").
		Where("su.role_id = ?", constants.RoleMediator).
		Where("su.status = 1")

	if orgID > 0 {
		userDB = userDB.Where("su.organization_id = ?", orgID)
	}

	var users []map[string]interface{}
	userDB.Find(&users)

	count := 0
	for _, user := range users {
		userID, _ := user["id"].(int64)

		internalReq := &struct {
			UserID int64
			Period int32
			Year   int32
			Month  int32
		}{
			UserID: userID,
			Period: constants.PerformancePeriodMonth,
			Year:   int32(year),
			Month:  int32(month),
		}

		calculateSingleMediatorScore(internalReq, userInfo)
		count++
	}

	return count
}

func calculateSingleMediatorScore(req *struct {
	UserID int64
	Period int32
	Year   int32
	Month  int32
}, userInfo *middleware.UserClaims) error {
	var user struct {
		RealName       string `gorm:"column:real_name"`
		OrganizationID int64  `gorm:"column:organization_id"`
	}
	database.GetDB().Table("sys_user").
		Select("real_name, organization_id").
		Where("id = ?", req.UserID).
		First(&user)

	var startDate, endDate time.Time
	year := int(req.Year)
	month := int(req.Month)
	startDate = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate = startDate.AddDate(0, 1, 0)

	var totalCases int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Count(&totalCases)

	var closedCases int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Count(&closedCases)

	var successCases int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("mediation_result = ?", constants.MediationResultSuccess).
		Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Count(&successCases)

	var avgDays float64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Select("AVG(TIMESTAMPDIFF(DAY, created_at, closed_time))").
		Scan(&avgDays)

	var avgSatisfaction float64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", req.UserID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("closed_time >= ? AND closed_time < ?", startDate, endDate).
		Where("deleted_at IS NULL").
		Where("satisfaction_score > 0").
		Select("AVG(satisfaction_score)").
		Scan(&avgSatisfaction)

	var urgeCount int64
	database.GetDB().Table("workflow_urge").
		Where("current_handler_id = ?", req.UserID).
		Where("created_at >= ? AND created_at < ?", startDate, endDate).
		Count(&urgeCount)

	var indicators []map[string]interface{}
	database.GetDB().Table("performance_indicator_config").
		Where("status = 1").
		Find(&indicators)

	indicatorScoreMap := make(map[string]float64)
	totalWeight := 0.0
	totalScore := 0.0

	for _, ind := range indicators {
		code, _ := ind["indicator_code"].(string)
		weight, _ := ind["weight"].(float64)
		if code == "" {
			continue
		}
		totalWeight += weight
		score := calculateIndicatorScore(code, totalCases, closedCases, successCases, avgDays, avgSatisfaction, urgeCount)
		indicatorScoreMap[code] = score
		totalScore += score * weight
	}

	if totalWeight > 0 {
		totalScore = totalScore / totalWeight
	}
	totalScore = math.Round(totalScore*100) / 100

	level := "C"
	if totalScore >= 90 {
		level = "S"
	} else if totalScore >= 80 {
		level = "A"
	} else if totalScore >= 70 {
		level = "B"
	} else if totalScore >= 60 {
		level = "C"
	} else {
		level = "D"
	}

	closeRate := 0.0
	if totalCases > 0 {
		closeRate = math.Round(float64(closedCases)/float64(totalCases)*10000) / 100
	}
	successRate := 0.0
	if closedCases > 0 {
		successRate = math.Round(float64(successCases)/float64(closedCases)*10000) / 100
	}

	scoreID := utils.GenerateID()
	scoreData := map[string]interface{}{
		"id":                scoreID,
		"user_id":           req.UserID,
		"user_name":         user.RealName,
		"period":            req.Period,
		"year":              req.Year,
		"month":             req.Month,
		"quarter":           0,
		"start_date":        startDate,
		"end_date":          endDate,
		"case_count":        totalCases,
		"closed_count":      closedCases,
		"success_count":     successCases,
		"close_rate":        closeRate,
		"success_rate":      successRate,
		"avg_days":          math.Round(avgDays*100) / 100,
		"satisfaction":      math.Round(avgSatisfaction*100) / 100,
		"urge_count":        urgeCount,
		"total_score":       totalScore,
		"level":             level,
		"calculated_by":     userInfo.UserID,
		"calculated_by_name": userInfo.RealName,
		"organization_id":   user.OrganizationID,
	}

	for code, sc := range indicatorScoreMap {
		scoreData[code+"_score"] = math.Round(sc*100) / 100
	}

	tx := database.GetDB().Begin()
	tx.Table("performance_score").Where("user_id = ? AND year = ? AND month = ? AND period = ?",
		req.UserID, req.Year, req.Month, req.Period).Delete(nil)
	if err := tx.Table("performance_score").Create(scoreData).Error; err != nil {
		tx.Rollback()
		logger.Error("Create performance score failed", logger.Error(err))
		return err
	}

	snapshotData := map[string]interface{}{
		"user_id":          req.UserID,
		"user_name":        user.RealName,
		"org_id":           user.OrganizationID,
		"year":             req.Year,
		"month":            req.Month,
		"case_count":       totalCases,
		"closed_count":     closedCases,
		"close_rate":       closeRate,
		"success_count":    successCases,
		"success_rate":     successRate,
		"avg_days":         math.Round(avgDays*100) / 100,
		"avg_satisfaction": math.Round(avgSatisfaction*100) / 100,
		"urge_count":       urgeCount,
		"total_score":      totalScore,
		"level":            level,
	}
	var orgName string
	database.GetDB().Table("sys_organization").Select("org_name").Where("id = ?", user.OrganizationID).Scan(&orgName)
	snapshotData["org_name"] = orgName
	tx.Table("performance_monthly_snapshot").Where("user_id = ? AND year = ? AND month = ?", req.UserID, req.Year, req.Month).Delete(nil)
	tx.Table("performance_monthly_snapshot").Create(snapshotData)

	tx.Commit()
	return nil
}

func CreatePerformanceInterview(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ScoreID         int64  `json:"scoreId"`
		UserID          int64  `json:"userId" binding:"required"`
		UserName        string `json:"userName"`
		PeriodType      int32  `json:"periodType" binding:"required"`
		PeriodValue     string `json:"periodValue" binding:"required"`
		TotalScore      float64 `json:"totalScore"`
		Level           string `json:"level"`
		InterviewTime   string `json:"interviewTime" binding:"required"`
		InterviewPlace  string `json:"interviewPlace"`
		InterviewType   int32  `json:"interviewType" binding:"required"`
		Strengths       string `json:"strengths"`
		Weaknesses      string `json:"weaknesses"`
		ImprovementPlan string `json:"improvementPlan"`
		TargetNextPeriod string `json:"targetNextPeriod"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限创建绩效面谈"))
		return
	}

	var userName string
	if req.UserName == "" {
		database.GetDB().Table("sys_user").Select("real_name").Where("id = ?", req.UserID).Scan(&userName)
	} else {
		userName = req.UserName
	}

	var orgID int64
	database.GetDB().Table("sys_user").Select("organization_id").Where("id = ?", req.UserID).Scan(&orgID)

	interviewNo := fmt.Sprintf("IV%s", time.Now().Format("20060102150405"))
	interviewID := utils.GenerateID()

	interviewData := map[string]interface{}{
		"id":               interviewID,
		"interview_no":     interviewNo,
		"score_id":         req.ScoreID,
		"user_id":          req.UserID,
		"user_name":        userName,
		"org_id":           orgID,
		"period_type":      req.PeriodType,
		"period_value":     req.PeriodValue,
		"total_score":      req.TotalScore,
		"level":            req.Level,
		"interviewer_id":   userInfo.UserID,
		"interviewer_name": userInfo.RealName,
		"interview_time":   req.InterviewTime,
		"interview_place":  req.InterviewPlace,
		"interview_type":   req.InterviewType,
		"strengths":        req.Strengths,
		"weaknesses":       req.Weaknesses,
		"improvement_plan": req.ImprovementPlan,
		"target_next_period": req.TargetNextPeriod,
		"status":           1,
	}

	if err := database.GetDB().Table("performance_interview").Create(interviewData).Error; err != nil {
		logger.Error("Create performance interview failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("创建面谈记录失败"))
		return
	}

	sendInterviewNotification(req, interviewID, userName, userInfo)

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id":           interviewID,
		"interviewNo":  interviewNo,
	}, "绩效面谈记录创建成功"))
}

func sendInterviewNotification(req *struct {
	ScoreID         int64
	UserID          int64
	UserName        string
	PeriodType      int32
	PeriodValue     string
	TotalScore      float64
	Level           string
	InterviewTime   string
	InterviewPlace  string
	InterviewType   int32
	Strengths       string
	Weaknesses      string
	ImprovementPlan string
	TargetNextPeriod string
}, interviewID int64, userName string, userInfo *middleware.UserClaims) {
	var receiver struct {
		RealName string `gorm:"column:real_name"`
		Phone    string `gorm:"column:phone"`
	}
	database.GetDB().Table("sys_user").
		Select("real_name, phone").
		Where("id = ?", req.UserID).
		First(&receiver)

	interviewTypeMap := map[int32]string{
		1: "绩效反馈", 2: "改进计划", 3: "表彰面谈", 4: "预警面谈",
	}
	interviewType := interviewTypeMap[req.InterviewType]

	title := "【绩效面谈】您有待确认的绩效面谈记录"
	content := fmt.Sprintf(`尊敬的%s您好：
您的%s绩效面谈已创建，请及时登录系统查看并确认。
面谈类型：%s
面谈时间：%s
面谈地点：%s
综合得分：%.1f分（%s）
面谈人：%s

请登录系统查看详细面谈内容，填写您的意见并确认。`,
		receiver.RealName, req.PeriodValue, interviewType, req.InterviewTime,
		req.InterviewPlace, req.TotalScore, req.Level, userInfo.RealName)

	now := time.Now()
	paramsJSON, _ := json.Marshal(map[string]interface{}{
		"mediatorName":   receiver.RealName,
		"periodValue":    req.PeriodValue,
		"interviewType":  interviewType,
		"interviewTime":  req.InterviewTime,
		"interviewPlace": req.InterviewPlace,
		"totalScore":     fmt.Sprintf("%.1f", req.TotalScore),
		"level":          req.Level,
		"interviewerName": userInfo.RealName,
		"interviewId":    interviewID,
	})

	notificationRecord := map[string]interface{}{
		"receiver_id":   req.UserID,
		"receiver_name": receiver.RealName,
		"template_code": "TPL_PERFORMANCE_INTERVIEW",
		"title":         title,
		"content":       content,
		"channel_type":  "app",
		"receiver_type": 2,
		"status":        1,
		"params":        string(paramsJSON),
		"send_time":     now,
		"sender_id":     userInfo.UserID,
		"sender_name":   userInfo.RealName,
		"msg_no":        fmt.Sprintf("MSG%s", time.Now().Format("20060102150405")),
		"biz_type":      6,
		"biz_id":        interviewID,
	}

	database.GetDB().Table("notification_record").Create(notificationRecord)

	go func(phone, name string) {
		msg := map[string]interface{}{
			"receiverId":   req.UserID,
			"receiverName": name,
			"phone":        phone,
			"title":        title,
			"content":      content,
			"notifyType":   "app,sms",
			"templateCode": "TPL_PERFORMANCE_INTERVIEW",
			"sentBy":       userInfo.RealName,
		}
		mq.SendMessage(constants.MQTopicNotification, msg)
	}(receiver.Phone, receiver.RealName)
}

func GetPerformanceInterviewList(ctx context.Context, c *app.RequestContext) {
	var req struct {
		common.BaseQuery
		UserID      int64  `form:"userId"`
		PeriodValue string `form:"periodValue"`
		InterviewType int32 `form:"interviewType"`
		Status      int32  `form:"status"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	db := database.GetDB().Table("performance_interview pi").
		Select("pi.*").
		Where("1=1")

	if userInfo.Role == constants.RoleMediator {
		db = db.Where("pi.user_id = ?", userInfo.UserID)
	} else if userInfo.Role == constants.RoleLeader {
		db = db.Where("pi.org_id = ?", userInfo.OrganizationID)
	}

	if req.UserID > 0 {
		db = db.Where("pi.user_id = ?", req.UserID)
	}
	if req.PeriodValue != "" {
		db = db.Where("pi.period_value = ?", req.PeriodValue)
	}
	if req.InterviewType > 0 {
		db = db.Where("pi.interview_type = ?", req.InterviewType)
	}
	if req.Status > 0 {
		db = db.Where("pi.status = ?", req.Status)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("pi.interview_time DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	interviewTypeMap := map[int]string{
		1: "绩效反馈", 2: "改进计划", 3: "表彰面谈", 4: "预警面谈",
	}
	statusMap := map[int]string{
		1: "待确认", 2: "已确认", 3: "已归档",
	}
	for _, item := range list {
		if v, ok := item["interview_type"].(int); ok {
			item["interview_type_name"] = interviewTypeMap[v]
		}
		if v, ok := item["status"].(int); ok {
			item["status_name"] = statusMap[v]
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetPerformanceInterviewDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var interview map[string]interface{}
	result := database.GetDB().Table("performance_interview").
		Where("id = ?", id).
		Find(&interview)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, response.NotFound("面谈记录不存在"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role == constants.RoleMediator {
		if uid, ok := interview["user_id"].(int64); !ok || uid != userInfo.UserID {
			c.JSON(http.StatusForbidden, response.Forbidden("无权限查看他人面谈记录"))
			return
		}
	}

	c.JSON(http.StatusOK, response.Success(interview))
}

func ConfirmPerformanceInterview(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleMediator {
		c.JSON(http.StatusForbidden, response.Forbidden("仅调解员可确认面谈记录"))
		return
	}

	var req struct {
		MediatorComment string `json:"mediatorComment"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	updates := map[string]interface{}{
		"status":          2,
		"confirmed_at":    time.Now(),
		"mediator_comment": req.MediatorComment,
	}
	result := database.GetDB().Table("performance_interview").
		Where("id = ? AND user_id = ? AND status = 1", id, userInfo.UserID).
		Updates(updates)

	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("面谈记录不存在或已确认"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "确认成功"))
}

func ExportPerformanceExcel(ctx context.Context, c *app.RequestContext) {
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	orgID, _ := strconv.ParseInt(c.DefaultQuery("organizationId"), 10, 64)

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role == constants.RoleLeader && orgID == 0 {
		orgID = userInfo.OrganizationID
	}

	db := database.GetDB().Table("performance_monthly_snapshot pms").
		Where("pms.year = ? AND pms.month = ?", year, month)

	if orgID > 0 {
		db = db.Where("pms.org_id = ?", orgID)
	}
	if userInfo.Role == constants.RoleMediator {
		db = db.Where("pms.user_id = ?", userInfo.UserID)
	}

	var snapshots []map[string]interface{}
	db.Order("pms.total_score DESC").Find(&snapshots)

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="performance_%d_%02d.csv"`, year, month))

	bom := []byte{0xEF, 0xBB, 0xBF}
	c.Write(bom)

	writer := csv.NewWriter(c)
	header := []string{"排名", "调解员", "组织", "受理数", "办结数", "办结率(%)", "成功数", "成功率(%)", "平均天数", "满意度", "被催办次数", "综合得分", "等级"}
	writer.Write(header)

	for i, s := range snapshots {
		row := []string{
			strconv.Itoa(i + 1),
			fmt.Sprintf("%v", s["user_name"]),
			fmt.Sprintf("%v", s["org_name"]),
			fmt.Sprintf("%v", s["case_count"]),
			fmt.Sprintf("%v", s["closed_count"]),
			fmt.Sprintf("%.1f", s["close_rate"]),
			fmt.Sprintf("%v", s["success_count"]),
			fmt.Sprintf("%.1f", s["success_rate"]),
			fmt.Sprintf("%.1f", s["avg_days"]),
			fmt.Sprintf("%.1f", s["avg_satisfaction"]),
			fmt.Sprintf("%v", s["urge_count"]),
			fmt.Sprintf("%.1f", s["total_score"]),
			fmt.Sprintf("%v", s["level"]),
		}
		writer.Write(row)
	}

	writer.Flush()
}
