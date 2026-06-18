package workflow

import (
	"fmt"
	"strconv"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"go.uber.org/zap"
)

const (
	DisputeApprovalProcessKey = "dispute_approval"
	BusinessKeyPrefix         = "case_"
	ActionApproved            = "approved"
	ActionRejected            = "rejected"
	ActionReturned            = "returned"
)

func initFlowableClientIfNeeded() {
	GetFlowableClient()
}

func StartDisputeApproval(caseId int64, caseData *model.DisputeCase) (string, error) {
	initFlowableClientIfNeeded()
	client := GetFlowableClient()

	businessKey := fmt.Sprintf("%s%d", BusinessKeyPrefix, caseId)

	variables := map[string]interface{}{
		"caseId":         caseId,
		"caseNo":         caseData.CaseNo,
		"caseTitle":      caseData.Title,
		"applicantName":  caseData.ApplicantName,
		"respondentName": caseData.RespondentName,
		"mediatorId":     strconv.FormatInt(caseData.MediatorID, 10),
		"mediatorName":   caseData.MediatorName,
		"organizationId": caseData.OrganizationID,
		"caseType":       caseData.TypeName,
		"caseLevel":      caseData.Level,
		"startUserId":    caseData.CreatedBy,
	}

	processInstanceId, err := client.StartProcess(DisputeApprovalProcessKey, businessKey, variables)
	if err != nil {
		logger.Error("Start dispute approval process failed",
			zap.Int64("caseId", caseId),
			logger.Error(err),
		)
		return "", err
	}

	logger.Info("Dispute approval process started",
		zap.Int64("caseId", caseId),
		zap.String("processInstanceId", processInstanceId),
	)

	return processInstanceId, nil
}

func CompleteApprovalTask(taskId string, approverId int64, action int, remark string) error {
	initFlowableClientIfNeeded()
	client := GetFlowableClient()

	variables := map[string]interface{}{
		"approverId":    strconv.FormatInt(approverId, 10),
		"approverRemark": remark,
		"approveAction":  action,
		"actionTime":     time.Now().Format(time.RFC3339),
	}

	switch action {
	case constants.ApprovalActionPass:
		variables["approved"] = true
		variables["action"] = ActionApproved
		if err := client.CompleteTask(taskId, variables); err != nil {
			return err
		}
	case constants.ApprovalActionReject:
		if err := client.RejectTask(taskId, remark); err != nil {
			return err
		}
	case constants.ApprovalActionReturn:
		variables["approved"] = false
		variables["action"] = ActionReturned
		variables["returnReason"] = remark
		if err := client.CompleteTask(taskId, variables); err != nil {
			return err
		}
	case constants.ApprovalActionTransfer:
		return fmt.Errorf("transfer action should use DelegateTask instead")
	case constants.ApprovalActionAddSign:
		return fmt.Errorf("add sign action should use AddMultiInstance instead")
	default:
		return fmt.Errorf("unknown approval action: %d", action)
	}

	logger.Info("Approval task completed",
		zap.String("taskId", taskId),
		zap.Int64("approverId", approverId),
		zap.Int("action", action),
	)

	return nil
}

func DelegateApprovalTask(taskId string, fromUserId int64, toUserId int64, remark string) error {
	initFlowableClientIfNeeded()
	client := GetFlowableClient()

	if err := client.AddComment(taskId, fmt.Sprintf("用户%d转办给用户%d: %s", fromUserId, toUserId, remark)); err != nil {
		logger.Warn("Add delegate comment failed", zap.String("taskId", taskId), logger.Error(err))
	}

	if err := client.DelegateTask(taskId, strconv.FormatInt(toUserId, 10)); err != nil {
		return err
	}

	logger.Info("Approval task delegated",
		zap.String("taskId", taskId),
		zap.Int64("fromUserId", fromUserId),
		zap.Int64("toUserId", toUserId),
	)

	return nil
}

func AddSignTask(taskId string, fromUserId int64, signUserId int64, remark string) error {
	initFlowableClientIfNeeded()
	client := GetFlowableClient()

	if err := client.AddComment(taskId, fmt.Sprintf("用户%d加签给用户%d: %s", fromUserId, signUserId, remark)); err != nil {
		logger.Warn("Add sign comment failed", zap.String("taskId", taskId), logger.Error(err))
	}

	if err := client.AddMultiInstance(taskId, strconv.FormatInt(signUserId, 10)); err != nil {
		return err
	}

	logger.Info("Approval task added sign",
		zap.String("taskId", taskId),
		zap.Int64("fromUserId", fromUserId),
		zap.Int64("signUserId", signUserId),
	)

	return nil
}

