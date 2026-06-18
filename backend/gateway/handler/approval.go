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
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type SubmitApprovalRequest struct {
	CaseID  int64  `json:"caseId" binding:"required"`
	DefCode string `json:"defCode"`
	Remark  string `json:"remark"`
}

type ProcessApprovalRequest struct {
	Action        int32  `json:"action" binding:"required"`
	Opinion       string `json:"opinion"`
	NextApproverID int64 `json:"nextApproverId"`
}

func SubmitApproval(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var caseData struct {
		CaseNo          string `gorm:"column:case_no"`
		Title           string `gorm:"column:title"`
		Status          int32  `gorm:"column:status"`
		MediatorID      int64  `gorm:"column:mediator_id"`
		MediatorName    string `gorm:"column:mediator_name"`
		OrganizationID  int64  `gorm:"column:organization_id"`
	}

	database.GetDB().Table("dispute_case").
		Where("id = ?", caseID).
		First(&caseData)

	if caseData.Status != constants.CaseStatusMediating {
		c.JSON(http.StatusBadRequest, response.BadRequest("只有调解中状态的案件才能提交审批"))
		return
	}

	if caseData.MediatorID != userInfo.UserID && userInfo.Role > constants.RoleMediator {
		c.JSON(http.StatusForbidden, response.Forbidden("只有案件调解员才能提交审批"))
		return
	}

	defCode := "DEFAULT_APPROVAL"
	var def struct {
		ID             int64  `gorm:"column:id"`
		DefCode        string `gorm:"column:def_code"`
		DefName        string `gorm:"column:def_name"`
		ApprovalNodes  string `gorm:"column:approval_nodes"`
		TimeoutConfig  string `gorm:"column:timeout_config"`
	}
	database.GetDB().Table("workflow_approval_definition").
		Where("def_code = ? AND status = 1", defCode).
		First(&def)

	instanceNo := utils.GenerateApprovalNo()
	instanceID := utils.GenerateID()

	var approvalNodes []map[string]interface{}
	json.Unmarshal([]byte(def.ApprovalNodes), &approvalNodes)

	currentNode := approvalNodes[1]
	approverRole := currentNode["approverRole"].(string)

	var approver struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:real_name"`
	}

	roleMap := map[string]int32{
		"mediator": constants.RoleMediator,
		"leader":   constants.RoleLeader,
		"director": constants.RoleDirector,
	}

	database.GetDB().Table("sys_user").
		Where("role = ? AND organization_id = ? AND status = 1", 
			roleMap[approverRole], caseData.OrganizationID).
		Order("id ASC").
		Limit(1).
		First(&approver)

	timeout := int(currentNode["timeout"].(float64))

	tx := database.GetDB().Begin()

	instance := map[string]interface{}{
		"id":                instanceID,
		"instance_no":       instanceNo,
		"case_id":           caseID,
		"case_no":           caseData.CaseNo,
		"def_id":            def.ID,
		"def_code":          def.DefCode,
		"current_node_code": currentNode["code"].(string),
		"current_node_name": currentNode["name"].(string),
		"approver_id":       approver.ID,
		"approver_name":     approver.Name,
		"status":            constants.ApprovalStatusPending,
		"submit_user_id":    userInfo.UserID,
		"submit_user_name":  userInfo.RealName,
		"submit_time":       time.Now(),
		"total_nodes":       len(approvalNodes),
		"current_node_index": 1,
		"timeout_time":      time.Now().Add(time.Duration(timeout) * time.Second),
	}
	tx.Table("workflow_approval_instance").Create(instance)

	record := map[string]interface{}{
		"instance_id":     instanceID,
		"case_id":         caseID,
		"case_no":         caseData.CaseNo,
		"node_code":       "MEDIATOR_SUBMIT",
		"node_name":       "调解员提交",
		"node_type":       1,
		"approver_id":     userInfo.UserID,
		"approver_name":   userInfo.RealName,
		"approval_action": constants.ApprovalActionPass,
		"approval_opinion": "提交审批",
		"approval_time":   time.Now(),
	}
	tx.Table("workflow_approval_record").Create(record)

	tx.Table("dispute_case").
		Where("id = ?", caseID).
		Updates(map[string]interface{}{
			"status": constants.CaseStatusApproving,
		})

	history := map[string]interface{}{
		"case_id":       caseID,
		"case_no":       caseData.CaseNo,
		"operation_type": "SUBMIT_APPROVAL",
		"operation_detail": fmt.Sprintf("提交审批，当前节点: %s，审批人: %s", 
			currentNode["name"].(string), approver.Name),
		"operator_id":   userInfo.UserID,
		"operator_name": userInfo.RealName,
		"old_status":    constants.CaseStatusMediating,
		"new_status":    constants.CaseStatusApproving,
	}
	tx.Table("dispute_case_history").Create(history)

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	go func() {
		msg := map[string]interface{}{
			"instanceId":   instanceID,
			"instanceNo":   instanceNo,
			"caseId":       caseID,
			"caseNo":       caseData.CaseNo,
			"caseTitle":    caseData.Title,
			"approverId":   approver.ID,
			"approverName": approver.Name,
			"nodeName":     currentNode["name"].(string),
			"submitBy":     userInfo.RealName,
			"submitTime":   time.Now().Format(time.RFC3339),
		}
		mq.SendMessage(constants.MQTopicApprovalNotify, msg)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"instanceId": instanceID,
		"instanceNo": instanceNo,
	}, "提交审批成功"))
}

