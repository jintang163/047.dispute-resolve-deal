package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type MediatorScore struct {
	UserID      int64   `json:"userId"`
	RealName    string  `json:"realName"`
	TotalScore  float64 `json:"totalScore"`
	LoadScore   float64 `json:"loadScore"`
	SpecialtyScore float64 `json:"specialtyScore"`
	LocationScore  float64 `json:"locationScore"`
	SuccessScore   float64 `json:"successScore"`
	UrgencyScore   float64 `json:"urgencyScore"`
}

type DispatchRequest struct {
	CaseID    int64   `json:"caseId" binding:"required"`
	Algorithm string  `json:"algorithm"`
	ExcludeIDs []int64 `json:"excludeIds"`
	AutoAssign bool    `json:"autoAssign"`
}

type AutoDispatchConfig struct {
	LoadWeight      float64 `json:"loadWeight"`
	SpecialtyWeight float64 `json:"specialtyWeight"`
	LocationWeight  float64 `json:"locationWeight"`
	SuccessWeight   float64 `json:"successWeight"`
	UrgencyWeight   float64 `json:"urgencyWeight"`
	Enabled         bool    `json:"enabled"`
}

var defaultDispatchConfig = AutoDispatchConfig{
	LoadWeight:      0.25,
	SpecialtyWeight: 0.30,
	LocationWeight:  0.15,
	SuccessWeight:   0.20,
	UrgencyWeight:   0.10,
	Enabled:         true,
}

func IntelligentDispatch(ctx context.Context, c *app.RequestContext) {
	var req DispatchRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限进行智能分派"))
		return
	}

	var caseData struct {
		ID              int64   `gorm:"column:id"`
		CaseNo          string  `gorm:"column:case_no"`
		Title           string  `gorm:"column:title"`
		TypeID          int64   `gorm:"column:type_id"`
		CaseLevel       int32   `gorm:"column:case_level"`
		Longitude       float64 `gorm:"column:longitude"`
		Latitude        float64 `gorm:"column:latitude"`
		Status          int32   `gorm:"column:status"`
		OrganizationID  int64   `gorm:"column:organization_id"`
	}

	result := database.GetDB().Table("dispute_case").
		Where("id = ?", req.CaseID).
		First(&caseData)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("案件不存在"))
		return
	}

	if caseData.Status != constants.CaseStatusPending {
		c.JSON(http.StatusBadRequest, response.BadRequest("只有待分派状态的案件才能分派"))
		return
	}

	config := getDispatchConfig(caseData.OrganizationID)

	mediators, err := calculateMediatorScores(&caseData, &config, req.ExcludeIDs)
	if err != nil {
		logger.Error("Calculate mediator scores failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("计算调解员分数失败"))
		return
	}

	if len(mediators) == 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("没有可用的调解员"))
		return
	}

	if req.AutoAssign {
		bestMediator := mediators[0]
		err := assignCaseToMediator(&caseData, bestMediator.UserID, userInfo, "智能分派")
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ServerError("自动分派失败"))
			return
		}

		c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
			"assignedMediator": bestMediator,
			"candidates":       mediators,
			"config":           config,
		}, "智能分派成功"))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"candidates": mediators,
		"config":     config,
	}))
}

