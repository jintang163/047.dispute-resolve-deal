package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type PerformanceScoreRequest struct {
	UserID    int64   `json:"userId" binding:"required"`
	Period    int32   `json:"period" binding:"required"`
	Year      int32   `json:"year" binding:"required"`
	Month     int32   `json:"month"`
	Quarter   int32   `json:"quarter"`
	CaseCount float64 `json:"caseCount"`
	CloseRate float64 `json:"closeRate"`
	SuccessRate float64 `json:"successRate"`
	AvgDays   float64 `json:"avgDays"`
	Satisfaction float64 `json:"satisfaction"`
	Remark    string  `json:"remark"`
}

func GetPerformanceScoreList(ctx context.Context, c *app.RequestContext) {
	var req struct {
		common.BaseQuery
		Period         int32  `form:"period"`
		Year           int32  `form:"year"`
		UserID         int64  `form:"userId"`
		OrganizationID int64  `form:"organizationId"`
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
		UserID int64 `json:"userId" binding:"required"`
		Period int32 `json:"period" binding:"required"`
		Year   int32 `json:"year" binding:"required"`
		Month  int32 `json:"month"`
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

	var indicators []map[string]interface{}
	database.GetDB().Table("performance_indicator_config").
		Where("status = 1").
		Find(&indicators)

	caseCountScore := 0.0
	closeRateScore := 0.0
	successRateScore := 0.0
	avgDaysScore := 0.0
	satisfactionScore := 0.0

	for _, ind := range indicators {
		code := ind["indicator_code"].(string)
		weight := ind["weight"].(float64)
		fullScore := ind["full_score"].(float64)

		switch code {
		case "CASE_COUNT":
			target := 10.0
			if totalCases >= int64(target) {
				caseCountScore = fullScore
			} else {
				caseCountScore = float64(totalCases) / target * fullScore
			}
		case "CLOSE_RATE":
			if totalCases > 0 {
				closeRate := float64(closedCases) / float64(totalCases) * 100
				if closeRate >= 90 {
					closeRateScore = fullScore
				} else if closeRate >= 80 {
					closeRateScore = fullScore * 0.8
				} else if closeRate >= 70 {
					closeRateScore = fullScore * 0.6
				} else {
					closeRateScore = fullScore * 0.4
				}
			}
		case "SUCCESS_RATE":
			if closedCases > 0 {
				successRate := float64(successCases) / float64(closedCases) * 100
				if successRate >= 90 {
					successRateScore = fullScore
				} else if successRate >= 80 {
					successRateScore = fullScore * 0.8
				} else if successRate >= 70 {
					successRateScore = fullScore * 0.6
				} else {
					successRateScore = fullScore * 0.4
				}
			}
		case "AVG_DAYS":
			if avgDays > 0 {
				if avgDays <= 7 {
					avgDaysScore = fullScore
				} else if avgDays <= 15 {
					avgDaysScore = fullScore * 0.8
				} else if avgDays <= 30 {
					avgDaysScore = fullScore * 0.6
				} else {
					avgDaysScore = fullScore * 0.4
				}
			} else {
				avgDaysScore = fullScore
			}
		case "SATISFACTION":
			if avgSatisfaction >= 4.5 {
				satisfactionScore = fullScore
			} else if avgSatisfaction >= 4.0 {
				satisfactionScore = fullScore * 0.8
			} else if avgSatisfaction >= 3.5 {
				satisfactionScore = fullScore * 0.6
			} else {
				satisfactionScore = fullScore * 0.4
			}
		}
	}

	totalScore := caseCountScore + closeRateScore + successRateScore + avgDaysScore + satisfactionScore

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

	scoreID := utils.GenerateID()
	scoreData := map[string]interface{}{
		"id":              scoreID,
		"user_id":         req.UserID,
		"user_name":       user.RealName,
		"period":          req.Period,
		"year":            req.Year,
		"month":           req.Month,
		"quarter":         req.Quarter,
		"start_date":      startDate,
		"end_date":        endDate,
		"case_count":      totalCases,
		"closed_count":    closedCases,
		"success_count":   successCases,
		"close_rate":      fmt.Sprintf("%.1f%%", float64(closedCases)/float64(totalCases)*100),
		"success_rate":    fmt.Sprintf("%.1f%%", float64(successCases)/float64(closedCases)*100),
		"avg_days":        fmt.Sprintf("%.1f", avgDays),
		"satisfaction":    fmt.Sprintf("%.1f", avgSatisfaction),
		"case_count_score": caseCountScore,
		"close_rate_score": closeRateScore,
		"success_rate_score": successRateScore,
		"avg_days_score":   avgDaysScore,
		"satisfaction_score": satisfactionScore,
		"total_score":     totalScore,
		"level":           level,
		"calculated_by":   userInfo.UserID,
		"calculated_by_name": userInfo.RealName,
		"organization_id": user.OrganizationID,
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("performance_score").Create(scoreData).Error; err != nil {
		tx.Rollback()
		logger.Error("Create performance score failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("考核计算失败"))
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id":           scoreID,
		"totalScore":   totalScore,
		"level":        level,
		"caseCount":    totalCases,
		"closedCount":  closedCases,
		"successCount": successCases,
		"closeRate":    fmt.Sprintf("%.1f%%", float64(closedCases)/float64(totalCases)*100),
		"successRate":  fmt.Sprintf("%.1f%%", float64(successCases)/float64(closedCases)*100),
		"avgDays":      fmt.Sprintf("%.1f", avgDays),
		"satisfaction": fmt.Sprintf("%.1f", avgSatisfaction),
	}, "考核计算成功"))
}

func GetPerformanceTrend(ctx context.Context, c *app.RequestContext) {
	userID, _ := strconv.ParseInt(c.Query("userId"), 10, 64)
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))

	userInfo := middleware.GetUserInfo(c)
	if userID == 0 {
		userID = userInfo.UserID
	}

	var scores []map[string]interface{}
	database.GetDB().Table("performance_score").
		Select("month, total_score, level, case_count, closed_count").
		Where("user_id = ?", userID).
		Where("year = ?", year).
		Where("period = ?", constants.PerformancePeriodMonth).
		Order("month ASC").
		Find(&scores)

	monthData := make([]map[string]interface{}, 12)
	for i := 0; i < 12; i++ {
		monthData[i] = map[string]interface{}{
			"month":      i + 1,
			"totalScore": 0,
			"level":      "-",
			"caseCount":  0,
			"hasData":    false,
		}
	}

	for _, s := range scores {
		month := int(s["month"].(int32)) - 1
		if month >= 0 && month < 12 {
			monthData[month]["totalScore"] = s["total_score"]
			monthData[month]["level"] = s["level"]
			monthData[month]["caseCount"] = s["case_count"]
			monthData[month]["hasData"] = true
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

	c.JSON(http.StatusOK, response.Success(configs))
}
