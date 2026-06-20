package impl

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/gateway/service"
)

const (
	PendingUrgeThresholdHours    = 24
	MediatingUrgeThresholdDays   = 7
	UrgeToEscalateThresholdHours = 24
)

type TimeoutUrgeServiceImpl struct{}

func NewTimeoutUrgeService() service.TimeoutUrgeService {
	return &TimeoutUrgeServiceImpl{}
}

func (s *TimeoutUrgeServiceImpl) DetectAndUrgePendingCases(ctx context.Context) (int, error) {
	thresholdTime := time.Now().Add(-time.Duration(PendingUrgeThresholdHours) * time.Hour)

	var pendingCases []struct {
		ID             int64      `gorm:"column:id"`
		CaseNo         string     `gorm:"column:case_no"`
		Title          string     `gorm:"column:title"`
		OrganizationID int64      `gorm:"column:organization_id"`
		OrganizationName string   `gorm:"column:org_name"`
		CreatedAt      *time.Time `gorm:"column:created_at"`
		UrgencyCount   int        `gorm:"column:urgency_count"`
		UrgencyTime    *time.Time `gorm:"column:urgency_time"`
	}

	err := database.GetDB().Table("dispute_case").
		Where("status = ? AND created_at <= ? AND deleted_at IS NULL",
			constants.CaseStatusPending, thresholdTime).
		Find(&pendingCases).Error
	if err != nil {
		logger.Error("查询待分派超时案件失败", logger.Error(err))
		return 0, err
	}

	counted := 0
	for _, c := range pendingCases {
		if c.UrgencyTime != nil && time.Since(*c.UrgencyTime).Hours() < UrgeToEscalateThresholdHours {
			continue
		}

		content := fmt.Sprintf("【系统自动催办】案件[%s]待分派已超过%d小时，请综治受理员尽快分派处理。",
			c.CaseNo, PendingUrgeThresholdHours)

		err := s.UrgeCase(ctx, c.ID, constants.UrgeTypeSystem, content, 0, "系统自动")
		if err != nil {
			logger.Error(fmt.Sprintf("催办待分派案件失败 caseID=%d", c.ID), logger.Error(err))
			continue
		}
		counted++
	}

	logger.Info(fmt.Sprintf("待分派案件超时自动催办完成，共催办%d个案件", counted))
	return counted, nil
}

func (s *TimeoutUrgeServiceImpl) DetectAndUrgeMediatingCases(ctx context.Context) (int, error) {
	thresholdTime := time.Now().AddDate(0, 0, -MediatingUrgeThresholdDays)

	var mediatingCases []struct {
		ID               int64      `gorm:"column:id"`
		CaseNo           string     `gorm:"column:case_no"`
		Title            string     `gorm:"column:title"`
		MediatorID       int64      `gorm:"column:mediator_id"`
		MediatorName     string     `gorm:"column:mediator_name"`
		OrganizationID   int64      `gorm:"column:organization_id"`
		MediationStartTime *time.Time `gorm:"column:mediation_start_time"`
		LastProgressTime *time.Time `gorm:"column:last_progress_time"`
		CreatedAt        *time.Time `gorm:"column:created_at"`
		UrgencyCount     int        `gorm:"column:urgency_count"`
		UrgencyTime      *time.Time `gorm:"column:urgency_time"`
	}

	db := database.GetDB().Table("dispute_case").
		Where("status = ? AND deleted_at IS NULL", constants.CaseStatusMediating)

	db = db.Where(`(
		(last_progress_time IS NOT NULL AND last_progress_time <= ?)
		OR (last_progress_time IS NULL AND mediation_start_time IS NOT NULL AND mediation_start_time <= ?)
		OR (last_progress_time IS NULL AND mediation_start_time IS NULL AND created_at <= ?)
	)`, thresholdTime, thresholdTime, thresholdTime)

	err := db.Find(&mediatingCases).Error
	if err != nil {
		logger.Error("查询调解中超时案件失败", logger.Error(err))
		return 0, err
	}

	counted := 0
	for _, c := range mediatingCases {
		if c.UrgencyTime != nil && time.Since(*c.UrgencyTime).Hours() < UrgeToEscalateThresholdHours {
			continue
		}

		content := fmt.Sprintf("【系统自动催办】案件[%s]调解中已超过%d天无进展，请调解员[%s]及时跟进处理。",
			c.CaseNo, MediatingUrgeThresholdDays, c.MediatorName)

		err := s.UrgeCase(ctx, c.ID, constants.UrgeTypeSystem, content, 0, "系统自动")
		if err != nil {
			logger.Error(fmt.Sprintf("催办调解中案件失败 caseID=%d", c.ID), logger.Error(err))
			continue
		}
		counted++
	}

	logger.Info(fmt.Sprintf("调解中案件超时自动催办完成，共催办%d个案件", counted))
	return counted, nil
}

