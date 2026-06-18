package impl

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type DispatchServiceImpl struct{}

func NewDispatchService() service.DispatchService {
	return &DispatchServiceImpl{}
}

type DispatchConfig struct {
	LoadWeight      float64 `json:"loadWeight"`
	SpecialtyWeight float64 `json:"specialtyWeight"`
	LocationWeight  float64 `json:"locationWeight"`
	SuccessWeight   float64 `json:"successWeight"`
	UrgencyWeight   float64 `json:"urgencyWeight"`
	AutoAssign      bool    `json:"autoAssign"`
	MaxLoadPerDay   int     `json:"maxLoadPerDay"`
}

var defaultDispatchConfig = DispatchConfig{
	LoadWeight:      0.25,
	SpecialtyWeight: 0.30,
	LocationWeight:  0.15,
	SuccessWeight:   0.20,
	UrgencyWeight:   0.10,
	AutoAssign:      false,
	MaxLoadPerDay:   5,
}

type MediatorCandidate struct {
	UserID          int64   `json:"userId"`
	RealName        string  `json:"realName"`
	Avatar          string  `json:"avatar"`
	Role            int     `json:"role"`
	OrgName         string  `json:"orgName"`
	Specialties     string  `json:"specialties"`
	Longitude       float64 `json:"longitude"`
	Latitude        float64 `json:"latitude"`
	CurrentLoad     int     `json:"currentLoad"`
	TodayAssigned   int     `json:"todayAssigned"`
	SuccessRate     float64 `json:"successRate"`
	TotalCases      int     `json:"totalCases"`
	LoadScore       float64 `json:"loadScore"`
	SpecialtyScore  float64 `json:"specialtyScore"`
	LocationScore   float64 `json:"locationScore"`
	SuccessScore    float64 `json:"successScore"`
	UrgencyScore    float64 `json:"urgencyScore"`
	TotalScore      float64 `json:"totalScore"`
	Distance        float64 `json:"distance"`
}

func (s *DispatchServiceImpl) IntelligentDispatch(ctx context.Context, caseID int64, algorithm string, excludeIDs []int64, autoAssign bool, orgID int64) (map[string]interface{}, error) {
	caseInfo, err := s.getCaseInfo(caseID)
	if err != nil {
		return nil, err
	}

	if caseInfo["status"].(int) != constants.CaseStatusPending {
		return nil, nil
	}

	candidates, err := s.GetDispatchCandidates(ctx, caseID, orgID, 10)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return map[string]interface{}{
			"caseId":     caseID,
			"candidates": []interface{}{},
			"recommend":  nil,
			"autoAssign": false,
			"message":    "暂无符合条件的调解员",
		}, nil
	}

	excludeIDMap := make(map[int64]bool)
	for _, id := range excludeIDs {
		excludeIDMap[id] = true
	}

	filteredCandidates := make([]map[string]interface{}, 0)
	for _, c := range candidates {
		uid := c["userId"].(int64)
		if !excludeIDMap[uid] {
			filteredCandidates = append(filteredCandidates, c)
		}
	}

	if len(filteredCandidates) == 0 {
		filteredCandidates = candidates
	}

	recommend := filteredCandidates[0]
	result := map[string]interface{}{
		"caseId":     caseID,
		"caseInfo":   caseInfo,
		"candidates": filteredCandidates,
		"recommend":  recommend,
		"algorithm":  algorithm,
		"autoAssign": autoAssign,
	}

	if autoAssign {
		recommendUserID := recommend["userId"].(int64)
		err = s.assignCase(caseID, recommendUserID, 0)
		if err != nil {
			return nil, err
		}
		result["assigned"] = true
		result["assignedTo"] = recommendUserID
	}

	return result, nil
}