func GetApprovalTimeline(processInstanceId string) ([]*ApprovalNode, error) {
	initFlowableClientIfNeeded()
	client := GetFlowableClient()

	status, err := client.GetProcessStatus(processInstanceId)
	if err != nil {
		return nil, err
	}

	timeline := make([]*ApprovalNode, 0)

	nodeOrder := 0
	for _, task := range status.CompletedTasks {
		node := convertFlowableTaskToApprovalNode(task, nodeOrder)
		nodeOrder++
		timeline = append(timeline, node)
	}

	for _, task := range status.CurrentTasks {
		node := convertFlowableTaskToApprovalNode(task, nodeOrder)
		nodeOrder++
		timeline = append(timeline, node)
	}

	sortApprovalNodes(timeline)

	return timeline, nil
}

func convertFlowableTaskToApprovalNode(task *FlowableTask, order int) *ApprovalNode {
	node := &ApprovalNode{
		NodeName:   task.Name,
		SortOrder:  order + 1,
		CreateTime: task.CreateTime,
		Status:     constants.ApprovalStatusPending,
		StatusText: "待审批",
	}

	if task.Assignee != "" {
		if id, err := strconv.ParseInt(task.Assignee, 10, 64); err == nil {
			node.ApproverID = id
			node.Approver = getUserNameByID(id)
		}
	}

	if !task.DueDate.IsZero() {
		node.HandleTime = task.DueDate
		node.Status = constants.ApprovalStatusProcessing
		node.StatusText = "审批中"
	}

	return node
}

func getUserNameByID(userId int64) string {
	var user model.User
	result := database.GetDB().Model(&model.User{}).Where("id = ?", userId).First(&user)
	if result.Error != nil {
		return strconv.FormatInt(userId, 10)
	}
	return user.RealName
}

func sortApprovalNodes(nodes []*ApprovalNode) {
	for i := range nodes {
		for j := i + 1; j < len(nodes); j++ {
			if nodes[i].SortOrder > nodes[j].SortOrder {
				nodes[i], nodes[j] = nodes[j], nodes[i]
			}
		}
	}
}

func GetCurrentTaskIDByProcessInstance(processInstanceId string) (string, error) {
	initFlowableClientIfNeeded()
	client := GetFlowableClient()

	tasks, err := client.GetProcessInstanceTasks(processInstanceId)
	if err != nil {
		return "", err
	}

	if len(tasks) == 0 {
		return "", fmt.Errorf("no active task found for process instance: %s", processInstanceId)
	}

	return tasks[0].ID, nil
}

func GetProcessInstanceByBusinessKey(caseId int64) (string, error) {
	initFlowableClientIfNeeded()
	client := GetFlowableClient()

	businessKey := fmt.Sprintf("%s%d", BusinessKeyPrefix, caseId)
	path := "/runtime/process-instances?businessKey=" + businessKey

	var resp struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Total int `json:"total"`
	}

	if err := client.doRequest("GET", path, nil, &resp); err != nil {
		return "", err
	}

	if resp.Total == 0 || len(resp.Data) == 0 {
		return "", fmt.Errorf("no process instance found for case: %d", caseId)
	}

	return resp.Data[0].ID, nil
}

func SyncApprovalRecordsFromFlowable(caseId int64) error {
	initFlowableClientIfNeeded()

	processInstanceId, err := GetProcessInstanceByBusinessKey(caseId)
	if err != nil {
		logger.Warn("Get process instance failed, skip sync",
			zap.Int64("caseId", caseId),
			logger.Error(err),
		)
		return nil
	}

	timeline, err := GetApprovalTimeline(processInstanceId)
	if err != nil {
		return err
	}

	tx := database.GetDB().Begin()
	if err := tx.Where("case_id = ?", caseId).Delete(&model.ApprovalRecord{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, node := range timeline {
		record := &model.ApprovalRecord{
			CaseID:       caseId,
			NodeName:     node.NodeName,
			ApproverID:   node.ApproverID,
			ApproverName: node.Approver,
			Status:       node.Status,
			Remark:       node.Remark,
			SortOrder:    node.SortOrder,
		}
		if !node.HandleTime.IsZero() {
			record.ApprovedAt = &node.HandleTime
		}
		if !node.CreateTime.IsZero() {
			record.CreatedAt = node.CreateTime
		}
		if err := tx.Create(record).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	logger.Info("Approval records synced from flowable",
		zap.Int64("caseId", caseId),
		zap.String("processInstanceId", processInstanceId),
		zap.Int("count", len(timeline)),
	)

	return nil
}