func (s *TimeoutUrgeServiceImpl) DetectAndEscalateUrgedCases(ctx context.Context) (int, error) {
	thresholdTime := time.Now().Add(-time.Duration(UrgeToEscalateThresholdHours) * time.Hour)

	var urgedCases []struct {
		ID               int64      `gorm:"column:id"`
		CaseNo           string     `gorm:"column:case_no"`
		Title            string     `gorm:"column:title"`
		Status           int32      `gorm:"column:status"`
		MediatorID       int64      `gorm:"column:mediator_id"`
		MediatorName     string     `gorm:"column:mediator_name"`
		OrganizationID   int64      `gorm:"column:organization_id"`
		OrganizationName string     `gorm:"column:org_name"`
		EscalateLevel    int32      `gorm:"column:escalate_level"`
		UrgencyTime      *time.Time `gorm:"column:urgency_time"`
		UrgencyCount     int        `gorm:"column:urgency_count"`
		CreatedAt        *time.Time `gorm:"column:created_at"`
		MediationStartTime *time.Time `gorm:"column:mediation_start_time"`
		LastProgressTime *time.Time `gorm:"column:last_progress_time"`
	}

	err := database.GetDB().Table("dispute_case").
		Where("urgency_time IS NOT NULL AND urgency_time <= ? AND escalate_level < 3 AND status < ? AND deleted_at IS NULL",
			thresholdTime, constants.CaseStatusClosed).
		Find(&urgedCases).Error
	if err != nil {
		logger.Error("查询催办后超时未处理案件失败", logger.Error(err))
		return 0, err
	}

	counted := 0
	for _, c := range urgedCases {
		var escalateType int
		var reason string
		if c.Status == constants.CaseStatusPending {
			escalateType = model.EscalationTypePendingTimeout
			timeoutHours := int(time.Since(*c.CreatedAt).Hours())
			reason = fmt.Sprintf("案件待分派超过%d小时，经催办后%d小时仍未分派处理，自动升级。",
				timeoutHours, UrgeToEscalateThresholdHours)
		} else {
			escalateType = model.EscalationTypeMediatingTimeout
			var baseTime *time.Time
			if c.LastProgressTime != nil {
				baseTime = c.LastProgressTime
			} else if c.MediationStartTime != nil {
				baseTime = c.MediationStartTime
			} else {
				baseTime = c.CreatedAt
			}
			timeoutHours := int(time.Since(*baseTime).Hours())
			reason = fmt.Sprintf("案件调解中超过%d小时无进展，经催办后%d小时仍未处理，自动升级。",
				timeoutHours, UrgeToEscalateThresholdHours)
		}

		_, err := s.EscalateCase(ctx, c.ID, escalateType, reason)
		if err != nil {
			logger.Error(fmt.Sprintf("自动升级案件失败 caseID=%d", c.ID), logger.Error(err))
			continue
		}
		counted++
	}

	logger.Info(fmt.Sprintf("催办后超时自动升级完成，共升级%d个案件", counted))
	return counted, nil
}