func calculateMediatorScores(caseData *struct {
	ID              int64
	CaseNo          string
	Title           string
	TypeID          int64
	CaseLevel       int32
	Longitude       float64
	Latitude        float64
	Status          int32
	OrganizationID  int64
}, config *AutoDispatchConfig, excludeIDs []int64) ([]MediatorScore, error) {

	var orgs []int64
	database.GetDB().Table("sys_organization").
		Select("id").
		Where("parent_id = ? OR id = ?", caseData.OrganizationID, caseData.OrganizationID).
		Pluck("id", &orgs)

	var mediators []map[string]interface{}
	db := database.GetDB().Table("sys_user su").
		Select("su.id, su.real_name, su.phone, su.specialty, su.organization_id, so.org_name, so.longitude as org_longitude, so.latitude as org_latitude").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("su.role = ?", constants.RoleMediator).
		Where("su.status = 1").
		Where("su.deleted_at IS NULL").
		Where("su.organization_id IN ?", orgs)

	if len(excludeIDs) > 0 {
		db = db.Where("su.id NOT IN ?", excludeIDs)
	}

	db.Find(&mediators)

	if len(mediators) == 0 {
		return nil, fmt.Errorf("no available mediators")
	}

	var typePath string
	var typeName string
	database.GetDB().Table("dispute_type").
		Select("level_path, type_name").
		Where("id = ?", caseData.TypeID).
		Row().Scan(&typePath, &typeName)

	var scores []MediatorScore

	for _, m := range mediators {
		userID := m["id"].(int64)
		score := MediatorScore{
			UserID:   userID,
			RealName: m["real_name"].(string),
		}

		loadScore := calculateLoadScore(userID)
		specialtyScore := calculateSpecialtyScore(m["specialty"], typePath, typeName)
		locationScore := calculateLocationScore(caseData.Longitude, caseData.Latitude, m)
		successScore := calculateSuccessScore(userID)
		urgencyScore := calculateUrgencyScore(caseData.CaseLevel)

		totalScore := loadScore*config.LoadWeight +
			specialtyScore*config.SpecialtyWeight +
			locationScore*config.LocationWeight +
			successScore*config.SuccessWeight +
			urgencyScore*config.UrgencyWeight

		score.LoadScore = loadScore
		score.SpecialtyScore = specialtyScore
		score.LocationScore = locationScore
		score.SuccessScore = successScore
		score.UrgencyScore = urgencyScore
		score.TotalScore = math.Round(totalScore*100) / 100

		scores = append(scores, score)
	}

	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].TotalScore > scores[i].TotalScore {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	return scores, nil
}

func calculateLoadScore(userID int64) float64 {
	var currentCases int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", userID).
		Where("status IN (?)", []int32{constants.CaseStatusMediating, constants.CaseStatusWaiting, constants.CaseStatusApproving}).
		Where("deleted_at IS NULL").
		Count(&currentCases)

	const optimalLoad = 5
	const maxLoad = 15

	if currentCases >= maxLoad {
		return 0
	}

	loadFactor := 1.0 - float64(currentCases)/float64(maxLoad)
	return math.Round(loadFactor*100) / 100
}

func calculateSpecialtyScore(specialty interface{}, typePath, typeName string) float64 {
	if specialty == nil {
		return 0.5
	}

	specialtyStr, ok := specialty.(string)
	if !ok || specialtyStr == "" {
		return 0.5
	}

	specialtyLower := strings.ToLower(specialtyStr)
	typePathLower := strings.ToLower(typePath)
	typeNameLower := strings.ToLower(typeName)

	if strings.Contains(specialtyLower, typeNameLower) {
		return 1.0
	}

	if strings.Contains(typePathLower, specialtyLower) {
		return 0.9
	}

	keywords := []string{"婚姻", "家庭", "邻里", "财产", "债务", "劳动", "消费", "医疗", "交通", "物业"}
	for _, kw := range keywords {
		if strings.Contains(specialtyLower, kw) && strings.Contains(typeNameLower, kw) {
			return 0.85
		}
	}

	if strings.Contains(specialtyLower, "纠纷") || strings.Contains(specialtyLower, "调解") {
		return 0.7
	}

	return 0.5
}

func calculateLocationScore(caseLng, caseLat float64, mediator map[string]interface{}) float64 {
	mediatorLng, _ := mediator["org_longitude"].(float64)
	mediatorLat, _ := mediator["org_latitude"].(float64)

	if caseLng == 0 || caseLat == 0 || mediatorLng == 0 || mediatorLat == 0 {
		return 0.5
	}

	distance := calculateHaversineDistance(caseLng, caseLat, mediatorLng, mediatorLat)

	const maxDistance = 20.0
	if distance >= maxDistance {
		return 0.1
	}

	score := 1.0 - (distance / maxDistance)
	return math.Round(score*100) / 100
}