func ApproveApproval(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var req ProcessApprovalRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	var instance struct {
		ID                int64  `gorm:"column:id"`
		InstanceNo        string `gorm:"column:instance_no"`
		CaseNo            string `gorm:"column:case_no"`
		DefID             int64  `gorm:"column:def_id"`
		CurrentNodeCode   string `gorm:"column:current_node_code"`
		CurrentNodeName   string `gorm:"column:current_node_name"`
		ApproverID        int64  `gorm:"column:approver_id"`
		Status            int32  `gorm:"column:status"`
		TotalNodes        int    `gorm:"column:total_nodes"`
		CurrentNodeIndex  int    `gorm:"column:current_node_index"`
		SubmitUserID      int64  `gorm:"column:submit_user_id"`
		SubmitUserName    string `gorm:"column:submit_user_name"`
	}

	database.GetDB().Table("workflow_approval_instance").
		Where("case_id = ? AND status = ?", caseID, constants.ApprovalStatusPending).
		Order("id DESC").
		First(&instance)

	if instance.ID == 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该案件没有待处理的审批"))
		return
	}

	if instance.ApproverID != userInfo.UserID {
		c.JSON(http.StatusForbidden, response.Forbidden("您不是当前审批人，无权处理"))
		return
	}

	var def struct {
		ApprovalNodes string `gorm:"column:approval_nodes"`
	}
	database.GetDB().Table("workflow_approval_definition").
		Select("approval_nodes").
		Where("id = ?", instance.DefID).
		First(&def)

	var approvalNodes []map[string]interface{}
	json.Unmarshal([]byte(def.ApprovalNodes), &approvalNodes)

	isLastNode := instance.CurrentNodeIndex >= instance.TotalNodes-1

	tx := database.GetDB().Begin()

	record := map[string]interface{}{
		"instance_id":     instance.ID,
		"case_id":         caseID,
		"case_no":         instance.CaseNo,
		"node_code":       instance.CurrentNodeCode,
		"node_name":       instance.CurrentNodeName,
		"node_type":       1,
		"approver_id":     userInfo.UserID,
		"approver_name":   userInfo.RealName,
		"approval_action": req.Action,
		"approval_opinion": req.Opinion,
		"approval_time":   time.Now(),
	}
	tx.Table("workflow_approval_record").Create(record)

	if req.Action == constants.ApprovalActionPass {
		if isLastNode {
			tx.Table("workflow_approval_instance").
				Where("id = ?", instance.ID).
				Updates(map[string]interface{}{
					"status":     constants.ApprovalStatusPassed,
					"end_time":   time.Now(),
				})

			tx.Table("dispute_case").
				Where("id = ?", caseID).
				Updates(map[string]interface{}{
					"status":          constants.CaseStatusClosed,
					"close_reason":    "审批通过，正常结案",
					"close_user_id":   userInfo.UserID,
					"close_time":      time.Now(),
					"mediation_end_time": time.Now(),
				})

			history := map[string]interface{}{
				"case_id":       caseID,
				"case_no":       instance.CaseNo,
				"operation_type": "APPROVAL_PASS",
				"operation_detail": fmt.Sprintf("审批通过: %s", req.Opinion),
				"operator_id":   userInfo.UserID,
				"operator_name": userInfo.RealName,
				"old_status":    constants.CaseStatusApproving,
				"new_status":    constants.CaseStatusClosed,
			}
			tx.Table("dispute_case_history").Create(history)

			go func() {
				msg := map[string]interface{}{
					"caseId":     caseID,
					"caseNo":     instance.CaseNo,
					"action":     "COMPLETE",
					"result":     "审批通过，案件已结案",
					"notifyUserIds": []int64{instance.SubmitUserID},
				}
				mq.SendMessage(constants.MQTopicApprovalNotify, msg)
			}()
		} else {
			nextNode := approvalNodes[instance.CurrentNodeIndex+1]
			nextRole := nextNode["approverRole"].(string)

			var nextApprover struct {
				ID   int64  `gorm:"column:id"`
				Name string `gorm:"column:real_name"`
			}

			roleMap := map[string]int32{
				"mediator": constants.RoleMediator,
				"leader":   constants.RoleLeader,
				"director": constants.RoleDirector,
			}

			var orgID int64
			database.GetDB().Table("dispute_case").
				Select("organization_id").
				Where("id = ?", caseID).
				Scan(&orgID)

			database.GetDB().Table("sys_user").
				Where("role = ? AND organization_id = ? AND status = 1",
					roleMap[nextRole], orgID).
				Order("id ASC").
				Limit(1).
				First(&nextApprover)

			timeout := int(nextNode["timeout"].(float64))

			tx.Table("workflow_approval_instance").
				Where("id = ?", instance.ID).
				Updates(map[string]interface{}{
					"current_node_code":  nextNode["code"].(string),
					"current_node_name":  nextNode["name"].(string),
					"approver_id":        nextApprover.ID,
					"approver_name":      nextApprover.Name,
					"current_node_index": instance.CurrentNodeIndex + 1,
					"timeout_time":       time.Now().Add(time.Duration(timeout) * time.Second),
				})

			history := map[string]interface{}{
				"case_id":       caseID,
				"case_no":       instance.CaseNo,
				"operation_type": "APPROVAL_PASS",
				"operation_detail": fmt.Sprintf("节点[%s]审批通过，流转至下一节点[%s]，审批人: %s",
					instance.CurrentNodeName, nextNode["name"].(string), nextApprover.Name),
				"operator_id":   userInfo.UserID,
				"operator_name": userInfo.RealName,
			}
			tx.Table("dispute_case_history").Create(history)

			go func() {
				msg := map[string]interface{}{
					"caseId":         caseID,
					"caseNo":         instance.CaseNo,
					"action":         "NEXT_NODE",
					"currentNode":    nextNode["name"].(string),
					"approverId":     nextApprover.ID,
					"approverName":   nextApprover.Name,
					"prevApprover":   userInfo.RealName,
				}
				mq.SendMessage(constants.MQTopicApprovalNotify, msg)
			}()
		}
	} else if req.Action == constants.ApprovalActionReject {
		tx.Table("workflow_approval_instance").
			Where("id = ?", instance.ID).
			Updates(map[string]interface{}{
				"status":   constants.ApprovalStatusRejected,
				"end_time": time.Now(),
			})

		tx.Table("dispute_case").
			Where("id = ?", caseID).
			Update("status", constants.CaseStatusMediating)

		history := map[string]interface{}{
			"case_id":       caseID,
			"case_no":       instance.CaseNo,
			"operation_type": "APPROVAL_REJECT",
			"operation_detail": fmt.Sprintf("审批驳回: %s", req.Opinion),
			"operator_id":   userInfo.UserID,
			"operator_name": userInfo.RealName,
			"old_status":    constants.CaseStatusApproving,
			"new_status":    constants.CaseStatusMediating,
		}
		tx.Table("dispute_case_history").Create(history)

		go func() {
			msg := map[string]interface{}{
				"caseId":     caseID,
				"caseNo":     instance.CaseNo,
				"action":     "REJECT",
				"result":     fmt.Sprintf("审批被驳回: %s", req.Opinion),
				"notifyUserIds": []int64{instance.SubmitUserID},
			}
			mq.SendMessage(constants.MQTopicApprovalNotify, msg)
		}()
	} else if req.Action == constants.ApprovalActionReturn {
		tx.Table("workflow_approval_instance").
			Where("id = ?", instance.ID).
			Updates(map[string]interface{}{
				"current_node_index": 0,
				"current_node_code":  "MEDIATOR_SUBMIT",
				"current_node_name":  "调解员修改",
				"approver_id":        instance.SubmitUserID,
				"approver_name":      instance.SubmitUserName,
			})

		tx.Table("dispute_case").
			Where("id = ?", caseID).
			Update("status", constants.CaseStatusMediating)

		history := map[string]interface{}{
			"case_id":       caseID,
			"case_no":       instance.CaseNo,
			"operation_type": "APPROVAL_RETURN",
			"operation_detail": fmt.Sprintf("退回修改: %s", req.Opinion),
			"operator_id":   userInfo.UserID,
			"operator_name": userInfo.RealName,
			"old_status":    constants.CaseStatusApproving,
			"new_status":    constants.CaseStatusMediating,
		}
		tx.Table("dispute_case_history").Create(history)
	} else if req.Action == constants.ApprovalActionTransfer {
		if req.NextApproverID == 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, response.BadRequest("转审时必须指定下一审批人"))
			return
		}

		var nextApprover struct {
			RealName string `gorm:"column:real_name"`
			Role     int32  `gorm:"column:role"`
		}
		database.GetDB().Table("sys_user").
			Select("real_name, role").
			Where("id = ?", req.NextApproverID).
			First(&nextApprover)

		tx.Table("workflow_approval_instance").
			Where("id = ?", instance.ID).
			Updates(map[string]interface{}{
				"approver_id":    req.NextApproverID,
				"approver_name":  nextApprover.RealName,
			})

		record := map[string]interface{}{
			"instance_id":     instance.ID,
			"case_id":         caseID,
			"case_no":         instance.CaseNo,
			"node_code":       instance.CurrentNodeCode,
			"node_name":       instance.CurrentNodeName,
			"node_type":       3,
			"approver_id":     userInfo.UserID,
			"approver_name":   userInfo.RealName,
			"approval_action": constants.ApprovalActionTransfer,
			"approval_opinion": fmt.Sprintf("转审给%s: %s", nextApprover.RealName, req.Opinion),
			"approval_time":   time.Now(),
			"next_approver_id": req.NextApproverID,
		}
		tx.Table("workflow_approval_record").Create(record)

		history := map[string]interface{}{
			"case_id":       caseID,
			"case_no":       instance.CaseNo,
			"operation_type": "APPROVAL_TRANSFER",
			"operation_detail": fmt.Sprintf("转审给%s: %s", nextApprover.RealName, req.Opinion),
			"operator_id":   userInfo.UserID,
			"operator_name": userInfo.RealName,
		}
		tx.Table("dispute_case_history").Create(history)

		go func() {
			msg := map[string]interface{}{
				"caseId":         caseID,
				"caseNo":         instance.CaseNo,
				"action":         "TRANSFER",
				"approverId":     req.NextApproverID,
				"approverName":   nextApprover.RealName,
				"transferFrom":   userInfo.RealName,
				"opinion":        req.Opinion,
			}
			mq.SendMessage(constants.MQTopicApprovalNotify, msg)
		}()
	} else if req.Action == constants.ApprovalActionAddSign {
		if req.NextApproverID == 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, response.BadRequest("加签时必须指定加签人"))
			return
		}

		var signApprover struct {
			RealName string `gorm:"column:real_name"`
		}
		database.GetDB().Table("sys_user").
			Select("real_name").
			Where("id = ?", req.NextApproverID).
			First(&signApprover)

		record := map[string]interface{}{
			"instance_id":     instance.ID,
			"case_id":         caseID,
			"case_no":         instance.CaseNo,
			"node_code":       instance.CurrentNodeCode + "_ADD_SIGN",
			"node_name":       "加签审批",
			"node_type":       3,
			"approver_id":     req.NextApproverID,
			"approver_name":   signApprover.RealName,
			"approval_action": 0,
			"approval_opinion": "",
			"approval_time":   time.Now(),
		}
		tx.Table("workflow_approval_record").Create(record)

		history := map[string]interface{}{
			"case_id":       caseID,
			"case_no":       instance.CaseNo,
			"operation_type": "APPROVAL_ADD_SIGN",
			"operation_detail": fmt.Sprintf("加签给%s: %s", signApprover.RealName, req.Opinion),
			"operator_id":   userInfo.UserID,
			"operator_name": userInfo.RealName,
		}
		tx.Table("dispute_case_history").Create(history)
	}

	tx.Commit()

	cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
	cache.Del(ctx, cacheKey)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "审批处理成功"))
}

