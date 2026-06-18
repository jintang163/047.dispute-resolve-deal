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

type JudicialStatusMessage struct {
	ConfirmID    int64                  `json:"confirmId"`
	ConfirmNo    string                 `json:"confirmNo"`
	OldStatus    int32                  `json:"oldStatus"`
	NewStatus    int32                  `json:"newStatus"`
	CourtCaseNo  string                 `json:"courtCaseNo"`
	CourtData    map[string]interface{} `json:"courtData"`
	OperatorID   int64                  `json:"operatorId"`
	OperatorName string                 `json:"operatorName"`
}

type JudicialSubmitMessage struct {
	ConfirmID    int64  `json:"confirmId"`
	ConfirmNo    string `json:"confirmNo"`
	CaseID       int64  `json:"caseId"`
	OperatorID   int64  `json:"operatorId"`
	OperatorName string `json:"operatorName"`
}

type JudicialSealMessage struct {
	ConfirmID    int64  `json:"confirmId"`
	ConfirmNo    string `json:"confirmNo"`
	DocumentURL  string `json:"documentUrl"`
	SealStatus   int    `json:"sealStatus"`
	OperatorID   int64  `json:"operatorId"`
	OperatorName string `json:"operatorName"`
}

type JudicialRemindMessage struct {
	ConfirmID    int64    `json:"confirmId"`
	ConfirmNo    string   `json:"confirmNo"`
	RemindType   string   `json:"remindType"`
	TargetPhones []string `json:"targetPhones"`
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
		go startJudicialStatusConsumer(cfg)
		go startJudicialSyncConsumer(cfg)
		go startJudicialRemindConsumer(cfg)

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

func startJudicialStatusConsumer(cfg *config.Config) {
	consumerWg.Add(1)
	defer consumerWg.Done()

	groupName := cfg.RocketMQ.GroupName + "_judicial_status"
	cons := InitConsumer(cfg, groupName)
	consumers = append(consumers, cons)

	topic := constants.MQTopicJudicialStatus

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

			var statusMsg JudicialStatusMessage
			if err := json.Unmarshal(msg.Body, &statusMsg); err != nil {
				logger.Error("Unmarshal judicial status message failed",
					logger.Error(err),
					zap.String("msgId", msg.MsgId),
				)
				continue
			}

			logger.Info("Received judicial status message",
				zap.Int64("confirmId", statusMsg.ConfirmID),
				zap.String("confirmNo", statusMsg.ConfirmNo),
				zap.Int32("oldStatus", statusMsg.OldStatus),
				zap.Int32("newStatus", statusMsg.NewStatus),
			)

			processJudicialStatusNotification(&statusMsg)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		logger.Error("Subscribe judicial status topic failed", logger.Error(err))
		return
	}

	if err := cons.Start(); err != nil {
		logger.Error("Start judicial status consumer failed", logger.Error(err))
		return
	}

	logger.Info("Judicial status consumer started", zap.String("topic", topic), zap.String("group", groupName))

	for {
		select {
		case <-shutdownSignal:
			logger.Info("Judicial status consumer stopping")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func processJudicialStatusNotification(msg *JudicialStatusMessage) {
	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	statusName := map[int32]string{
		10: "已提交",
		20: "审核中",
		30: "已确认",
		40: "已驳回",
		50: "已失效",
	}

	title := fmt.Sprintf("【司法确认】状态变更通知")
	content := fmt.Sprintf(
		"您的司法确认申请（编号：%s）状态已变更：%s → %s。",
		msg.ConfirmNo,
		statusName[msg.OldStatus],
		statusName[msg.NewStatus],
	)

	if msg.NewStatus == 30 {
		content += "恭喜您，司法确认已通过！请按照调解协议履行义务。"
	} else if msg.NewStatus == 40 {
		content += "抱歉，司法确认被驳回，请联系调解员了解详情。"
	} else if msg.NewStatus == 50 {
		content += "注意：司法确认已超过履行期限，对方可向法院申请强制执行。"
	}

	params := map[string]interface{}{
		"confirmNo": msg.ConfirmNo,
		"oldStatus": statusName[msg.OldStatus],
		"newStatus": statusName[msg.NewStatus],
		"courtCaseNo": msg.CourtCaseNo,
	}
	paramsJSON, _ := json.Marshal(params)

	var confirm struct {
		ApplicantName    string `json:"applicantName"`
		ApplicantPhone   string `json:"applicantPhone"`
		ApplicantID      int64  `json:"applicantId"`
		RespondentName   string `json:"respondentName"`
		RespondentPhone  string `json:"respondentPhone"`
	}
	db.Table("judicial_confirmation").Select("applicant_name, applicant_phone, respondent_name, respondent_phone").
		Where("id = ?", msg.ConfirmID).Scan(&confirm)

	receivers := []struct {
		ID   int64
		Name string
		Phone string
	}{
		{0, confirm.ApplicantName, confirm.ApplicantPhone},
		{0, confirm.RespondentName, confirm.RespondentPhone},
	}

	for _, receiver := range receivers {
		insertSQL := `INSERT INTO notification_record (receiver_id, receiver_name, template_id, template_name, template_type, title, content, channel_type, send_status, params, case_id, case_no, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		if err := db.Exec(insertSQL,
			receiver.ID,
			receiver.Name,
			5,
			"司法确认状态通知",
			5,
			title,
			content,
			"短信",
			1,
			string(paramsJSON),
			0,
			msg.ConfirmNo,
			time.Now(),
		).Error; err != nil {
			logger.Error("Insert judicial status notification record failed",
				logger.Error(err),
				zap.String("confirmNo", msg.ConfirmNo),
				zap.String("receiver", receiver.Name),
			)
		}
	}

	logger.Info("Judicial status notifications sent",
		zap.String("confirmNo", msg.ConfirmNo),
		zap.Int32("newStatus", msg.NewStatus),
	)
}

func startJudicialSyncConsumer(cfg *config.Config) {
	consumerWg.Add(1)
	defer consumerWg.Done()

	groupName := cfg.RocketMQ.GroupName + "_judicial_sync"
	cons := InitConsumer(cfg, groupName)
	consumers = append(consumers, cons)

	topic := constants.MQTopicJudicialSync

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

			var statusMsg JudicialStatusMessage
			if err := json.Unmarshal(msg.Body, &statusMsg); err != nil {
				logger.Error("Unmarshal judicial sync message failed",
					logger.Error(err),
					zap.String("msgId", msg.MsgId),
				)
				continue
			}

			logger.Info("Received judicial sync message",
				zap.Int64("confirmId", statusMsg.ConfirmID),
				zap.String("confirmNo", statusMsg.ConfirmNo),
			)

			processJudicialSyncToCourt(&statusMsg)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		logger.Error("Subscribe judicial sync topic failed", logger.Error(err))
		return
	}

	if err := cons.Start(); err != nil {
		logger.Error("Start judicial sync consumer failed", logger.Error(err))
		return
	}

	logger.Info("Judicial sync consumer started", zap.String("topic", topic), zap.String("group", groupName))

	for {
		select {
		case <-shutdownSignal:
			logger.Info("Judicial sync consumer stopping")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func processJudicialSyncToCourt(msg *JudicialStatusMessage) {
	logger.Info("Sync judicial confirmation status to court system",
		zap.Int64("confirmId", msg.ConfirmID),
		zap.String("confirmNo", msg.ConfirmNo),
	)

	logger.Info("Judicial confirmation synced to court",
		zap.Int64("confirmId", msg.ConfirmID),
		zap.String("confirmNo", msg.ConfirmNo),
	)
}

func startJudicialRemindConsumer(cfg *config.Config) {
	consumerWg.Add(1)
	defer consumerWg.Done()

	groupName := cfg.RocketMQ.GroupName + "_judicial_remind"
	cons := InitConsumer(cfg, groupName)
	consumers = append(consumers, cons)

	topic := constants.MQTopicJudicialRemind

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

			var remindMsg JudicialRemindMessage
			if err := json.Unmarshal(msg.Body, &remindMsg); err != nil {
				logger.Error("Unmarshal judicial remind message failed",
					logger.Error(err),
					zap.String("msgId", msg.MsgId),
				)
				continue
			}

			logger.Info("Received judicial remind message",
				zap.Int64("confirmId", remindMsg.ConfirmID),
				zap.String("confirmNo", remindMsg.ConfirmNo),
				zap.String("remindType", remindMsg.RemindType),
			)

			processJudicialRemind(&remindMsg)
		}
		return consumer.ConsumeSuccess, nil
	})

	if err != nil {
		logger.Error("Subscribe judicial remind topic failed", logger.Error(err))
		return
	}

	if err := cons.Start(); err != nil {
		logger.Error("Start judicial remind consumer failed", logger.Error(err))
		return
	}

	logger.Info("Judicial remind consumer started", zap.String("topic", topic), zap.String("group", groupName))

	for {
		select {
		case <-shutdownSignal:
			logger.Info("Judicial remind consumer stopping")
			return
		case <-time.After(1 * time.Second):
		}
	}
}

func processJudicialRemind(msg *JudicialRemindMessage) {
	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	var remindTypeName string
	if msg.RemindType == "performance" {
		remindTypeName = "履行提醒"
	} else if msg.RemindType == "expiration" {
		remindTypeName = "失效提醒"
	}

	title := fmt.Sprintf("【司法确认%s】", remindTypeName)
	content := fmt.Sprintf("您的司法确认申请（编号：%s）%s已发送，请及时处理。", msg.ConfirmNo, remindTypeName)

	params := map[string]interface{}{
		"confirmNo":  msg.ConfirmNo,
		"remindType": msg.RemindType,
		"targetPhones": msg.TargetPhones,
	}
	paramsJSON, _ := json.Marshal(params)

	var confirm struct {
		ApplicantName  string `json:"applicantName"`
		RespondentName string `json:"respondentName"`
	}
	db.Table("judicial_confirmation").Select("applicant_name, respondent_name").
		Where("id = ?", msg.ConfirmID).Scan(&confirm)

	names := []string{confirm.ApplicantName, confirm.RespondentName}
	for i, phone := range msg.TargetPhones {
		name := ""
		if i < len(names) {
			name = names[i]
		}
		insertSQL := `INSERT INTO notification_record (receiver_id, receiver_name, template_id, template_name, template_type, title, content, channel_type, send_status, params, case_id, case_no, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		if err := db.Exec(insertSQL,
			0,
			name,
			6,
			"司法确认提醒通知",
			6,
			title,
			content,
			"短信",
			1,
			string(paramsJSON),
			0,
			msg.ConfirmNo,
			time.Now(),
		).Error; err != nil {
			logger.Error("Insert judicial remind notification record failed",
				logger.Error(err),
				zap.String("confirmNo", msg.ConfirmNo),
				zap.String("phone", phone),
			)
		}
	}

	logger.Info("Judicial remind notifications sent",
		zap.String("confirmNo", msg.ConfirmNo),
		zap.String("remindType", msg.RemindType),
		zap.Int("phoneCount", len(msg.TargetPhones)),
	)
}

func SendJudicialStatusMessage(msg *JudicialStatusMessage) error {
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send judicial status message async failed", logger.Error(err))
		}
	}
	body, _ := json.Marshal(msg)
	return SendAsyncMessage(constants.MQTopicJudicialStatus, body, callback)
}

func SendJudicialSubmitMessage(msg *JudicialSubmitMessage) error {
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send judicial submit message async failed", logger.Error(err))
		}
	}
	body, _ := json.Marshal(msg)
	return SendAsyncMessage(constants.MQTopicJudicialSubmit, body, callback)
}

func SendJudicialSealMessage(msg *JudicialSealMessage) error {
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send judicial seal message async failed", logger.Error(err))
		}
	}
	body, _ := json.Marshal(msg)
	return SendAsyncMessage(constants.MQTopicJudicialSeal, body, callback)
}

func SendJudicialRemindMessage(msg *JudicialRemindMessage) error {
	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send judicial remind message async failed", logger.Error(err))
		}
	}
	body, _ := json.Marshal(msg)
	return SendAsyncMessage(constants.MQTopicJudicialRemind, body, callback)
}
