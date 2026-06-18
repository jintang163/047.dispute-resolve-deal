package impl

import (
	"context"
	"errors"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type ApprovalServiceImpl struct{}

func NewApprovalService() service.ApprovalService {
	return &ApprovalServiceImpl{}
}

func (s *ApprovalServiceImpl) SubmitApproval(ctx context.Context, caseID int64, userID int64, workflowID int64) (*model.ApprovalRecord, error) {
	var caseData model.DisputeCase
	database.GetDB().Where("id = ?", caseID).First(&caseData)

	if caseData.Status != constants.CaseStatusMediating {
		return nil, errors.New("只有调解中的案件才能提交审批")
	}

	var workflow model.WorkflowDefinition
	if workflowID == 0 {
		database.GetDB().Where("dispute_type_id = ? AND status = 1", caseData.TypeID).Order("version DESC").First(&workflow)
	} else {
		database.GetDB().Where("id = ?", workflowID).First(&workflow)
	}
	if workflow.ID == 0 {
		database.GetDB().Where("status = 1").Order("version DESC").First(&workflow)
	}

	tx := database.GetDB().Begin()

	roleMap := map[string]int32{
		"mediator": constants.RoleMediator,
		"leader":   constants.RoleLeader,
		"director": constants.RoleDirector,
	}

	approvers := []struct {
		Role  string `json:"role"`
		Level int32  `json:"level"`
	}{
		{"mediator", 1},
		{"leader", 2},
		{"director", 3},
	}

	for i, approver := range approvers {
		var user model.User
		if i == 0 {
			user.ID = caseData.MediatorID
		} else {
			database.GetDB().Where("role = ? AND organization_id = ?", roleMap[approver.Role], caseData.OrganizationID).
				Order("id ASC").First(&user)
		}

		record := &model.ApprovalRecord{
			CaseID:        caseID,
			WorkflowID:    workflow.ID,
			WorkflowName:  workflow.Name,
			NodeType:      1,
			NodeName:      []string{"调解员", "调解组长", "综治主任"}[i],
			ApproverID:    user.ID,
			ApproverName:  user.RealName,
			Status:        constants.ApprovalStatusPending,
			SortOrder:     int32(i + 1),
			Level:         approver.Level,
			TimeoutLevel:  0,
		}

		if i == 0 {
			record.Status = constants.ApprovalStatusProcessing
			record.Deadline = time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")
		}

		if err := tx.Create(record).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	caseData.Status = constants.CaseStatusApproving
	caseData.ApprovalWorkflowID = workflow.ID
	tx.Model(&caseData).Updates(caseData)

	history := &model.DisputeHistory{
		CaseID:     caseID,
		ActionType: constants.HistoryActionApproval,
		ActionName: "提交审批",
		Remark:     "提交审批，进入审批流程",
		OperatorID: userID,
	}
	tx.Create(history)

	tx.Commit()

	var firstApproval model.ApprovalRecord
	database.GetDB().Where("case_id = ? AND sort_order = 1", caseID).First(&firstApproval)

	mq.SendAsync(constants.MQTopicApprovalCreated, map[string]interface{}{
		"caseId":     caseID,
		"approvalId": firstApproval.ID,
		"approverId": firstApproval.ApproverID,
		"workflowId": workflow.ID,
	})

	return &firstApproval, nil
}

func (s *ApprovalServiceImpl) ApproveApproval(ctx context.Context, approvalID int64, userID int64, remark string) error {
	var record model.ApprovalRecord
	database.GetDB().Where("id = ?", approvalID).First(&record)

	if record.Status != constants.ApprovalStatusProcessing {
		return errors.New("该审批节点不处于处理中状态")
	}
	if record.ApproverID != userID {
		return errors.New("无权限处理该审批")
	}

	tx := database.GetDB().Begin()

	record.Status = constants.ApprovalStatusApproved
	record.ApproveAction = constants.ApprovalActionApprove
	record.ActionName = "通过"
	record.Remark = remark
	record.ApprovedAt = time.Now()
	tx.Save(&record)

	var nextRecord model.ApprovalRecord
	database.GetDB().Where("case_id = ? AND sort_order = ?", record.CaseID, record.SortOrder+1).First(&nextRecord)

	if nextRecord.ID > 0 {
		nextRecord.Status = constants.ApprovalStatusProcessing
		nextRecord.Deadline = time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")
		tx.Save(&nextRecord)

		mq.SendAsync(constants.MQTopicApprovalCreated, map[string]interface{}{
			"caseId":     record.CaseID,
			"approvalId": nextRecord.ID,
			"approverId": nextRecord.ApproverID,
		})
	} else {
		var caseData model.DisputeCase
		database.GetDB().Where("id = ?", record.CaseID).First(&caseData)
		caseData.Status = constants.CaseStatusClosed
		caseData.ClosedAt = time.Now()
		tx.Save(&caseData)

		history := &model.DisputeHistory{
			CaseID:     record.CaseID,
			ActionType: constants.HistoryActionApproval,
			ActionName: "审批完成",
			Remark:     "案件审批通过，已结案",
			OperatorID: userID,
		}
		tx.Create(history)
	}

	history := &model.DisputeHistory{
		CaseID:     record.CaseID,
		ActionType: constants.HistoryActionApproval,
		ActionName: "审批通过",
		Remark:     record.NodeName + "审批通过: " + remark,
		OperatorID: userID,
	}
	tx.Create(history)

	tx.Commit()
	return nil
}

func (s *ApprovalServiceImpl) RejectApproval(ctx context.Context, approvalID int64, userID int64, remark string) error {
	var record model.ApprovalRecord
	database.GetDB().Where("id = ?", approvalID).First(&record)

	if record.Status != constants.ApprovalStatusProcessing {
		return errors.New("该审批节点不处于处理中状态")
	}
	if record.ApproverID != userID {
		return errors.New("无权限处理该审批")
	}

	tx := database.GetDB().Begin()

	record.Status = constants.ApprovalStatusRejected
	record.ApproveAction = constants.ApprovalActionReject
	record.ActionName = "驳回"
	record.Remark = remark
	record.ApprovedAt = time.Now()
	tx.Save(&record)

	var caseData model.DisputeCase
	database.GetDB().Where("id = ?", record.CaseID).First(&caseData)
	caseData.Status = constants.CaseStatusMediating
	tx.Save(&caseData)

	database.GetDB().Model(&model.ApprovalRecord{}).
		Where("case_id = ? AND sort_order > ?", record.CaseID, record.SortOrder).
		Update("status", constants.ApprovalStatusCanceled)

	history := &model.DisputeHistory{
		CaseID:     record.CaseID,
		ActionType: constants.HistoryActionApproval,
		ActionName: "审批驳回",
		Remark:     record.NodeName + "审批驳回: " + remark,
		OperatorID: userID,
	}
	tx.Create(history)

	tx.Commit()
	return nil
}

func (s *ApprovalServiceImpl) ReturnApproval(ctx context.Context, approvalID int64, userID int64, remark string) error {
	var record model.ApprovalRecord
	database.GetDB().Where("id = ?", approvalID).First(&record)

	if record.Status != constants.ApprovalStatusProcessing {
		return errors.New("该审批节点不处于处理中状态")
	}
	if record.ApproverID != userID {
		return errors.New("无权限处理该审批")
	}

	tx := database.GetDB().Begin()

	record.Status = constants.ApprovalStatusReturned
	record.ApproveAction = constants.ApprovalActionReturn
	record.ActionName = "退回修改"
	record.Remark = remark
	record.ApprovedAt = time.Now()
	tx.Save(&record)

	var caseData model.DisputeCase
	database.GetDB().Where("id = ?", record.CaseID).First(&caseData)
	caseData.Status = constants.CaseStatusMediating
	tx.Save(&caseData)

	database.GetDB().Model(&model.ApprovalRecord{}).
		Where("case_id = ? AND sort_order > ?", record.CaseID, record.SortOrder).
		Update("status", constants.ApprovalStatusCanceled)

	history := &model.DisputeHistory{
		CaseID:     record.CaseID,
		ActionType: constants.HistoryActionApproval,
		ActionName: "退回修改",
		Remark:     record.NodeName + "退回修改: " + remark,
		OperatorID: userID,
	}
	tx.Create(history)

	tx.Commit()
	return nil
}

func (s *ApprovalServiceImpl) AddSignApproval(ctx context.Context, approvalID int64, userID int64, signUserID int64, remark string) error {
	var record model.ApprovalRecord
	database.GetDB().Where("id = ?", approvalID).First(&record)

	if record.Status != constants.ApprovalStatusProcessing {
		return errors.New("该审批节点不处于处理中状态")
	}
	if record.ApproverID != userID {
		return errors.New("无权限处理该审批")
	}

	var signUser model.User
	database.GetDB().Where("id = ?", signUserID).First(&signUser)

	tx := database.GetDB().Begin()

	newRecord := &model.ApprovalRecord{
		CaseID:        record.CaseID,
		WorkflowID:    record.WorkflowID,
		WorkflowName:  record.WorkflowName,
		NodeType:      2,
		NodeName:      "加签审批",
		ApproverID:    signUserID,
		ApproverName:  signUser.RealName,
		Status:        constants.ApprovalStatusProcessing,
		Remark:        remark,
		SortOrder:     record.SortOrder,
		Level:         record.Level,
		SignUserID:    signUserID,
		SignUserName:  signUser.RealName,
		Deadline:      time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
	}
	tx.Create(newRecord)

	record.Status = constants.ApprovalStatusWaiting
	record.ApproveAction = constants.ApprovalActionAddSign
	record.ActionName = "加签"
	record.SignUserID = signUserID
	record.SignUserName = signUser.RealName
	record.Remark = remark
	tx.Save(&record)

	history := &model.DisputeHistory{
		CaseID:     record.CaseID,
		ActionType: constants.HistoryActionApproval,
		ActionName: "加签审批",
		Remark:     record.NodeName + "加签给" + signUser.RealName + ": " + remark,
		OperatorID: userID,
	}
	tx.Create(history)

	tx.Commit()

	mq.SendAsync(constants.MQTopicApprovalCreated, map[string]interface{}{
		"caseId":     record.CaseID,
		"approvalId": newRecord.ID,
		"approverId": signUserID,
	})

	return nil
}

func (s *ApprovalServiceImpl) TransferApproval(ctx context.Context, approvalID int64, userID int64, transferUserID int64, remark string) error {
	var record model.ApprovalRecord
	database.GetDB().Where("id = ?", approvalID).First(&record)

	if record.Status != constants.ApprovalStatusProcessing {
		return errors.New("该审批节点不处于处理中状态")
	}
	if record.ApproverID != userID {
		return errors.New("无权限处理该审批")
	}

	var transferUser model.User
	database.GetDB().Where("id = ?", transferUserID).First(&transferUser)

	tx := database.GetDB().Begin()

	record.Status = constants.ApprovalStatusTransferred
	record.ApproveAction = constants.ApprovalActionTransfer
	record.ActionName = "转审"
	record.TransferUserID = transferUserID
	record.TransferUserName = transferUser.RealName
	record.Remark = remark
	tx.Save(&record)

	newRecord := &model.ApprovalRecord{
		CaseID:          record.CaseID,
		WorkflowID:      record.WorkflowID,
		WorkflowName:    record.WorkflowName,
		NodeType:        3,
		NodeName:        "转审审批",
		ApproverID:      transferUserID,
		ApproverName:    transferUser.RealName,
		Status:          constants.ApprovalStatusProcessing,
		SortOrder:       record.SortOrder,
		Level:           record.Level,
		TransferUserID:  transferUserID,
		TransferUserName: transferUser.RealName,
		Deadline:        time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
	}
	tx.Create(newRecord)

	history := &model.DisputeHistory{
		CaseID:     record.CaseID,
		ActionType: constants.HistoryActionApproval,
		ActionName: "转审审批",
		Remark:     record.NodeName + "转审给" + transferUser.RealName + ": " + remark,
		OperatorID: userID,
	}
	tx.Create(history)

	tx.Commit()

	mq.SendAsync(constants.MQTopicApprovalCreated, map[string]interface{}{
		"caseId":     record.CaseID,
		"approvalId": newRecord.ID,
		"approverId": transferUserID,
	})

	return nil
}

func (s *ApprovalServiceImpl) GetApprovalProgress(ctx context.Context, caseID int64) ([]*model.ApprovalRecord, error) {
	var records []*model.ApprovalRecord
	result := database.GetDB().Where("case_id = ?", caseID).Order("sort_order, created_at ASC").Find(&records)
	if result.Error != nil {
		return nil, result.Error
	}
	return records, nil
}

func (s *ApprovalServiceImpl) GetApprovalTodoList(ctx context.Context, userID int64, page, pageSize int) ([]*model.ApprovalRecord, int64, error) {
	var records []*model.ApprovalRecord
	var total int64

	db := database.GetDB().Model(&model.ApprovalRecord{}).
		Where("approver_id = ? AND status IN (?)", userID, []int32{constants.ApprovalStatusPending, constants.ApprovalStatusProcessing})

	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Preload("Case").Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records)

	return records, total, nil
}

func (s *ApprovalServiceImpl) GetApprovalDoneList(ctx context.Context, userID int64, page, pageSize int) ([]*model.ApprovalRecord, int64, error) {
	var records []*model.ApprovalRecord
	var total int64

	db := database.GetDB().Model(&model.ApprovalRecord{}).
		Where("approver_id = ? AND status IN (?)", userID, []int32{constants.ApprovalStatusApproved, constants.ApprovalStatusRejected, constants.ApprovalStatusReturned, constants.ApprovalStatusTransferred})

	db.Count(&total)
	offset := (page - 1) * pageSize
	db.Preload("Case").Offset(offset).Limit(pageSize).Order("approved_at DESC").Find(&records)

	return records, total, nil
}

func (s *ApprovalServiceImpl) ProcessTimeoutUpgrade(ctx context.Context) error {
	now := time.Now()
	var records []model.ApprovalRecord

	database.GetDB().Where("status = ? AND deadline IS NOT NULL", constants.ApprovalStatusProcessing).
		Where("deadline < ?", now.Format("2006-01-02 15:04:05")).
		Find(&records)

	processed := 0
	for _, record := range records {
		record.TimeoutLevel++

		upgrades := []struct {
			Level int32
			Role  int32
		}{
			{1, constants.RoleLeader},
			{2, constants.RoleDirector},
			{3, constants.RoleAdmin},
		}

		if record.TimeoutLevel <= 3 {
			var upgradeUser model.User
			database.GetDB().Where("role = ? AND organization_id = (SELECT organization_id FROM dispute_case WHERE id = ?)",
				upgrades[record.TimeoutLevel-1].Role, record.CaseID).Order("id ASC").First(&upgradeUser)

			if upgradeUser.ID > 0 {
				mq.SendAsync(constants.MQTopicApprovalTimeout, map[string]interface{}{
					"caseId":      record.CaseID,
					"approvalId":  record.ID,
					"timeoutLevel": record.TimeoutLevel,
					"upgradeUserId": upgradeUser.ID,
				})

				newRecord := &model.ApprovalRecord{
					CaseID:        record.CaseID,
					WorkflowID:    record.WorkflowID,
					WorkflowName:  record.WorkflowName,
					NodeType:      4,
					NodeName:      "超时升级审批",
					ApproverID:    upgradeUser.ID,
					ApproverName:  upgradeUser.RealName,
					Status:        constants.ApprovalStatusProcessing,
					SortOrder:     record.SortOrder,
					Level:         record.Level,
					TimeoutLevel:  record.TimeoutLevel,
					Deadline:      time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
				}
				database.GetDB().Create(newRecord)

				processed++
			}
		}

		database.GetDB().Model(&record).Update("timeout_level", record.TimeoutLevel)
	}

	logger.Info("Process timeout upgrade", logger.Int("processed", processed))
	return nil
}