func (s *DispatchServiceImpl) GetDispatchCandidates(ctx context.Context, caseID int64, orgID int64, topN int) ([]map[string]interface{}, error) {
	caseInfo, err := s.getCaseInfo(caseID)
	if err != nil {
		return nil, err
	}

	mediators, err := s.getAvailableMediators(orgID)
	if err != nil {
		return nil, err
	}

	if len(mediators) == 0 {
		return []map[string]interface{}{}, nil
	}

	config, err := s.getDispatchConfigInternal(orgID)
	if err != nil {
		config = defaultDispatchConfig
	}

	caseTypeID, _ := caseInfo["dispute_type_id"].(int64)
	caseLevel, _ := caseInfo["level"].(int)
	longitude, _ := caseInfo["longitude"].(float64)
	latitude, _ := caseInfo["latitude"].(float64)

	caseTypePath := s.getDisputeTypePath(caseTypeID)

	candidates := make([]*MediatorCandidate, 0, len(mediators))
	for _, med := range mediators {
		candidate := s.calculateMediatorScore(med, caseTypePath, caseLevel, longitude, latitude, config)
		candidates = append(candidates, candidate)
	}

	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].TotalScore > candidates[i].TotalScore {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	if topN > 0 && topN < len(candidates) {
		candidates = candidates[:topN]
	}

	result := make([]map[string]interface{}, 0, len(candidates))
	for rank, c := range candidates {
		result = append(result, map[string]interface{}{
			"userId":         c.UserID,
			"realName":       c.RealName,
			"avatar":         c.Avatar,
			"role":           c.Role,
			"roleText":       constants.RoleMap[c.Role],
			"orgName":        c.OrgName,
			"specialties":    c.Specialties,
			"currentLoad":    c.CurrentLoad,
			"todayAssigned":  c.TodayAssigned,
			"successRate":    c.SuccessRate,
			"totalCases":     c.TotalCases,
			"distance":       c.Distance,
			"loadScore":      c.LoadScore,
			"specialtyScore": c.SpecialtyScore,
			"locationScore":  c.LocationScore,
			"successScore":   c.SuccessScore,
			"urgencyScore":   c.UrgencyScore,
			"totalScore":     c.TotalScore,
			"rank":           rank + 1,
		})
	}

	return result, nil
}

func (s *DispatchServiceImpl) GetDispatchConfig(ctx context.Context, orgID int64) (map[string]interface{}, error) {
	config, err := s.getDispatchConfigInternal(orgID)
	if err != nil {
		config = defaultDispatchConfig
	}

	return map[string]interface{}{
		"orgId":           orgID,
		"loadWeight":      config.LoadWeight,
		"specialtyWeight": config.SpecialtyWeight,
		"locationWeight":  config.LocationWeight,
		"successWeight":   config.SuccessWeight,
		"urgencyWeight":   config.UrgencyWeight,
		"autoAssign":      config.AutoAssign,
		"maxLoadPerDay":   config.MaxLoadPerDay,
	}, nil
}

func (s *DispatchServiceImpl) UpdateDispatchConfig(ctx context.Context, orgID int64, config map[string]interface{}) error {
	configJSON, err := utils.ToJSON(config)
	if err != nil {
		return err
	}

	var existing model.DispatchConfig
	database.GetDB().Where("org_id = ?", orgID).First(&existing)

	if existing.ID == 0 {
		existing = model.DispatchConfig{
			OrgID:         orgID,
			ConfigContent: configJSON,
			Status:        1,
		}
		return database.GetDB().Create(&existing).Error
	}

	return database.GetDB().Model(&existing).Update("config_content", configJSON).Error
}

func (s *DispatchServiceImpl) GetMediatorLoadStats(ctx context.Context, orgID int64) ([]map[string]interface{}, error) {
	mediators, err := s.getAvailableMediators(orgID)
	if err != nil {
		return nil, err
	}

	result := make([]map[string]interface{}, 0, len(mediators))
	for _, med := range mediators {
		userID := med["id"].(int64)
		currentLoad := s.getCurrentLoad(userID)
		todayAssigned := s.getTodayAssigned(userID)

		result = append(result, map[string]interface{}{
			"userId":        userID,
			"realName":      med["real_name"],
			"orgName":       med["org_name"],
			"currentLoad":   currentLoad,
			"todayAssigned": todayAssigned,
			"maxLoad":       defaultDispatchConfig.MaxLoadPerDay,
			"loadPercent":   float64(todayAssigned) / float64(defaultDispatchConfig.MaxLoadPerDay) * 100,
		})
	}

	return result, nil
}

func (s *DispatchServiceImpl) BatchIntelligentDispatch(ctx context.Context, caseIDs []int64, orgID int64, operatorID int64) (int, int, error) {
	successCount := 0
	failCount := 0

	for _, caseID := range caseIDs {
		result, err := s.IntelligentDispatch(ctx, caseID, "weighted", []int64{}, true, orgID)
		if err != nil {
			failCount++
			logger.Error("Batch dispatch error", logger.Int64("caseId", caseID), logger.Error(err))
			continue
		}

		if result["assigned"] == true {
			successCount++
		} else {
			failCount++
		}
	}

	return successCount, failCount, nil
}

func (s *DispatchServiceImpl) getCaseInfo(caseID int64) (map[string]interface{}, error) {
	var caseInfo map[string]interface{}
	result := database.GetDB().Table("dispute_case").
		Where("id = ? AND deleted_at IS NULL", caseID).
		First(&caseInfo)
	return caseInfo, result.Error
}