func (s *TimeoutUrgeServiceImpl) UrgeCase(ctx context.Context, caseID int64, urgeType int, urgeContent string, operatorID int64, operatorName string) error {
	var caseData struct {
		ID             int64  `gorm:"column:id"`
		CaseNo         string `gorm:"column:case_no"`
		Title          string `gorm:"column:title"`
		Status         int32  `gorm:"column:status"`
		MediatorID     int64  `gorm:"column:mediator_id"`
		MediatorName   string `gorm:"column:mediator_name"`
		OrganizationID int64  `gorm:"column:organization_id"`
		UrgencyCount   int    `gorm:"column:urgency_count"`
	}

	err := database.GetDB().Table("dispute_case").
		Where("id = ? AND deleted_at IS NULL", caseID).
		First(&caseData).Error
	if err != nil {
		return err
	}

	if caseData.Status >= constants.CaseStatusClosed {
		return fmt.Errorf("案件已结案，无法催办")
	}

	tx := database.GetDB().Begin()

	urge := &model.DisputeUrge{
		CaseID:           caseID,
		UrgeType:         urgeType,
		UrgeLevel:        1,
		EscalateTriggered: 0,
		UrgeContent:      urgeContent,
		OperatorID:       operatorID,
		OperatorName:     operatorName,
	}
	if err := tx.Create(urge).Error; err != nil {
		tx.Rollback()
		return err
	}

	now := time.Now()
	if err := tx.Table("dispute_case").
		Where("id = ?", caseID).
		Updates(map[string]interface{}{
			"urgency_time":  now,
			"urgency_count": caseData.UrgencyCount + 1,
		}).Error; err != nil {
		tx.Rollback()
		return err
	}

	detail, _ := json.Marshal(map[string]interface{}{
		"urgeType":    urgeType,
		"urgeLevel":   1,
		"urgeContent": urgeContent,
	})
	history := map[string]interface{}{
		"case_id":          caseID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "AUTO_URGE",
		"operation_detail": string(detail),
		"operator_id":      operatorID,
		"operator_name":    operatorName,
		"remark":           urgeContent,
	}
	if err := tx.Table("dispute_case_history").Create(history).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	go func() {
		handlerID := caseData.MediatorID
		handlerName := caseData.MediatorName
		if caseData.Status == constants.CaseStatusPending {
			handlerID = 0
			handlerName = "综治受理员"
		}
		msg := map[string]interface{}{
			"caseId":         caseID,
			"caseNo":         caseData.CaseNo,
			"title":          caseData.Title,
			"orgId":          caseData.OrganizationID,
			"urgencyContent": urgeContent,
			"handlerId":      handlerID,
			"handlerName":    handlerName,
			"urgeType":       urgeType,
			"urgeBy":         operatorName,
			"isAuto":         true,
		}
		mq.SendMessage(constants.MQTopicCaseUrge, msg)
	}()

	return nil
}