func GetApprovalProgress(ctx context.Context, c *app.RequestContext) {
	caseID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var instance map[string]interface{}
	database.GetDB().Table("workflow_approval_instance").
		Where("case_id = ?", caseID).
		Order("id DESC").
		Limit(1).
		Find(&instance)

	if instance == nil {
		c.JSON(http.StatusOK, response.Success(map[string]interface{}{
			"hasApproval": false,
			"timeline":    []interface{}{},
		}))
		return
	}

	var records []map[string]interface{}
	database.GetDB().Table("workflow_approval_record").
		Where("instance_id = ?", instance["id"]).
		Order("approval_time ASC").
		Find(&records)

	timeline := make([]map[string]interface{}, 0)
	for _, record := range records {
		action := int(record["approval_action"].(int32))
		item := map[string]interface{}{
			"nodeName":   record["node_name"],
			"approver":   record["approver_name"],
			"action":     action,
			"actionName": constants.ApprovalActionMap[action],
			"opinion":    record["approval_opinion"],
			"time":       record["approval_time"],
			"nodeType":   record["node_type"],
		}
		timeline = append(timeline, item)
	}

	if instance["status"].(int32) == constants.ApprovalStatusPending {
		timeline = append(timeline, map[string]interface{}{
			"nodeName":   instance["current_node_name"],
			"approver":   instance["approver_name"],
			"action":     0,
			"actionName": "待处理",
			"opinion":    "",
			"time":       nil,
			"nodeType":   1,
			"isCurrent":  true,
		})
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"hasApproval": true,
		"instance":    instance,
		"timeline":    timeline,
	}))
}

