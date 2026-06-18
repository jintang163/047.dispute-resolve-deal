package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"go.uber.org/zap"
)

var (
	consumers      []rocketmq.PushConsumer
	consumerOnce   sync.Once
	consumerWg     sync.WaitGroup
	shutdownSignal chan struct{}
)

type CaseAssignMessage struct {
	CaseID         int64  `json:"caseId"`
	CaseNo         string `json:"caseNo"`
	CaseTitle      string `json:"caseTitle"`
	CaseType       string `json:"caseType"`
	CaseLevel      int    `json:"caseLevel"`
	MediatorID     int64  `json:"mediatorId"`
	MediatorName   string `json:"mediatorName"`
	MediatorPhone  string `json:"mediatorPhone"`
	ApplicantName  string `json:"applicantName"`
	ApplicantPhone string `json:"applicantPhone"`
	AssignTime     string `json:"assignTime"`
}

type ApprovalTodoMessage struct {
	CaseID        int64  `json:"caseId"`
	CaseNo        string `json:"caseNo"`
	CaseTitle     string `json:"caseTitle"`
	ApprovalNo    string `json:"approvalNo"`
	NodeName      string `json:"nodeName"`
	ApproverID    int64  `json:"approverId"`
	ApproverName  string `json:"approverName"`
	ApproverPhone string `json:"approverPhone"`
	SubmitterName string `json:"submitterName"`
	SubmitTime    string `json:"submitTime"`
	Deadline      string `json:"deadline"`
}

type AIProcessMessage struct {
	CaseID       int64  `json:"caseId"`
	CaseNo       string `json:"caseNo"`
	RecordID     int64  `json:"recordId"`
	RecordType   int    `json:"recordType"`
	ProcessType  int    `json:"processType"`
	Content      string `json:"content"`
	OperatorID   int64  `json:"operatorId"`
	OperatorName string `json:"operatorName"`
}

type TimeoutUpgradeMessage struct {
	CaseID           int64  `json:"caseId"`
	CaseNo           string `json:"caseNo"`
	CaseTitle        string `json:"caseTitle"`
	CurrentNode      string `json:"currentNode"`
	HandlerID        int64  `json:"handlerId"`
	HandlerName      string `json:"handlerName"`
	HandlerPhone     string `json:"handlerPhone"`
	OverdueHours     int    `json:"overdueHours"`
	TimeoutLevel     int    `json:"timeoutLevel"`
	NextEscalateTime string `json:"nextEscalateTime"`
	NextEscalateRole string `json:"nextEscalateRole"`
}

type SatisfactionEvalMessage struct {
	CaseID          int64  `json:"caseId"`
	CaseNo          string `json:"caseNo"`
	CaseTitle       string `json:"caseTitle"`
	UserID          int64  `json:"userId"`
	UserName        string `json:"userName"`
	UserPhone       string `json:"userPhone"`
	MediatorID      int64  `json:"mediatorId"`
	MediatorName    string `json:"mediatorName"`
	CloseTime       string `json:"closeTime"`
	MediationResult string `json:"mediationResult"`
	EvalUrl         string `json:"evalUrl"`
}

func StartConsumers() {
	consumerOnce.Do(func() {
		shutdownSignal = make(chan struct{})
		cfg := config.GlobalConfig

		if cfg == nil || cfg.RocketMQ.NameServer == nil {
			logger.Warn("RocketMQ config not found, skip starting consumers")
			return
		}

		logger.Info("Starting RocketMQ consumers...")

		go startCaseAssignConsumer(cfg)
		go startApprovalTodoConsumer(cfg)
		go startAIProcessConsumer(cfg)
		go startTimeoutUpgradeConsumer(cfg)
		go startSatisfactionEvalConsumer(cfg)

		go monitorShutdownSignal()

		logger.Info("All RocketMQ consumers started")
	})
}

func monitorShutdownSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("Received shutdown signal, stopping consumers...")
		close(shutdownSignal)
		ShutdownAllConsumers()
	case <-shutdownSignal:
	}
}