func calculateHaversineDistance(lng1, lat1, lng2, lat2 float64) float64 {
	const earthRadius = 6371.0

	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLng := (lng2 - lng1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLng/2)*math.Sin(deltaLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

func calculateSuccessScore(userID int64) float64 {
	var totalClosed, successCases int64

	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", userID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("deleted_at IS NULL").
		Count(&totalClosed)

	if totalClosed == 0 {
		return 0.7
	}

	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", userID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("mediation_result = ?", constants.MediationResultSuccess).
		Where("deleted_at IS NULL").
		Count(&successCases)

	successRate := float64(successCases) / float64(totalClosed)

	var avgSatisfaction float64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ?", userID).
		Where("status = ?", constants.CaseStatusClosed).
		Where("satisfaction_score > 0").
		Where("deleted_at IS NULL").
		Select("AVG(satisfaction_score)").
		Scan(&avgSatisfaction)

	satisfactionScore := 0.7
	if avgSatisfaction > 0 {
		satisfactionScore = avgSatisfaction / 5.0
	}

	totalScore := (successRate*0.6 + satisfactionScore*0.4)
	return math.Round(totalScore*100) / 100
}

func calculateUrgencyScore(caseLevel int32) float64 {
	switch caseLevel {
	case constants.CaseLevelExtraUrgent:
		return 1.0
	case constants.CaseLevelUrgent:
		return 0.8
	case constants.CaseLevelNormal:
		return 0.6
	case constants.CaseLevelCommon:
		return 0.4
	default:
		return 0.5
	}
}

func assignCaseToMediator(caseData *struct {
	ID             int64
	CaseNo         string
	Title          string
}, mediatorID int64, userInfo *auth.UserInfo, reason string) error {
	var mediator struct {
		RealName       string `gorm:"column:real_name"`
		Phone          string `gorm:"column:phone"`
		OrganizationID int64  `gorm:"column:organization_id"`
	}

	result := database.GetDB().Table("sys_user").
		Where("id = ? AND status = 1", mediatorID).
		First(&mediator)

	if result.Error != nil {
		return fmt.Errorf("mediator not found")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"mediator_id":        mediatorID,
		"mediator_name":      mediator.RealName,
		"mediator_time":      now,
		"status":             constants.CaseStatusMediating,
		"mediation_start_time": now,
	}

	tx := database.GetDB().Begin()

	if err := tx.Table("dispute_case").Where("id = ?", caseData.ID).Updates(updates).Error; err != nil {
		tx.Rollback()
		return err
	}

	history := map[string]interface{}{
		"case_id":          caseData.ID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "ASSIGN",
		"operation_detail": fmt.Sprintf("分派给调解员: %s，原因: %s", mediator.RealName, reason),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
		"operator_role":    userInfo.Role,
		"old_status":       constants.CaseStatusPending,
		"new_status":       constants.CaseStatusMediating,
		"remark":           reason,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseData.ID)
	cache.Del(context.Background(), cacheKey)

	go func() {
		msg := map[string]interface{}{
			"caseId":       caseData.ID,
			"caseNo":       caseData.CaseNo,
			"title":        caseData.Title,
			"mediatorId":   mediatorID,
			"mediatorName": mediator.RealName,
			"mediatorPhone": mediator.Phone,
			"assignBy":     userInfo.UserID,
			"assignByName": userInfo.RealName,
			"assignTime":   now.Format(time.RFC3339),
			"isAuto":       true,
		}
		mq.SendMessage(constants.MQTopicCaseAssign, msg)
	}()

	return nil
}

func GetDispatchConfig(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	config := getDispatchConfig(userInfo.OrganizationID)
	c.JSON(http.StatusOK, response.Success(config))
}

func UpdateDispatchConfig(ctx context.Context, c *app.RequestContext) {
	var config AutoDispatchConfig
	if err := c.BindAndValidate(&config); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限修改分派配置"))
		return
	}

	cacheKey := fmt.Sprintf("dispatch:config:%d", userInfo.OrganizationID)
	configJSON, _ := json.Marshal(config)
	cache.Set(ctx, cacheKey, string(configJSON), 24*time.Hour)

	c.JSON(http.StatusOK, response.SuccessWithMessage(config, "配置更新成功"))
}

func getDispatchConfig(orgID int64) AutoDispatchConfig {
	cacheKey := fmt.Sprintf("dispatch:config:%d", orgID)
	cachedData, err := cache.Get(context.Background(), cacheKey)
	if err == nil && cachedData != "" {
		var config AutoDispatchConfig
		if json.Unmarshal([]byte(cachedData), &config) == nil {
			return config
		}
	}
	return defaultDispatchConfig
}

func GetMediatorLoadStats(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	var orgs []int64
	database.GetDB().Table("sys_organization").
		Select("id").
		Where("parent_id = ? OR id = ?", userInfo.OrganizationID, userInfo.OrganizationID).
		Pluck("id", &orgs)

	var mediators []map[string]interface{}
	database.GetDB().Table("sys_user su").
		Select("su.id, su.real_name, so.org_name").
		Joins("LEFT JOIN sys_organization so ON su.organization_id = so.id").
		Where("su.role = ?", constants.RoleMediator).
		Where("su.status = 1").
		Where("su.deleted_at IS NULL").
		Where("su.organization_id IN ?", orgs).
		Find(&mediators)

	var stats []map[string]interface{}
	for _, m := range mediators {
		userID := m["id"].(int64)

		var pendingCount, mediatingCount, closedCount, totalCount int64

		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("deleted_at IS NULL").
			Count(&totalCount)

		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("status = ?", constants.CaseStatusMediating).
			Where("deleted_at IS NULL").
			Count(&mediatingCount)

		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("status IN (?)", []int32{constants.CaseStatusWaiting, constants.CaseStatusApproving}).
			Where("deleted_at IS NULL").
			Count(&pendingCount)

		database.GetDB().Table("dispute_case").
			Where("mediator_id = ?", userID).
			Where("status = ?", constants.CaseStatusClosed).
			Where("deleted_at IS NULL").
			Count(&closedCount)

		loadLevel := "正常"
		currentLoad := mediatingCount + pendingCount
		if currentLoad >= 10 {
			loadLevel = "繁忙"
		} else if currentLoad >= 5 {
			loadLevel = "适中"
		}

		stats = append(stats, map[string]interface{}{
			"userId":         userID,
			"realName":       m["real_name"],
			"orgName":        m["org_name"],
			"totalCount":     totalCount,
			"mediatingCount": mediatingCount,
			"pendingCount":   pendingCount,
			"closedCount":    closedCount,
			"currentLoad":    currentLoad,
			"loadLevel":      loadLevel,
			"loadScore":      calculateLoadScore(userID),
		})
	}

	c.JSON(http.StatusOK, response.Success(stats))
}

func BatchIntelligentDispatch(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限进行批量智能分派"))
		return
	}

	var req struct {
		CaseIDs []int64 `json:"caseIds" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	results := make(map[int64]interface{})
	successCount := 0
	failCount := 0

	for _, caseID := range req.CaseIDs {
		var caseData struct {
			ID              int64   `gorm:"column:id"`
			CaseNo          string  `gorm:"column:case_no"`
			Title           string  `gorm:"column:title"`
			TypeID          int64   `gorm:"column:type_id"`
			CaseLevel       int32   `gorm:"column:case_level"`
			Longitude       float64 `gorm:"column:longitude"`
			Latitude        float64 `gorm:"column:latitude"`
			Status          int32   `gorm:"column:status"`
			OrganizationID  int64   `gorm:"column:organization_id"`
		}

		database.GetDB().Table("dispute_case").
			Where("id = ?", caseID).
			First(&caseData)

		if caseData.Status != constants.CaseStatusPending {
			results[caseID] = map[string]interface{}{
				"success": false,
				"message": "案件状态不是待分派",
			}
			failCount++
			continue
		}

		config := getDispatchConfig(caseData.OrganizationID)
		mediators, err := calculateMediatorScores(&caseData, &config, nil)

		if err != nil || len(mediators) == 0 {
			results[caseID] = map[string]interface{}{
				"success": false,
				"message": "没有可用的调解员",
			}
			failCount++
			continue
		}

		bestMediator := mediators[0]
		err = assignCaseToMediator(&caseData, bestMediator.UserID, userInfo, "批量智能分派")

		if err != nil {
			results[caseID] = map[string]interface{}{
				"success": false,
				"message": err.Error(),
			}
			failCount++
			continue
		}

		results[caseID] = map[string]interface{}{
			"success":    true,
			"mediatorId": bestMediator.UserID,
			"mediatorName": bestMediator.RealName,
			"score":      bestMediator.TotalScore,
		}
		successCount++
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"successCount": successCount,
		"failCount":    failCount,
		"results":      results,
	}, fmt.Sprintf("批量分派完成，成功%d个，失败%d个", successCount, failCount)))
}