func (s *DispatchServiceImpl) getAvailableMediators(orgID int64) ([]map[string]interface{}, error) {
	db := database.GetDB().Table("user u").
		Select("u.id, u.username, u.real_name, u.avatar, u.role, u.specialties, u.longitude, u.latitude, o.org_name, o.org_path").
		Joins("LEFT JOIN organization o ON u.org_id = o.id").
		Where("u.role = ? AND u.status = 1 AND u.deleted_at IS NULL", constants.RoleMediator)

	if orgID > 0 {
		db = db.Where("u.org_id = ?", orgID)
	}

	var mediators []map[string]interface{}
	result := db.Find(&mediators)
	return mediators, result.Error
}

func (s *DispatchServiceImpl) calculateMediatorScore(med map[string]interface{}, caseTypePath string, caseLevel int, caseLng, caseLat float64, config DispatchConfig) *MediatorCandidate {
	userID := med["id"].(int64)
	specialties, _ := med["specialties"].(string)
	longitude, _ := med["longitude"].(float64)
	latitude, _ := med["latitude"].(float64)

	currentLoad := s.getCurrentLoad(userID)
	todayAssigned := s.getTodayAssigned(userID)
	successRate := s.getSuccessRate(userID)
	totalCases := s.getTotalCases(userID)

	loadScore := s.calculateLoadScore(currentLoad, todayAssigned, config.MaxLoadPerDay)
	specialtyScore := s.calculateSpecialtyScore(specialties, caseTypePath)
	locationScore := s.calculateLocationScore(longitude, latitude, caseLng, caseLat)
	successScore := s.calculateSuccessScore(successRate, totalCases)
	urgencyScore := s.calculateUrgencyScore(caseLevel, currentLoad)

	totalScore := loadScore*config.LoadWeight +
		specialtyScore*config.SpecialtyWeight +
		locationScore*config.LocationWeight +
		successScore*config.SuccessWeight +
		urgencyScore*config.UrgencyWeight

	distance := 0.0
	if caseLng != 0 && caseLat != 0 && longitude != 0 && latitude != 0 {
		distance = calculateHaversineDistance(longitude, latitude, caseLng, caseLat)
	}

	realName, _ := med["real_name"].(string)
	avatar, _ := med["avatar"].(string)
	role, _ := med["role"].(int)
	orgName, _ := med["org_name"].(string)

	return &MediatorCandidate{
		UserID:         userID,
		RealName:       realName,
		Avatar:         avatar,
		Role:           role,
		OrgName:        orgName,
		Specialties:    specialties,
		Longitude:      longitude,
		Latitude:       latitude,
		CurrentLoad:    currentLoad,
		TodayAssigned:  todayAssigned,
		SuccessRate:    successRate,
		TotalCases:     totalCases,
		LoadScore:      loadScore,
		SpecialtyScore: specialtyScore,
		LocationScore:  locationScore,
		SuccessScore:   successScore,
		UrgencyScore:   urgencyScore,
		TotalScore:     roundFloat(totalScore, 2),
		Distance:       distance,
	}
}

func (s *DispatchServiceImpl) calculateLoadScore(currentLoad, todayAssigned, maxLoadPerDay int) float64 {
	if todayAssigned >= maxLoadPerDay {
		return 0
	}
	if currentLoad >= 10 {
		return 30
	}
	if currentLoad >= 5 {
		return 60
	}
	if currentLoad >= 3 {
		return 80
	}
	return 100
}

func (s *DispatchServiceImpl) calculateSpecialtyScore(specialties, caseTypePath string) float64 {
	if specialties == "" || caseTypePath == "" {
		return 60
	}

	specialtyList := strings.Split(specialties, ",")
	typeParts := strings.Split(caseTypePath, "/")

	matchCount := 0
	for _, specialty := range specialtyList {
		for _, typePart := range typeParts {
			if strings.Contains(strings.ToLower(typePart), strings.ToLower(specialty)) ||
				strings.Contains(strings.ToLower(specialty), strings.ToLower(typePart)) {
				matchCount++
			}
		}
	}

	if matchCount >= 3 {
		return 100
	}
	if matchCount >= 2 {
		return 85
	}
	if matchCount >= 1 {
		return 70
	}
	return 50
}

