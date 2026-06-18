package main

import (
	"context"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	workflow "github.com/dispute-resolve/workflow-service/kitex_gen/workflow"
)

type WorkflowServiceImpl struct{}

func (s *WorkflowServiceImpl) SubmitApproval(ctx context.Context, req *workflow.SubmitApprovalRequest) (resp *workflow.SubmitApprovalResponse, err error) {
	resp = &workflow.SubmitApprovalResponse{Code: 0, Message: "success"}

	var caseData model.DisputeCase
	database.GetDB().Where("id = ? AND deleted_at IS NULL", req.CaseId).First(&caseData)
	if caseData.ID == 0 {
		resp.Code = 404
		resp.Message = "案件不存在"
		return resp, nil
	}

	var workflowDef model.WorkflowDefinition
	if req.WorkflowId > 0 {
		database.GetDB().Where("id = ? AND status = 1", req.WorkflowId).First(&workflowDef)
	} else {
		database.GetDB().Where("dispute_type_id = ? AND status = 1", caseData.TypeID).Order("version DESC").First(&workflowDef)
	}
	if workflowDef.ID == 0 {
		database.GetDB().Where("dispute_type_id IS NULL AND status = 1").Order("version DESC").First(&workflowDef)
	}

	if workflowDef.ID == 0 {
		resp.Code = 500
		resp.Message = "未找到审批流程定义"
		return resp, nil
	}

	tx := database.GetDB().Begin()

	approvalChain := s.buildApprovalChain(workflowDef, caseData)

	var firstApproval *model.ApprovalRecord
	for i, node := range approvalChain {
		record := &model.ApprovalRecord{
			CaseID:       req.CaseId,
			WorkflowID:   workflowDef.ID,
			WorkflowName: workflowDef.Name,
			NodeType:     int(node.NodeType),
			NodeName:     node.NodeName,
			ApproverID:   node.ApproverID,
			ApproverName: node.ApproverName,
			Status:       int32(constants.ApprovalStatusPending),
			Remark:       "",
			SortOrder:    i + 1,
			Level:        node.Level,
			Deadline:     time.Now().Add(time.Duration(node.TimeoutHours) * time.Hour),
		}

		if i == 0 {
			record.Status = constants.ApprovalStatusPending
			firstApproval = record
		}

		if err := tx.Create(record).Error; err != nil {
			tx.Rollback()
			resp.Code = 500
			resp.Message = "创建审批记录失败"
			logger.Error("Create approval record error", logger.Error(err))
			return resp, nil
		}
	}

	tx.Model(&model.DisputeCase{}).Where("id = ?", req.CaseId).Updates(map[string]interface{}{
		"status":       constants.CaseStatusApproving,
		"workflow_id":  workflowDef.ID,
	})

	history := &model.DisputeHistory{
		CaseID:     req.CaseId,
		ActionType: 3,
		ActionName: "提交审批",
		Remark:     "提交审批流程：" + workflowDef.Name,
		OperatorID: req.UserId,
	}
	tx.Create(history)

	tx.Commit()

	if firstApproval != nil {
		resp.Record = &workflow.ApprovalRecord{
			Id:           firstApproval.ID,
			CaseId:       firstApproval.CaseID,
			WorkflowId:   firstApproval.WorkflowID,
			WorkflowName: firstApproval.WorkflowName,
			NodeType:     int32(firstApproval.NodeType),
			NodeName:     firstApproval.NodeName,
			ApproverId:   firstApproval.ApproverID,
			ApproverName: firstApproval.ApproverName,
			Status:       firstApproval.Status,
			SortOrder:    int32(firstApproval.SortOrder),
			Level:        int32(firstApproval.Level),
			CreatedAt:    firstApproval.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *WorkflowServiceImpl) ProcessApproval(ctx context.Context, req *workflow.ProcessApprovalRequest) (resp *workflow.ProcessApprovalResponse, err error) {
	resp = &workflow.ProcessApprovalResponse{Code: 0, Message: "success"}

	var approval model.ApprovalRecord
	database.GetDB().Where("id = ?", req.ApprovalId).First(&approval)
	if approval.ID == 0 {
		resp.Code = 404
		resp.Message = "审批记录不存在"
		return resp, nil
	}

	if approval.ApproverID != req.UserId {
		resp.Code = 403
		resp.Message = "无权限处理此审批"
		return resp, nil
	}

	if approval.Status != constants.ApprovalStatusPending {
		resp.Code = 400
		resp.Message = "审批状态不允许处理"
		return resp, nil
	}

	tx := database.GetDB().Begin()

	actionName := constants.ApprovalActionMap[int(req.Action)]
	updates := map[string]interface{}{
		"status":         constants.ApprovalStatusPassed,
		"approve_action": int(req.Action),
		"action_name":    actionName,
		"remark":         req.Remark,
		"approved_at":    time.Now(),
	}

	switch req.Action {
	case constants.ApprovalActionReject:
		updates["status"] = constants.ApprovalStatusRejected
	case constants.ApprovalActionReturn:
		updates["status"] = constants.ApprovalStatusRejected
	case constants.ApprovalActionAddSign:
		updates["sign_user_id"] = req.TargetUserId
		var signUser model.User
		database.GetDB().Select("real_name").Where("id = ?", req.TargetUserId).First(&signUser)
		updates["sign_user_name"] = signUser.RealName

		signRecord := &model.ApprovalRecord{
			CaseID:       approval.CaseID,
			WorkflowID:   approval.WorkflowID,
			WorkflowName: approval.WorkflowName,
			NodeType:     2,
			NodeName:     "加签审批",
			ApproverID:   req.TargetUserId,
			ApproverName: signUser.RealName,
			Status:       constants.ApprovalStatusPending,
			SortOrder:    approval.SortOrder + 1,
			Level:        approval.Level,
			Deadline:     time.Now().Add(24 * time.Hour),
		}
		tx.Create(signRecord)
	case constants.ApprovalActionTransfer:
		updates["transfer_user_id"] = req.TargetUserId
		var transferUser model.User
		database.GetDB().Select("real_name").Where("id = ?", req.TargetUserId).First(&transferUser)
		updates["transfer_user_name"] = transferUser.RealName

		transferRecord := &model.ApprovalRecord{
			CaseID:       approval.CaseID,
			WorkflowID:   approval.WorkflowID,
			WorkflowName: approval.WorkflowName,
			NodeType:     3,
			NodeName:     "转审",
			ApproverID:   req.TargetUserId,
			ApproverName: transferUser.RealName,
			Status:       constants.ApprovalStatusPending,
			SortOrder:    approval.SortOrder + 1,
			Level:        approval.Level,
			Deadline:     time.Now().Add(24 * time.Hour),
		}
		tx.Create(transferRecord)
	}

	tx.Model(&approval).Updates(updates)

	if req.Action == constants.ApprovalActionPass {
		var nextApproval model.ApprovalRecord
		tx.Where("case_id = ? AND sort_order = ? AND status = ?",
			approval.CaseID, approval.SortOrder+1, constants.ApprovalStatusPending).
			Order("sort_order ASC").First(&nextApproval)

		if nextApproval.ID > 0 {
			tx.Model(&nextApproval).Update("status", constants.ApprovalStatusPending)
		} else {
			tx.Model(&model.DisputeCase{}).Where("id = ?", approval.CaseID).Updates(map[string]interface{}{
				"status": constants.CaseStatusClosed,
			})
		}
	}

	history := &model.DisputeHistory{
		CaseID:     approval.CaseID,
		ActionType: 4,
		ActionName: "审批处理",
		Remark:     actionName + "：" + req.Remark,
		OperatorID: req.UserId,
	}
	tx.Create(history)

	tx.Commit()

	return resp, nil
}

func (s *WorkflowServiceImpl) GetApprovalProgress(ctx context.Context, req *workflow.GetApprovalProgressRequest) (resp *workflow.GetApprovalProgressResponse, err error) {
	resp = &workflow.GetApprovalProgressResponse{Code: 0, Message: "success"}

	var records []model.ApprovalRecord
	database.GetDB().Where("case_id = ?", req.CaseId).Order("sort_order ASC, created_at ASC").Find(&records)

	resp.Records = make([]*workflow.ApprovalRecord, len(records))
	for i, r := range records {
		resp.Records[i] = &workflow.ApprovalRecord{
			Id:               r.ID,
			CaseId:           r.CaseID,
			WorkflowId:       r.WorkflowID,
			WorkflowName:     r.WorkflowName,
			NodeType:         int32(r.NodeType),
			NodeName:         r.NodeName,
			ApproverId:       r.ApproverID,
			ApproverName:     r.ApproverName,
			Status:           r.Status,
			Remark:           r.Remark,
			ApproveAction:    int32(r.ApproveAction),
			ActionName:       r.ActionName,
			SortOrder:        int32(r.SortOrder),
			SignUserId:       r.SignUserID,
			SignUserName:     r.SignUserName,
			TransferUserId:   r.TransferUserID,
			TransferUserName: r.TransferUserName,
			Level:            int32(r.Level),
			TimeoutLevel:     int32(r.TimeoutLevel),
			CreatedAt:        r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if r.ApprovedAt != nil {
			resp.Records[i].ApprovedAt = r.ApprovedAt.Format("2006-01-02 15:04:05")
		}
		if r.Deadline != nil {
			resp.Records[i].Deadline = r.Deadline.Format("2006-01-02 15:04:05")
		}
	}

	return resp, nil
}

func (s *WorkflowServiceImpl) GetApprovalTodoList(ctx context.Context, req *workflow.GetApprovalListRequest) (resp *workflow.GetApprovalListResponse, err error) {
	resp = &workflow.GetApprovalListResponse{Code: 0, Message: "success"}

	var records []model.ApprovalRecord
	var total int64

	db := database.GetDB().Model(&model.ApprovalRecord{}).
		Where("approver_id = ? AND status = ?", req.UserId, constants.ApprovalStatusPending)

	if req.Status > 0 {
		db = db.Where("status = ?", req.Status)
	}

	db.Count(&total)

	offset := int((req.Page - 1) * req.PageSize)
	db.Preload("Case").
		Offset(offset).Limit(int(req.PageSize)).
		Order("created_at DESC").
		Find(&records)

	resp.Total = total
	resp.Records = make([]*workflow.ApprovalRecord, len(records))
	for i, r := range records {
		resp.Records[i] = &workflow.ApprovalRecord{
			Id:           r.ID,
			CaseId:       r.CaseID,
			WorkflowName: r.WorkflowName,
			NodeType:     int32(r.NodeType),
			NodeName:     r.NodeName,
			ApproverName: r.ApproverName,
			Status:       r.Status,
			SortOrder:    int32(r.SortOrder),
			CreatedAt:    r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}

	return resp, nil
}

func (s *WorkflowServiceImpl) GetApprovalDoneList(ctx context.Context, req *workflow.GetApprovalListRequest) (resp *workflow.GetApprovalListResponse, err error) {
	resp = &workflow.GetApprovalListResponse{Code: 0, Message: "success"}

	var records []model.ApprovalRecord
	var total int64

	db := database.GetDB().Model(&model.ApprovalRecord{}).
		Where("approver_id = ? AND status IN ?", req.UserId, []int{constants.ApprovalStatusPassed, constants.ApprovalStatusRejected})

	if req.Status > 0 {
		db = db.Where("status = ?", req.Status)
	}

	db.Count(&total)

	offset := int((req.Page - 1) * req.PageSize)
	db.Offset(offset).Limit(int(req.PageSize)).
		Order("approved_at DESC").
		Find(&records)

	resp.Total = total
	resp.Records = make([]*workflow.ApprovalRecord, len(records))
	for i, r := range records {
		resp.Records[i] = &workflow.ApprovalRecord{
			Id:            r.ID,
			CaseId:        r.CaseID,
			WorkflowName:  r.WorkflowName,
			NodeType:      int32(r.NodeType),
			NodeName:      r.NodeName,
			ApproverName:  r.ApproverName,
			Status:        r.Status,
			ApproveAction: int32(r.ApproveAction),
			ActionName:    r.ActionName,
			Remark:        r.Remark,
			CreatedAt:     r.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		if r.ApprovedAt != nil {
			resp.Records[i].ApprovedAt = r.ApprovedAt.Format("2006-01-02 15:04:05")
		}
	}

	return resp, nil
}

func (s *WorkflowServiceImpl) ProcessTimeoutUpgrade(ctx context.Context) (resp *workflow.ProcessTimeoutUpgradeResponse, err error) {
	resp = &workflow.ProcessTimeoutUpgradeResponse{Code: 0, Message: "success"}

	now := time.Now()
	var pendingApprovals []model.ApprovalRecord
	database.GetDB().Where("status = ? AND deadline < ? AND timeout_level < 3",
		constants.ApprovalStatusPending, now).Find(&pendingApprovals)

	processedCount := 0
	for _, approval := range pendingApprovals {
		timeoutLevel := approval.TimeoutLevel + 1
		if timeoutLevel > 3 {
			timeoutLevel = 3
		}

		database.GetDB().Model(&approval).Updates(map[string]interface{}{
			"timeout_level": timeoutLevel,
		})

		var caseData model.DisputeCase
		database.GetDB().Select("case_no, title, mediator_id, org_id").
			Where("id = ?", approval.CaseID).First(&caseData)

		urgeType := constants.UrgeTypeSystem
		if timeoutLevel == 2 {
			urgeType = constants.UrgeTypeEscalate
		}

		urge := &model.DisputeUrge{
			CaseID:       approval.CaseID,
			UrgeType:     urgeType,
			UrgeContent:  "审批超时自动升级：" + approval.NodeName,
			OperatorID:   0,
			OperatorName: "系统",
		}
		database.GetDB().Create(urge)

		processedCount++
		logger.Info("Timeout approval upgraded",
			logger.Int64("approvalId", approval.ID),
			logger.Int64("caseId", approval.CaseID),
			logger.Int("timeoutLevel", timeoutLevel))
	}

	resp.ProcessedCount = int32(processedCount)
	return resp, nil
}

type approvalNode struct {
	NodeType      int32
	NodeName      string
	ApproverID    int64
	ApproverName  string
	Level         int
	TimeoutHours  int
}

func (s *WorkflowServiceImpl) buildApprovalChain(workflowDef model.WorkflowDefinition, caseData model.DisputeCase) []approvalNode {
	var org model.Organization
	database.GetDB().Where("id = ?", caseData.OrgID).First(&org)

	nodes := make([]approvalNode, 0)

	var mediators []model.User
	database.GetDB().Where("org_id = ? AND role = ? AND status = 1",
		caseData.OrgID, constants.RoleMediator).
		Order("id ASC").Find(&mediators)

	if caseData.MediatorID > 0 {
		for _, m := range mediators {
			if m.ID == caseData.MediatorID {
				nodes = append(nodes, approvalNode{
					NodeType:     1,
					NodeName:     "调解员处理",
					ApproverID:   m.ID,
					ApproverName: m.RealName,
					Level:        1,
					TimeoutHours: 24,
				})
				break
			}
		}
	}

	if len(mediators) > 0 && caseData.MediatorID == 0 {
		nodes = append(nodes, approvalNode{
			NodeType:     1,
			NodeName:     "调解员处理",
			ApproverID:   mediators[0].ID,
			ApproverName: mediators[0].RealName,
			Level:        1,
			TimeoutHours: 24,
		})
	}

	var leaders []model.User
	database.GetDB().Where("org_id = ? AND role = ? AND status = 1",
		caseData.OrgID, constants.RoleLeader).
		Order("id ASC").Find(&leaders)

	if len(leaders) > 0 {
		nodes = append(nodes, approvalNode{
			NodeType:     1,
			NodeName:     "调解组长复核",
			ApproverID:   leaders[0].ID,
			ApproverName: leaders[0].RealName,
			Level:        2,
			TimeoutHours: 24,
		})
	}

	var directors []model.User
	database.GetDB().Where("org_id = ? AND role = ? AND status = 1",
		caseData.OrgID, constants.RoleDirector).
		Order("id ASC").Find(&directors)

	if len(directors) > 0 {
		nodes = append(nodes, approvalNode{
			NodeType:     1,
			NodeName:     "综治中心主任审批",
			ApproverID:   directors[0].ID,
			ApproverName: directors[0].RealName,
			Level:        3,
			TimeoutHours: 24,
		})
	} else {
		var orgDirectors []model.User
		database.GetDB().Where("role = ? AND status = 1", constants.RoleDirector).
			Order("id ASC").Find(&orgDirectors)
		if len(orgDirectors) > 0 {
			nodes = append(nodes, approvalNode{
				NodeType:     1,
				NodeName:     "综治中心主任审批",
				ApproverID:   orgDirectors[0].ID,
				ApproverName: orgDirectors[0].RealName,
				Level:        3,
				TimeoutHours: 24,
			})
		}
	}

	return nodes
}