func (s *TimeoutUrgeServiceImpl) EscalateCase(ctx context.Context, caseID int64, escalationType int, reason string) (*model.DisputeEscalation, error) {
	var caseData struct {
		ID               int64  `gorm:"column:id"`
		CaseNo           string `gorm:"column:case_no"`
		Title            string `gorm:"column:title"`
		Status           int32  `gorm:"column:status"`
		MediatorID       int64  `gorm:"column:mediator_id"`
		MediatorName     string `gorm:"column:mediator_name"`
		OrganizationID   int64  `gorm:"column:organization_id"`
		OrganizationName string `gorm:"column:org_name"`
		EscalateLevel    int32  `gorm:"column:escalate_level"`
		UrgencyCount     int    `gorm:"column:urgency_count"`
		UrgencyTime      *time.Time `gorm:"column:urgency_time"`
	}

	err := database.GetDB().Table("dispute_case").
		Where("id = ? AND deleted_at IS NULL", caseID).
		First(&caseData).Error
	if err != nil {
		return nil, err
	}

	if caseData.EscalateLevel >= 3 {
		return nil, fmt.Errorf("案件已达到最高升级级别")
	}

	newLevel := caseData.EscalateLevel + 1
	var targetRole int
	var targetRoleName string
	switch newLevel {
	case 1:
		targetRole = constants.RoleLeader
		targetRoleName = "组长"
	case 2:
		targetRole = constants.RoleDirector
		targetRoleName = "主任"
	default:
		targetRole = constants.RoleDirector
		targetRoleName = "领导"
		newLevel = 3
	}

	var targetUser struct {
		ID       int64  `gorm:"column:id"`
		RealName string `gorm:"column:real_name"`
	}
	err = database.GetDB().Table("sys_user").
		Where("organization_id = ? AND role = ? AND status = 1 AND deleted_at IS NULL",
			caseData.OrganizationID, targetRole).
		Order("id ASC").
		First(&targetUser).Error
	if err != nil {
		err = database.GetDB().Table("sys_user").
			Where("role <= ? AND status = 1 AND deleted_at IS NULL", targetRole).
			Order("role ASC, id ASC").
			First(&targetUser).Error
		if err != nil {
			logger.Error(fmt.Sprintf("未找到升级目标角色用户 role=%d orgId=%d", targetRole, caseData.OrganizationID), logger.Error(err))
			return nil, fmt.Errorf("未找到可升级的上级岗位人员")
		}
	}

	fromLevel := 0
	fromUserID := caseData.MediatorID
	fromUserName := caseData.MediatorName
	if caseData.Status == constants.CaseStatusPending {
		fromLevel = 0
		fromUserID = 0
		fromUserName = "未分派"
	} else if caseData.EscalateLevel == 0 {
		fromLevel = 1
	} else {
		fromLevel = int(caseData.EscalateLevel)
	}

	var firstUrgeTime *time.Time
	var urgeRecord model.DisputeUrge
	database.GetDB().Where("case_id = ?", caseID).Order("created_at ASC").First(&urgeRecord)
	if urgeRecord.ID > 0 {
		firstUrgeTime = &urgeRecord.CreatedAt
	}

	now := time.Now()
	timeoutHours := 0
	if caseData.UrgencyTime != nil {
		timeoutHours = int(now.Sub(*caseData.UrgencyTime).Hours())
	}

	tx := database.GetDB().Begin()

	escalation := &model.DisputeEscalation{
		CaseID:        caseID,
		CaseNo:        caseData.CaseNo,
		EscalateType:  escalationType,
		FromLevel:     fromLevel,
		ToLevel:       int(newLevel),
		FromUserID:    fromUserID,
		FromUserName:  fromUserName,
		ToUserID:      targetUser.ID,
		ToUserName:    targetUser.RealName,
		ToOrgID:       caseData.OrganizationID,
		ToOrgName:     caseData.OrganizationName,
		Reason:        reason,
		UrgeCount:     caseData.UrgencyCount,
		FirstUrgeTime: firstUrgeTime,
		TimeoutHours:  timeoutHours,
		OperatorID:    0,
		OperatorName:  "系统自动",
		Status:        model.EscalationStatusPending,
	}
	if err := tx.Create(escalation).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	updates := map[string]interface{}{
		"escalate_level": newLevel,
		"escalate_time":  now,
	}
	if caseData.Status == constants.CaseStatusPending {
		updates["mediator_id"] = targetUser.ID
		updates["mediator_name"] = targetUser.RealName
		updates["mediator_time"] = now
		updates["status"] = constants.CaseStatusMediating
		updates["mediation_start_time"] = now
		updates["last_progress_time"] = now
	} else {
		updates["mediator_id"] = targetUser.ID
		updates["mediator_name"] = targetUser.RealName
		updates["last_progress_time"] = now
	}

	if err := tx.Table("dispute_case").
		Where("id = ?", caseID).
		Updates(updates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Table("dispute_urge").
		Where("case_id = ? AND escalate_triggered = 0", caseID).
		Update("escalate_triggered", 1).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	detail, _ := json.Marshal(map[string]interface{}{
		"escalationType": escalationType,
		"fromLevel":      fromLevel,
		"toLevel":        newLevel,
		"toUserId":       targetUser.ID,
		"toUserName":     targetUser.RealName,
		"toRoleName":     targetRoleName,
		"reason":         reason,
	})
	history := map[string]interface{}{
		"case_id":          caseID,
		"case_no":          caseData.CaseNo,
		"operation_type":   "AUTO_ESCALATE",
		"operation_detail": string(detail),
		"operator_id":      0,
		"operator_name":    "系统自动",
		"old_status":       caseData.Status,
		"new_status":       updates["status"],
		"remark":           fmt.Sprintf("案件自动升级至%s[%s]处理", targetRoleName, targetUser.RealName),
	}
	if err := tx.Table("dispute_case_history").Create(history).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	go func() {
		msg := map[string]interface{}{
			"caseId":       caseID,
			"caseNo":       caseData.CaseNo,
			"title":        caseData.Title,
			"escalationId": escalation.ID,
			"fromLevel":    fromLevel,
			"toLevel":      newLevel,
			"toUserId":     targetUser.ID,
			"toUserName":   targetUser.RealName,
			"toRoleName":   targetRoleName,
			"reason":       reason,
			"urgeCount":    caseData.UrgencyCount,
		}
		mq.SendMessage(constants.MQTopicCaseStatus, msg)
		mq.SendMessage(constants.MQTopicNotification, map[string]interface{}{
			"type":    "case_escalation",
			"userId":  targetUser.ID,
			"title":   "案件升级通知",
			"content": fmt.Sprintf("案件[%s]已升级至您处理，请及时跟进。原因：%s", caseData.CaseNo, reason),
			"data":    msg,
		})
	}()

	return escalation, nil
}

func (s *TimeoutUrgeServiceImpl) GetEscalationList(ctx context.Context, orgID int64, toLevel int, status int32, page, pageSize int) ([]*model.DisputeEscalation, int64, error) {
	var list []*model.DisputeEscalation
	var total int64

	db := database.GetDB().Model(&model.DisputeEscalation{})
	if orgID > 0 {
		db = db.Where("to_org_id = ?", orgID)
	}
	if toLevel > 0 {
		db = db.Where("to_level = ?", toLevel)
	}
	if status > 0 {
		db = db.Where("status = ?", status)
	}

	db.Count(&total)
	offset := (page - 1) * pageSize
	err := db.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

func (s *TimeoutUrgeServiceImpl) GetEscalationDetail(ctx context.Context, escalationID int64) (*model.DisputeEscalation, error) {
	var escalation model.DisputeEscalation
	err := database.GetDB().Where("id = ?", escalationID).First(&escalation).Error
	if err != nil {
		return nil, err
	}
	return &escalation, nil
}

func (s *TimeoutUrgeServiceImpl) GetCaseEscalationList(ctx context.Context, caseID int64) ([]*model.DisputeEscalation, error) {
	var list []*model.DisputeEscalation
	err := database.GetDB().Where("case_id = ?", caseID).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (s *TimeoutUrgeServiceImpl) GetCaseUrgeList(ctx context.Context, caseID int64) ([]*model.DisputeUrge, error) {
	var list []*model.DisputeUrge
	err := database.GetDB().Where("case_id = ?", caseID).Order("created_at DESC").Find(&list).Error
	return list, err
}

func (s *TimeoutUrgeServiceImpl) HandleEscalation(ctx context.Context, escalationID int64, operatorID int64, operatorName string, remark string) error {
	tx := database.GetDB().Begin()

	err := tx.Model(&model.DisputeEscalation{}).
		Where("id = ?", escalationID).
		Updates(map[string]interface{}{
			"status":       model.EscalationStatusProcessing,
			"operator_id":  operatorID,
			"operator_name": operatorName,
			"remark":       remark,
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var escalation model.DisputeEscalation
	tx.Where("id = ?", escalationID).First(&escalation)

	history := map[string]interface{}{
		"case_id":          escalation.CaseID,
		"case_no":          escalation.CaseNo,
		"operation_type":   "ESCALATION_HANDLE",
		"operation_detail": remark,
		"operator_id":      operatorID,
		"operator_name":    operatorName,
		"remark":           "接收升级案件处理",
	}
	tx.Table("dispute_case_history").Create(history)

	return tx.Commit().Error
}

func (s *TimeoutUrgeServiceImpl) CloseEscalation(ctx context.Context, escalationID int64, operatorID int64, operatorName string, remark string) error {
	tx := database.GetDB().Begin()

	err := tx.Model(&model.DisputeEscalation{}).
		Where("id = ?", escalationID).
		Updates(map[string]interface{}{
			"status":        model.EscalationStatusClosed,
			"operator_id":   operatorID,
			"operator_name": operatorName,
			"remark":        remark,
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	var escalation model.DisputeEscalation
	tx.Where("id = ?", escalationID).First(&escalation)

	history := map[string]interface{}{
		"case_id":          escalation.CaseID,
		"case_no":          escalation.CaseNo,
		"operation_type":   "ESCALATION_CLOSE",
		"operation_detail": remark,
		"operator_id":      operatorID,
		"operator_name":    operatorName,
		"remark":           "关闭升级记录",
	}
	tx.Table("dispute_case_history").Create(history)

	return tx.Commit().Error
}
