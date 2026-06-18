package impl

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/workflow"
	"github.com/dispute-resolve/gateway/service"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ApprovalServiceImpl struct{}

func NewApprovalService() service.ApprovalService {
	workflow.InitFlowable()
	return &ApprovalServiceImpl{}
}

func (s *ApprovalServiceImpl) SubmitApproval(ctx context.Context, caseID int64, userID int64, workflowID int64) (*model.ApprovalRecord, error) {
	var caseData model.DisputeCase
	database.GetDB().Where("id = ?", caseID).First(&caseData)

	if caseData.Status != constants.CaseStatusMediating {
		return nil, errors.New("只有调解中的案件才能提交审批")
	}

	var wfDef model.WorkflowDefinition
	if workflowID == 0 {
		database.GetDB().Where("dispute_type_id = ? AND status = 1", caseData.TypeID).Order("version DESC").First(&wfDef)
	} else {
		database.GetDB().Where("id = ?", workflowID).First(&wfDef)
	}
	if wfDef.ID == 0 {
		database.GetDB().Where("status = 1").Order("version DESC").First(&wfDef)
	}

	tx := database.GetDB().Begin()

	processInstanceID, err := workflow.StartDisputeApproval(caseID, &caseData)
	useFlowable := err == nil

	if !useFlowable {
		logger.Warn("Flowable start process failed, use local mode",
			zap.Int64("caseId", caseID),
			logger.Error(err),
		)
	}

	caseData.Status = constants.CaseStatusApproving
	caseData.WorkflowID = wfDef.ID
	caseData.ApprovalWorkflowID = wfDef.ID
	tx.Model(&caseData).Updates(map[string]interface{}{
		"status":               caseData.Status,
		"workflow_id":          wfDef.ID,
		"approval_workflow_id": wfDef.ID,
	})

	var processInstIDStr string
	if useFlowable {
		processInstIDStr = processInstanceID
	} else {
		processInstIDStr = "local_" + strconv.FormatInt(caseID, 10)
	}

	history := &model.DisputeHistory{
		CaseID:     caseID,
		ActionType: constants.HistoryActionApproval,
		ActionName: "提交审批",
		Remark:     fmt.Sprintf("提交审批，启动流程实例: %s", processInstIDStr),
		OperatorID: userID,
	}
	tx.Create(history)

	var firstTaskID string
	var firstApproverID int64
	var firstApproverName string

	if useFlowable {
		tasks, taskErr := workflow.GetFlowableClient().GetProcessInstanceTasks(processInstanceID)
		if taskErr == nil && len(tasks) > 0 {
			firstTask := tasks[0]
			firstTaskID = firstTask.ID
			if firstTask.Assignee != "" {
				if id, parseErr := strconv.ParseInt(firstTask.Assignee, 10, 64); parseErr == nil {
					firstApproverID = id
					var approver model.User
					database.GetDB().Where("id = ?", id).First(&approver)
					firstApproverName = approver.RealName
				}
			}
		}
	}

	if firstApproverID == 0 {
		firstApproverID = caseData.MediatorID
		firstApproverName = caseData.MediatorName
	}

	deadline := time.Now().Add(24 * time.Hour)
	firstApproval := &model.ApprovalRecord{
		CaseID:        caseID,
		WorkflowID:    wfDef.ID,
		WorkflowName:  wfDef.Name,
		NodeType:      1,
		NodeName:      "调解员审批",
		ApproverID:    firstApproverID,
		ApproverName:  firstApproverName,
		Status:        constants.ApprovalStatusProcessing,
		SortOrder:     1,
		Level:         1,
		TimeoutLevel:  0,
		Deadline:      &deadline,
	}

	if err := tx.Create(firstApproval).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	roleMap := map[int]int32{
		1: constants.RoleLeader,
		2: constants.RoleDirector,
	}
	nodeNames := []string{"调解员", "调解组长", "综治主任"}

	for i := 1; i < 3; i++ {
		var user model.User
		database.GetDB().Where("role = ? AND organization_id = ?",
			roleMap[i],
			caseData.OrganizationID,
		).Order("id ASC").First(&user)

		record := &model.ApprovalRecord{
			CaseID:        caseID,
			WorkflowID:    wfDef.ID,
			WorkflowName:  wfDef.Name,
			NodeType:      1,
			NodeName:      nodeNames[i],
			ApproverID:    user.ID,
			ApproverName:  user.RealName,
			Status:        constants.ApprovalStatusPending,
			SortOrder:     int32(i + 1),
			Level:         int32(i + 1),
			TimeoutLevel:  0,
		}

		if err := tx.Create(record).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	tx.Commit()

	if useFlowable && firstTaskID != "" && firstApproverID > 0 {
		workflow.GetFlowableClient().ClaimTask(firstTaskID, strconv.FormatInt(firstApproverID, 10))
	}

	mq.SendAsync(constants.MQTopicApprovalNotify, map[string]interface{}{
		"caseId":            caseID,
		"approvalId":        firstApproval.ID,
		"approverId":        firstApproverID,
		"workflowId":        wfDef.ID,
		"processInstanceId": processInstIDStr,
		"taskId":            firstTaskID,
	})

	logger.Info("Approval submitted",
		zap.Int64("caseId", caseID),
		zap.String("processInstanceId", processInstIDStr),
		zap.Int64("firstApproverId", firstApproverID),
		zap.Bool("useFlowable", useFlowable),
	)

	return firstApproval, nil
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

	approvedAt := time.Now()
	record.Status = constants.ApprovalStatusApproved
	record.ApproveAction = constants.ApprovalActionPass
	record.ActionName = "通过"
	record.Remark = remark
	record.ApprovedAt = &approvedAt
	tx.Save(&record)

	s.callFlowableIfAvailable(record.CaseID, userID, constants.ApprovalActionPass, remark)

	s.localApproveNext(tx, &record)

	history := &model.DisputeHistory{
		CaseID:     record.CaseID,
		ActionType: constants.HistoryActionApproval,
		ActionName: "审批通过",
		Remark:     record.NodeName + "审批通过: " + remark,
		OperatorID: userID,
	}
	tx.Create(history)

	tx.Commit()

	go workflow.SyncApprovalRecordsFromFlowable(record.CaseID)

	return nil
}

func (s *ApprovalServiceImpl) localApproveNext(tx *gorm.DB, record *model.ApprovalRecord) {
	var nextRecord model.ApprovalRecord
	database.GetDB().Where("case_id = ? AND sort_order = ?", record.CaseID, record.SortOrder+1).First(&nextRecord)

	if nextRecord.ID > 0 {
		deadline := time.Now().Add(24 * time.Hour)
		tx.Model(&nextRecord).Updates(map[string]interface{}{
			"status":   constants.ApprovalStatusProcessing,
			"deadline": &deadline,
		})

		mq.SendAsync(constants.MQTopicApprovalNotify, map[string]interface{}{
			"caseId":     record.CaseID,
			"approvalId": nextRecord.ID,
			"approverId": nextRecord.ApproverID,
		})
	} else {
		closedAt := time.Now()
		tx.Model(&model.DisputeCase{}).Where("id = ?", record.CaseID).Updates(map[string]interface{}{
			"status":    constants.CaseStatusClosed,
			"closed_at": &closedAt,
		})

		history := &model.DisputeHistory{
			CaseID:     record.CaseID,
			ActionType: constants.HistoryActionApproval,
			ActionName: "审批完成",
			Remark:     "案件审批通过，已结案",
			OperatorID: record.ApproverID,
		}
		tx.Create(history)
	}
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

	approvedAt := time.Now()
	record.Status = constants.ApprovalStatusRejected
	record.ApproveAction = constants.ApprovalActionReject
	record.ActionName = "驳回"
	record.Remark = remark
	record.ApprovedAt = &approvedAt
	tx.Save(&record)

	s.callFlowableIfAvailable(record.CaseID, userID, constants.ApprovalActionReject, remark)

	tx.Model(&model.DisputeCase{}).Where("id = ?", record.CaseID).Update("status", constants.CaseStatusMediating)

	tx.Model(&model.ApprovalRecord{}).
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

	approvedAt := time.Now()
	record.Status = constants.ApprovalStatusReturned
	record.ApproveAction = constants.ApprovalActionReturn
	record.ActionName = "退回修改"
	record.Remark = remark
	record.ApprovedAt = &approvedAt
	tx.Save(&record)

	s.callFlowableIfAvailable(record.CaseID, userID, constants.ApprovalActionReturn, remark)

	tx.Model(&model.DisputeCase{}).Where("id = ?", record.CaseID).Update("status", constants.CaseStatusMediating)

	tx.Model(&model.ApprovalRecord{}).
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

	processInstanceID, procErr := workflow.GetProcessInstanceByBusinessKey(record.CaseID)
	if procErr == nil && processInstanceID != "" {
		taskID, taskErr := workflow.GetCurrentTaskIDByProcessInstance(processInstanceID)
		if taskErr == nil && taskID != "" {
			workflow.AddSignTask(taskID, userID, signUserID, remark)
		}
	}

	deadline := time.Now().Add(24 * time.Hour)
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
		Deadline:      &deadline,
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

	mq.SendAsync(constants.MQTopicApprovalNotify, map[string]interface{}{
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

	processInstanceID, procErr := workflow.GetProcessInstanceByBusinessKey(record.CaseID)
	if procErr == nil && processInstanceID != "" {
		taskID, taskErr := workflow.GetCurrentTaskIDByProcessInstance(processInstanceID)
		if taskErr == nil && taskID != "" {
			workflow.DelegateApprovalTask(taskID, userID, transferUserID, remark)
		}
	}

	record.Status = constants.ApprovalStatusTransferred
	record.ApproveAction = constants.ApprovalActionTransfer
	record.ActionName = "转审"
	record.TransferUserID = transferUserID
	record.TransferUserName = transferUser.RealName
	record.Remark = remark
	tx.Save(&record)

	deadline := time.Now().Add(24 * time.Hour)
	newRecord := &model.ApprovalRecord{
		CaseID:           record.CaseID,
		WorkflowID:       record.WorkflowID,
		WorkflowName:     record.WorkflowName,
		NodeType:         3,
		NodeName:         "转审审批",
		ApproverID:       transferUserID,
		ApproverName:     transferUser.RealName,
		Status:           constants.ApprovalStatusProcessing,
		SortOrder:        record.SortOrder,
		Level:            record.Level,
		TransferUserID:   transferUserID,
		TransferUserName: transferUser.RealName,
		Deadline:         &deadline,
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

	mq.SendAsync(constants.MQTopicApprovalNotify, map[string]interface{}{
		"caseId":     record.CaseID,
		"approvalId": newRecord.ID,
		"approverId": transferUserID,
	})

	return nil
}

func (s *ApprovalServiceImpl) GetApprovalProgress(ctx context.Context, caseID int64) ([]*model.ApprovalRecord, error) {
	processInstanceID, err := workflow.GetProcessInstanceByBusinessKey(caseID)
	if err == nil && processInstanceID != "" {
		timeline, timelineErr := workflow.GetApprovalTimeline(processInstanceID)
		if timelineErr == nil && len(timeline) > 0 {
			records := make([]*model.ApprovalRecord, 0, len(timeline))
			for _, node := range timeline {
				record := &model.ApprovalRecord{
					CaseID:        caseID,
					NodeName:      node.NodeName,
					ApproverID:    node.ApproverID,
					ApproverName:  node.Approver,
					Status:        node.Status,
					Remark:        node.Remark,
					SortOrder:     node.SortOrder,
					ApproveAction: node.Action,
					ActionName:    node.ActionText,
				}
				if !node.CreateTime.IsZero() {
					record.CreatedAt = node.CreateTime
				}
				if !node.HandleTime.IsZero() {
					record.ApprovedAt = &node.HandleTime
				}
				records = append(records, record)
			}
			return records, nil
		}
	}

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
		Where("deadline < ?", now).
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
				mq.SendAsync(constants.MQTopicApprovalNotify, map[string]interface{}{
					"caseId":        record.CaseID,
					"approvalId":    record.ID,
					"timeoutLevel":  record.TimeoutLevel,
					"upgradeUserId": upgradeUser.ID,
					"type":          "timeout",
				})

				deadline := time.Now().Add(24 * time.Hour)
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
					Deadline:      &deadline,
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

func (s *ApprovalServiceImpl) callFlowableIfAvailable(caseID int64, userID int64, action int, remark string) {
	processInstanceID, err := workflow.GetProcessInstanceByBusinessKey(caseID)
	if err != nil {
		logger.Debug("No flowable process found for case",
			zap.Int64("caseId", caseID),
			logger.Error(err),
		)
		return
	}

	taskID, taskErr := workflow.GetCurrentTaskIDByProcessInstance(processInstanceID)
	if taskErr != nil || taskID == "" {
		logger.Debug("No active task found",
			zap.Int64("caseId", caseID),
			zap.String("processInstanceId", processInstanceID),
			logger.Error(taskErr),
		)
		return
	}

	flowErr := workflow.CompleteApprovalTask(taskID, userID, action, remark)
	if flowErr != nil {
		logger.Warn("Flowable task action failed",
			zap.String("taskId", taskID),
			zap.Int("action", action),
			logger.Error(flowErr),
		)
	}
}