func (s *DispatchServiceImpl) calculateLocationScore(lng1, lat1, lng2, lat2 float64) float64 {
	if lng1 == 0 || lat1 == 0 || lng2 == 0 || lat2 == 0 {
		return 70
	}

	distance := calculateHaversineDistance(lng1, lat1, lng2, lat2)

	switch {
	case distance <= 5:
		return 100
	case distance <= 10:
		return 90
	case distance <= 20:
		return 75
	case distance <= 50:
		return 60
	case distance <= 100:
		return 40
	default:
		return 20
	}
}

func (s *DispatchServiceImpl) calculateSuccessScore(successRate float64, totalCases int) float64 {
	if totalCases == 0 {
		return 60
	}

	baseScore := successRate

	if totalCases >= 100 {
		baseScore *= 1.0
	} else if totalCases >= 50 {
		baseScore *= 0.95
	} else if totalCases >= 20 {
		baseScore *= 0.9
	} else if totalCases >= 10 {
		baseScore *= 0.85
	} else {
		baseScore *= 0.7
	}

	return math.Min(100, baseScore)
}

func (s *DispatchServiceImpl) calculateUrgencyScore(caseLevel, currentLoad int) float64 {
	if caseLevel == constants.CaseLevelExtraUrgent {
		if currentLoad < 3 {
			return 100
		} else if currentLoad < 5 {
			return 80
		} else {
			return 50
		}
	}
	if caseLevel == constants.CaseLevelUrgent {
		if currentLoad < 5 {
			return 90
		} else if currentLoad < 8 {
			return 70
		} else {
			return 40
		}
	}
	return 70
}

func (s *DispatchServiceImpl) getCurrentLoad(userID int64) int {
	var count int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ? AND status IN ? AND deleted_at IS NULL",
			userID, []int{constants.CaseStatusMediating, constants.CaseStatusApproving}).
		Count(&count)
	return int(count)
}

func (s *DispatchServiceImpl) getTodayAssigned(userID int64) int {
	today := time.Now().Format("2006-01-02")
	var count int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ? AND DATE(assigned_at) = ? AND deleted_at IS NULL", userID, today).
		Count(&count)
	return int(count)
}

func (s *DispatchServiceImpl) getSuccessRate(userID int64) float64 {
	var total, success int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ? AND status = ? AND deleted_at IS NULL", userID, constants.CaseStatusClosed).
		Count(&total)
	if total == 0 {
		return 0
	}

	database.GetDB().Table("dispute_case").
		Where("mediator_id = ? AND status = ? AND mediation_result = ? AND deleted_at IS NULL",
			userID, constants.CaseStatusClosed, constants.MediationResultSuccess).
		Count(&success)

	return roundFloat(float64(success)/float64(total)*100, 2)
}

func (s *DispatchServiceImpl) getTotalCases(userID int64) int {
	var count int64
	database.GetDB().Table("dispute_case").
		Where("mediator_id = ? AND deleted_at IS NULL", userID).
		Count(&count)
	return int(count)
}

func (s *DispatchServiceImpl) getDisputeTypePath(typeID int64) string {
	if typeID == 0 {
		return ""
	}

	var disputeType map[string]interface{}
	database.GetDB().Table("dispute_type").Where("id = ?", typeID).First(&disputeType)
	if disputeType == nil {
		return ""
	}

	path, _ := disputeType["type_path"].(string)
	typeName, _ := disputeType["type_name"].(string)

	if path != "" {
		return path + "/" + typeName
	}
	return typeName
}

func (s *DispatchServiceImpl) getDispatchConfigInternal(orgID int64) (DispatchConfig, error) {
	var config model.DispatchConfig
	result := database.GetDB().Where("org_id = ? AND status = 1", orgID).First(&config)
	if result.Error != nil || config.ID == 0 {
		return defaultDispatchConfig, nil
	}

	var dc DispatchConfig
	err := utils.FromJSON(config.ConfigContent, &dc)
	if err != nil {
		return defaultDispatchConfig, nil
	}

	return dc, nil
}

func (s *DispatchServiceImpl) assignCase(caseID, mediatorID, operatorID int64) error {
	updates := map[string]interface{}{
		"mediator_id": mediatorID,
		"status":      constants.CaseStatusMediating,
		"assigned_at": time.Now(),
		"assigned_by": operatorID,
	}

	err := database.GetDB().Table("dispute_case").
		Where("id = ?", caseID).
		Updates(updates).Error
	if err != nil {
		return err
	}

	mq.SendAsync(constants.MQTopicCaseAssign, map[string]interface{}{
		"caseId":     caseID,
		"mediatorId": mediatorID,
		"operatorId": operatorID,
		"assignTime": time.Now(),
	})

	return nil
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

	return earthRadius * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