func GetApprovalTodoList(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	var req common.BaseQuery
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("workflow_approval_instance ai").
		Select("ai.*, dc.title, dc.case_no, dt.type_name").
		Joins("LEFT JOIN dispute_case dc ON ai.case_id = dc.id").
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Where("ai.approver_id = ? AND ai.status = ?", userInfo.UserID, constants.ApprovalStatusPending)

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("ai.created_at DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	for _, item := range list {
		if timeout, ok := item["timeout_time"]; ok && timeout != nil {
			t := timeout.(time.Time)
			if time.Now().After(t) {
				item["isOverdue"] = true
				item["overdueHours"] = int(time.Since(t).Hours())
			} else {
				item["isOverdue"] = false
				item["remainingHours"] = int(time.Until(t).Hours())
			}
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetApprovalDoneList(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	var req common.BaseQuery
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("workflow_approval_record ar").
		Select("ar.*, ai.instance_no, ai.status as instance_status, dc.title, dc.case_no").
		Joins("LEFT JOIN workflow_approval_instance ai ON ar.instance_id = ai.id").
		Joins("LEFT JOIN dispute_case dc ON ai.case_id = dc.id").
		Where("ar.approver_id = ? AND ar.approval_action > 0", userInfo.UserID)

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("ar.approval_time DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	for _, item := range list {
		if action, ok := item["approval_action"].(int32); ok {
			item["action_name"] = constants.ApprovalActionMap[int(action)]
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func RejectApproval(ctx context.Context, c *app.RequestContext) {
	var req ProcessApprovalRequest
	req.Action = constants.ApprovalActionReject
	ApproveApproval(ctx, c)
}

func ReturnApproval(ctx context.Context, c *app.RequestContext) {
	var req ProcessApprovalRequest
	req.Action = constants.ApprovalActionReturn
	ApproveApproval(ctx, c)
}

func AddSignApproval(ctx context.Context, c *app.RequestContext) {
	var req ProcessApprovalRequest
	req.Action = constants.ApprovalActionAddSign
	ApproveApproval(ctx, c)
}

func TransferApproval(ctx context.Context, c *app.RequestContext) {
	var req ProcessApprovalRequest
	req.Action = constants.ApprovalActionTransfer
	ApproveApproval(ctx, c)
}