func ShutdownAllConsumers() {
	for _, c := range consumers {
		if c != nil {
			if err := c.Shutdown(); err != nil {
				logger.Error("Shutdown consumer failed", logger.Error(err))
			}
		}
	}
	consumerWg.Wait()
	logger.Info("All RocketMQ consumers shutdown completed")
}

func startCaseAssignConsumer(cfg *config.Config) {
	consumerWg.Add(1)
	defer consumerWg.Done()

	groupName := cfg.RocketMQ.GroupName + "_case_assign"
	cons := InitConsumer(cfg, groupName)
	consumers = append(consumers, cons)

	topic := constants.MQTopicCaseAssign

	err := cons.Subscribe(topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "*",
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			select {
			case <-shutdownSignal:
				return consumer.ConsumeRetryLater, nil
			default:
			}

			var caseMsg CaseAssignMessage
			if err := json.Unmarshal(msg.Body, &caseMsg); err != nil {
				logger.Error("Unmarshal case assign message failed",
					logger.Error(err),
					zap.String("msgId", msg.MsgId),
				)
				continue
			}

			logger.Info("Received case assign message",
				zap.Int64("caseId", caseMsg.CaseID),
				zap.String("caseNo", caseMsg.CaseNo),
				zap.Int64("mediatorId", caseMsg.MediatorID),
			)

			processCaseAssignNotification(&caseMsg)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		logger.Error("Subscribe case assign topic failed", logger.Error(err))
		return
	}

	if err := cons.Start(); err != nil {
		logger.Error("Start case assign consumer failed", logger.Error(err))
		return
	}

	logger.Info("Case assign consumer started", zap.String("topic", topic), zap.String("group", groupName))

	for {
		select {
		case <-shutdownSignal:
			logger.Info("Case assign consumer stopping")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func processCaseAssignNotification(msg *CaseAssignMessage) {
	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	channels := []int{1, 2, 3}
	channelNames := []string{"站内信", "短信", "微信"}

	title := fmt.Sprintf("【案件分派】您有新的纠纷案件待处理")
	content := fmt.Sprintf(
		"尊敬的%s您好：您被分派了新的纠纷案件，请及时处理。案件编号：%s，案件标题：%s，纠纷类型：%s，紧急程度：%s，报案人：%s，报案电话：%s，分派时间：%s",
		msg.MediatorName,
		msg.CaseNo,
		msg.CaseTitle,
		msg.CaseType,
		getCaseLevelName(msg.CaseLevel),
		msg.ApplicantName,
		msg.ApplicantPhone,
		msg.AssignTime,
	)

	params := map[string]interface{}{
		"mediatorName":   msg.MediatorName,
		"caseNo":         msg.CaseNo,
		"caseTitle":      msg.CaseTitle,
		"caseType":       msg.CaseType,
		"caseLevel":      getCaseLevelName(msg.CaseLevel),
		"applicantName":  msg.ApplicantName,
		"applicantPhone": msg.ApplicantPhone,
		"assignTime":     msg.AssignTime,
	}
	paramsJSON, _ := json.Marshal(params)

	for i, ch := range channels {
		record := &model.NotificationRecord{
			ReceiverID:   msg.MediatorID,
			ReceiverName: msg.MediatorName,
			TemplateID:   1,
			TemplateName: "案件分派通知",
			TemplateType: 1,
			Title:        title,
			Content:      content,
			Channel:      ch,
			Status:       0,
			Params:       string(paramsJSON),
			CreatedAt:    time.Now(),
		}
		_ = record
		_ = i
		_ = channelNames[i]

		insertSQL := `INSERT INTO notification_record (receiver_id, receiver_name, template_id, template_name, template_type, title, content, channel_type, send_status, params, case_id, case_no, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		if err := db.Exec(insertSQL,
			msg.MediatorID,
			msg.MediatorName,
			1,
			"案件分派通知",
			1,
			title,
			content,
			channelNames[i],
			1,
			string(paramsJSON),
			msg.CaseID,
			msg.CaseNo,
			time.Now(),
		).Error; err != nil {
			logger.Error("Insert case assign notification record failed",
				logger.Error(err),
				zap.Int64("caseId", msg.CaseID),
				zap.Int("channel", ch),
			)
		}
	}

	logger.Info("Case assign notifications sent",
		zap.Int64("caseId", msg.CaseID),
		zap.Int64("mediatorId", msg.MediatorID),
	)
}

func startApprovalTodoConsumer(cfg *config.Config) {
	consumerWg.Add(1)
	defer consumerWg.Done()

	groupName := cfg.RocketMQ.GroupName + "_approval_todo"
	cons := InitConsumer(cfg, groupName)
	consumers = append(consumers, cons)

	topic := constants.MQTopicApprovalNotify

	err := cons.Subscribe(topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "*",
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			select {
			case <-shutdownSignal:
				return consumer.ConsumeRetryLater, nil
			default:
			}

			var approvalMsg ApprovalTodoMessage
			if err := json.Unmarshal(msg.Body, &approvalMsg); err != nil {
				logger.Error("Unmarshal approval todo message failed",
					logger.Error(err),
					zap.String("msgId", msg.MsgId),
				)
				continue
			}

			logger.Info("Received approval todo message",
				zap.Int64("caseId", approvalMsg.CaseID),
				zap.String("caseNo", approvalMsg.CaseNo),
				zap.Int64("approverId", approvalMsg.ApproverID),
			)

			processApprovalTodoNotification(&approvalMsg)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		logger.Error("Subscribe approval todo topic failed", logger.Error(err))
		return
	}

	if err := cons.Start(); err != nil {
		logger.Error("Start approval todo consumer failed", logger.Error(err))
		return
	}

	logger.Info("Approval todo consumer started", zap.String("topic", topic), zap.String("group", groupName))

	for {
		select {
		case <-shutdownSignal:
			logger.Info("Approval todo consumer stopping")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func processApprovalTodoNotification(msg *ApprovalTodoMessage) {
	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	title := fmt.Sprintf("【审批待办】您有新的审批任务待处理")
	content := fmt.Sprintf(
		"尊敬的%s您好：您有新的审批任务待处理。审批编号：%s，关联案件：%s - %s，审批节点：%s，提交人：%s，提交时间：%s，请在%s前完成审批，超时将自动升级。",
		msg.ApproverName,
		msg.ApprovalNo,
		msg.CaseNo,
		msg.CaseTitle,
		msg.NodeName,
		msg.SubmitterName,
		msg.SubmitTime,
		msg.Deadline,
	)

	params := map[string]interface{}{
		"approverName":  msg.ApproverName,
		"approvalNo":    msg.ApprovalNo,
		"caseNo":        msg.CaseNo,
		"caseTitle":     msg.CaseTitle,
		"nodeName":      msg.NodeName,
		"submitterName": msg.SubmitterName,
		"submitTime":    msg.SubmitTime,
		"deadline":      msg.Deadline,
	}
	paramsJSON, _ := json.Marshal(params)

	insertSQL := `INSERT INTO notification_record (receiver_id, receiver_name, template_id, template_name, template_type, title, content, channel_type, send_status, params, case_id, case_no, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if err := db.Exec(insertSQL,
		msg.ApproverID,
		msg.ApproverName,
		2,
		"审批待办通知",
		2,
		title,
		content,
		"站内信",
		1,
		string(paramsJSON),
		msg.CaseID,
		msg.CaseNo,
		time.Now(),
	).Error; err != nil {
		logger.Error("Insert approval todo notification record failed",
			logger.Error(err),
			zap.Int64("caseId", msg.CaseID),
		)
	}

	logger.Info("Approval todo notification sent",
		zap.Int64("caseId", msg.CaseID),
		zap.Int64("approverId", msg.ApproverID),
	)
}

func startAIProcessConsumer(cfg *config.Config) {
	consumerWg.Add(1)
	defer consumerWg.Done()

	groupName := cfg.RocketMQ.GroupName + "_ai_process"
	cons := InitConsumer(cfg, groupName)
	consumers = append(consumers, cons)

	topic := constants.MQTopicAIProcess

	err := cons.Subscribe(topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "*",
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			select {
			case <-shutdownSignal:
				return consumer.ConsumeRetryLater, nil
			default:
			}

			var aiMsg AIProcessMessage
			if err := json.Unmarshal(msg.Body, &aiMsg); err != nil {
				logger.Error("Unmarshal AI process message failed",
					logger.Error(err),
					zap.String("msgId", msg.MsgId),
				)
				continue
			}

			logger.Info("Received AI process message",
				zap.Int64("caseId", aiMsg.CaseID),
				zap.String("caseNo", aiMsg.CaseNo),
				zap.Int("processType", aiMsg.ProcessType),
			)

			processAISummary(&aiMsg)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		logger.Error("Subscribe AI process topic failed", logger.Error(err))
		return
	}

	if err := cons.Start(); err != nil {
		logger.Error("Start AI process consumer failed", logger.Error(err))
		return
	}

	logger.Info("AI process consumer started", zap.String("topic", topic), zap.String("group", groupName))

	for {
		select {
		case <-shutdownSignal:
			logger.Info("AI process consumer stopping")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func processAISummary(msg *AIProcessMessage) {
	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	logger.Info("Starting AI summary generation",
		zap.Int64("caseId", msg.CaseID),
		zap.Int64("recordId", msg.RecordID),
		zap.Int("processType", msg.ProcessType),
	)

	summaryContent := generateMockAISummary(msg.Content, msg.ProcessType)
	suggestionContent := generateMockAISuggestion(msg.Content, msg.ProcessType)

	insertSQL := `INSERT INTO ai_mediation_summary (case_id, case_no, record_id, summary_type, original_content, ai_summary, ai_suggestion, risk_level, risk_points, ai_model, tokens_used, cost_time, is_approved, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if err := db.Exec(insertSQL,
		msg.CaseID,
		msg.CaseNo,
		msg.RecordID,
		msg.ProcessType,
		msg.Content,
		summaryContent,
		suggestionContent,
		1,
		"[]",
		"deepseek-mock",
		len(msg.Content)/4,
		1500,
		0,
		time.Now(),
	).Error; err != nil {
		logger.Error("Insert AI mediation summary failed",
			logger.Error(err),
			zap.Int64("caseId", msg.CaseID),
		)
		return
	}

	if msg.RecordID > 0 {
		updateSQL := `UPDATE dispute_mediation_record SET ai_summary = ? WHERE id = ?`
		if err := db.Exec(updateSQL, summaryContent, msg.RecordID).Error; err != nil {
			logger.Warn("Update mediation record ai_summary failed",
				logger.Error(err),
				zap.Int64("recordId", msg.RecordID),
			)
		}
	}

	aiRecordSQL := `INSERT INTO ai_assist_record (case_id, record_id, assist_type, ai_model, prompt, response, tokens_used, cost_time, operator_id, operator_name, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if err := db.Exec(aiRecordSQL,
		msg.CaseID,
		msg.RecordID,
		msg.ProcessType,
		"deepseek-mock",
		truncateString(msg.Content, 1000),
		summaryContent,
		len(msg.Content)/4,
		1500,
		msg.OperatorID,
		msg.OperatorName,
		time.Now(),
	).Error; err != nil {
		logger.Warn("Insert AI assist record failed",
			logger.Error(err),
			zap.Int64("caseId", msg.CaseID),
		)
	}

	logger.Info("AI summary generated and saved",
		zap.Int64("caseId", msg.CaseID),
		zap.Int64("recordId", msg.RecordID),
	)
}

func generateMockAISummary(content string, processType int) string {
	if len(content) == 0 {
		return "AI摘要：调解过程中双方就争议事项进行了充分沟通，调解员引导双方理性表达诉求，逐步缩小分歧。"
	}

	processTypeName := "调解"
	if processType == 2 {
		processTypeName = "审批"
	} else if processType == 3 {
		processTypeName = "风险评估"
	}

	runes := []rune(content)
	previewLen := len(runes)
	if previewLen > 50 {
		previewLen = 50
	}
	preview := string(runes[:previewLen])

	return fmt.Sprintf("AI%s摘要：本次%s围绕以下核心问题展开——%s...。经调解员梳理，双方争议焦点明确，建议后续围绕核心诉求进一步协商，争取达成互利共识。",
		processTypeName, processTypeName, preview)
}

func generateMockAISuggestion(content string, processType int) string {
	switch processType {
	case 1:
		return "调解建议：1. 建议双方保持冷静，避免情绪化表达；2. 核实相关证据材料，确保事实清楚；3. 可参考同类案例的处理方式；4. 必要时邀请第三方专业人士参与调解。"
	case 2:
		return "审批建议：1. 核查调解程序是否合规；2. 确认双方自愿性；3. 评估协议内容合法性与公平性；4. 关注后续履行风险。"
	case 3:
		return "风险提示：1. 关注双方情绪变化，防止矛盾激化；2. 对涉及金额较大的事项建议书面确认；3. 如调解不成，建议引导通过法律途径解决；4. 做好现场安全保障。"
	default:
		return "建议结合实际情况，灵活运用调解技巧，确保调解过程公平公正。"
	}
}

func startTimeoutUpgradeConsumer(cfg *config.Config) {
	consumerWg.Add(1)
	defer consumerWg.Done()

	groupName := cfg.RocketMQ.GroupName + "_timeout_upgrade"
	cons := InitConsumer(cfg, groupName)
	consumers = append(consumers, cons)

	topic := constants.MQTopicCaseUrge

	err := cons.Subscribe(topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "*",
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			select {
			case <-shutdownSignal:
				return consumer.ConsumeRetryLater, nil
			default:
			}

			var timeoutMsg TimeoutUpgradeMessage
			if err := json.Unmarshal(msg.Body, &timeoutMsg); err != nil {
				logger.Error("Unmarshal timeout upgrade message failed",
					logger.Error(err),
					zap.String("msgId", msg.MsgId),
				)
				continue
			}

			logger.Info("Received timeout upgrade message",
				zap.Int64("caseId", timeoutMsg.CaseID),
				zap.String("caseNo", timeoutMsg.CaseNo),
				zap.Int("timeoutLevel", timeoutMsg.TimeoutLevel),
			)

			processTimeoutUpgrade(&timeoutMsg)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		logger.Error("Subscribe timeout upgrade topic failed", logger.Error(err))
		return
	}

	if err := cons.Start(); err != nil {
		logger.Error("Start timeout upgrade consumer failed", logger.Error(err))
		return
	}

	logger.Info("Timeout upgrade consumer started", zap.String("topic", topic), zap.String("group", groupName))

	for {
		select {
		case <-shutdownSignal:
			logger.Info("Timeout upgrade consumer stopping")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func processTimeoutUpgrade(msg *TimeoutUpgradeMessage) {
	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	title := fmt.Sprintf("【超时提醒】案件处理超时，请尽快处理")
	content := fmt.Sprintf(
		"【超时提醒】案件编号%s已超时%d小时。案件标题：%s，当前节点：%s，当前处理人：%s，超时级别：%d级，请尽快处理，避免案件再次升级。系统将在%s自动升级到%s。",
		msg.CaseNo,
		msg.OverdueHours,
		msg.CaseTitle,
		msg.CurrentNode,
		msg.HandlerName,
		msg.TimeoutLevel,
		msg.NextEscalateTime,
		msg.NextEscalateRole,
	)

	params := map[string]interface{}{
		"caseNo":           msg.CaseNo,
		"caseTitle":        msg.CaseTitle,
		"currentNode":      msg.CurrentNode,
		"handlerName":      msg.HandlerName,
		"overdueHours":     msg.OverdueHours,
		"timeoutLevel":     msg.TimeoutLevel,
		"nextEscalateTime": msg.NextEscalateTime,
		"nextEscalateRole": msg.NextEscalateRole,
	}
	paramsJSON, _ := json.Marshal(params)

	channels := []string{"短信", "站内信"}
	for _, ch := range channels {
		insertSQL := `INSERT INTO notification_record (receiver_id, receiver_name, template_id, template_name, template_type, title, content, channel_type, send_status, params, case_id, case_no, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		if err := db.Exec(insertSQL,
			msg.HandlerID,
			msg.HandlerName,
			3,
			"超时提醒通知",
			3,
			title,
			content,
			ch,
			1,
			string(paramsJSON),
			msg.CaseID,
			msg.CaseNo,
			time.Now(),
		).Error; err != nil {
			logger.Error("Insert timeout notification record failed",
				logger.Error(err),
				zap.Int64("caseId", msg.CaseID),
				zap.String("channel", ch),
			)
		}
	}

	urgeType := constants.UrgeTypeSystem
	if msg.TimeoutLevel >= 3 {
		urgeType = constants.UrgeTypeEscalate
	}

	urgeSQL := `INSERT INTO dispute_urge (case_id, urge_type, urge_content, operator_id, operator_name, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	if err := db.Exec(urgeSQL,
		msg.CaseID,
		urgeType,
		content,
		0,
		"系统自动",
		time.Now(),
	).Error; err != nil {
		logger.Warn("Insert dispute urge record failed",
			logger.Error(err),
			zap.Int64("caseId", msg.CaseID),
		)
	}

	historySQL := `INSERT INTO dispute_case_history (case_id, case_no, operation_type, operation_detail, operator_id, operator_name, operator_role, remark, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	historyDetail := fmt.Sprintf(`{"timeoutLevel":%d,"overdueHours":%d,"currentNode":"%s","handlerId":%d}`,
		msg.TimeoutLevel, msg.OverdueHours, msg.CurrentNode, msg.HandlerID)
	if err := db.Exec(historySQL,
		msg.CaseID,
		msg.CaseNo,
		"TIMEOUT_UPGRADE",
		historyDetail,
		0,
		"系统自动",
		0,
		fmt.Sprintf("超时升级：%d级，超时%d小时", msg.TimeoutLevel, msg.OverdueHours),
		time.Now(),
	).Error; err != nil {
		logger.Warn("Insert case history record failed",
			logger.Error(err),
			zap.Int64("caseId", msg.CaseID),
		)
	}

	if msg.TimeoutLevel >= 2 {
		updateSQL := `UPDATE dispute_case SET escalate_level = ?, escalate_time = ?, urgency_count = urgency_count + 1 WHERE id = ?`
		if err := db.Exec(updateSQL, msg.TimeoutLevel, time.Now(), msg.CaseID).Error; err != nil {
			logger.Warn("Update case escalate level failed",
				logger.Error(err),
				zap.Int64("caseId", msg.CaseID),
			)
		}
	}

	logger.Info("Timeout upgrade processed",
		zap.Int64("caseId", msg.CaseID),
		zap.Int("timeoutLevel", msg.TimeoutLevel),
		zap.Int("overdueHours", msg.OverdueHours),
	)
}

func startSatisfactionEvalConsumer(cfg *config.Config) {
	consumerWg.Add(1)
	defer consumerWg.Done()

	groupName := cfg.RocketMQ.GroupName + "_satisfaction_eval"
	cons := InitConsumer(cfg, groupName)
	consumers = append(consumers, cons)

	topic := constants.MQTopicNotification

	err := cons.Subscribe(topic, consumer.MessageSelector{
		Type:       consumer.TAG,
		Expression: "satisfaction",
	}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			select {
			case <-shutdownSignal:
				return consumer.ConsumeRetryLater, nil
			default:
			}

			var evalMsg SatisfactionEvalMessage
			if err := json.Unmarshal(msg.Body, &evalMsg); err != nil {
				logger.Error("Unmarshal satisfaction eval message failed",
					logger.Error(err),
					zap.String("msgId", msg.MsgId),
				)
				continue
			}

			logger.Info("Received satisfaction eval message",
				zap.Int64("caseId", evalMsg.CaseID),
				zap.String("caseNo", evalMsg.CaseNo),
				zap.Int64("userId", evalMsg.UserID),
			)

			processSatisfactionEvalNotification(&evalMsg)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		logger.Error("Subscribe satisfaction eval topic failed", logger.Error(err))
		return
	}

	if err := cons.Start(); err != nil {
		logger.Error("Start satisfaction eval consumer failed", logger.Error(err))
		return
	}

	logger.Info("Satisfaction eval consumer started", zap.String("topic", topic), zap.String("group", groupName))

	for {
		select {
		case <-shutdownSignal:
			logger.Info("Satisfaction eval consumer stopping")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func processSatisfactionEvalNotification(msg *SatisfactionEvalMessage) {
	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	title := fmt.Sprintf("【服务评价】诚邀您对本次调解服务进行评价")
	content := fmt.Sprintf(
		"尊敬的%s您好：您的纠纷案件（编号：%s）已结案。案件标题：%s，调解员：%s，结案时间：%s，调解结果：%s。请您对本次调解服务进行满意度评价，您的反馈是我们改进服务的动力。点击链接参与评价：%s，本邀请72小时内有效。",
		msg.UserName,
		msg.CaseNo,
		msg.CaseTitle,
		msg.MediatorName,
		msg.CloseTime,
		msg.MediationResult,
		msg.EvalUrl,
	)

	params := map[string]interface{}{
		"userName":        msg.UserName,
		"caseNo":          msg.CaseNo,
		"caseTitle":       msg.CaseTitle,
		"mediatorName":    msg.MediatorName,
		"closeTime":       msg.CloseTime,
		"mediationResult": msg.MediationResult,
		"evalUrl":         msg.EvalUrl,
	}
	paramsJSON, _ := json.Marshal(params)

	channels := []string{"站内信", "短信", "微信"}
	for _, ch := range channels {
		insertSQL := `INSERT INTO notification_record (receiver_id, receiver_name, template_id, template_name, template_type, title, content, channel_type, send_status, params, case_id, case_no, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		if err := db.Exec(insertSQL,
			msg.UserID,
			msg.UserName,
			4,
			"满意度评价邀请",
			4,
			title,
			content,
			ch,
			1,
			string(paramsJSON),
			msg.CaseID,
			msg.CaseNo,
			time.Now(),
		).Error; err != nil {
			logger.Error("Insert satisfaction eval notification record failed",
				logger.Error(err),
				zap.Int64("caseId", msg.CaseID),
				zap.String("channel", ch),
			)
		}
	}

	logger.Info("Satisfaction eval notification sent",
		zap.Int64("caseId", msg.CaseID),
		zap.Int64("userId", msg.UserID),
	)
}

func getCaseLevelName(level int) string {
	switch level {
	case 1:
		return "特急"
	case 2:
		return "紧急"
	case 3:
		return "一般"
	case 4:
		return "普通"
	default:
		return "普通"
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

func SendCaseAssignMessage(msg *CaseAssignMessage) error {
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send case assign message async failed", logger.Error(err))
		}
	}
	body, _ := json.Marshal(msg)
	return SendAsyncMessage(constants.MQTopicCaseAssign, body, callback)
}

func SendApprovalTodoMessage(msg *ApprovalTodoMessage) error {
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send approval todo message async failed", logger.Error(err))
		}
	}
	body, _ := json.Marshal(msg)
	return SendAsyncMessage(constants.MQTopicApprovalNotify, body, callback)
}

func SendAIProcessMessage(msg *AIProcessMessage) error {
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send AI process message async failed", logger.Error(err))
		}
	}
	body, _ := json.Marshal(msg)
	return SendAsyncMessage(constants.MQTopicAIProcess, body, callback)
}

func SendTimeoutUpgradeMessage(msg *TimeoutUpgradeMessage, delayLevel int) error {
	body, _ := json.Marshal(msg)
	return SendDelayMessage(constants.MQTopicCaseUrge, body, delayLevel)
}

func SendSatisfactionEvalMessage(msg *SatisfactionEvalMessage, delayLevel int) error {
	body, _ := json.Marshal(msg)
	if delayLevel > 0 {
		return SendDelayMessage(constants.MQTopicNotification, body, delayLevel, "satisfaction")
	}
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send satisfaction eval message async failed", logger.Error(err))
		}
	}
	return SendAsyncMessage(constants.MQTopicNotification, body, callback, "satisfaction")
}
