package workflow

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"
	"go.uber.org/zap"
)

type FlowableClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

var flowableClient *FlowableClient

func NewFlowableClient() *FlowableClient {
	cfg := config.GetConfig()
	return &FlowableClient{
		baseURL:  cfg.Flowable.BaseURL,
		username: cfg.Flowable.Username,
		password: cfg.Flowable.Password,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func InitFlowable() {
	flowableClient = NewFlowableClient()
	logger.Info("Flowable client initialized", zap.String("baseURL", flowableClient.baseURL))
}

func GetFlowableClient() *FlowableClient {
	if flowableClient == nil {
		InitFlowable()
	}
	return flowableClient
}

func (c *FlowableClient) basicAuth() string {
	auth := c.username + ":" + c.password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func (c *FlowableClient) doRequest(method, path string, body interface{}, result interface{}) error {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body failed: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}

	req.Header.Set("Authorization", c.basicAuth())
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	logger.Debug("Flowable request",
		zap.String("method", method),
		zap.String("url", url),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("flowable request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %w", err)
	}

	logger.Debug("Flowable response",
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(respBody)),
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("flowable api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("unmarshal response failed: %w, body=%s", err, string(respBody))
		}
	}

	return nil
}

type StartProcessRequest struct {
	ProcessDefinitionKey string                 `json:"processDefinitionKey"`
	BusinessKey          string                 `json:"businessKey"`
	Variables            []FlowableVariable     `json:"variables"`
}

type StartProcessResponse struct {
	ID                  string `json:"id"`
	URL                 string `json:"url"`
	BusinessKey         string `json:"businessKey"`
	Suspended           bool   `json:"suspended"`
	ProcessDefinitionID string `json:"processDefinitionId"`
	ProcessDefinitionURL string `json:"processDefinitionUrl"`
	ActivityID          string `json:"activityId"`
	Variables           []FlowableVariable `json:"variables"`
}

func (c *FlowableClient) StartProcess(processKey string, businessKey string, variables map[string]interface{}) (string, error) {
	path := "/runtime/process-instances"

	flowVars := make([]FlowableVariable, 0, len(variables))
	for k, v := range variables {
		flowVars = append(flowVars, FlowableVariable{
			Name:  k,
			Value: v,
			Type:  detectVariableType(v),
		})
	}

	req := StartProcessRequest{
		ProcessDefinitionKey: processKey,
		BusinessKey:          businessKey,
		Variables:            flowVars,
	}

	var resp StartProcessResponse
	if err := c.doRequest(http.MethodPost, path, req, &resp); err != nil {
		logger.Error("Start process failed",
			zap.String("processKey", processKey),
			zap.String("businessKey", businessKey),
			logger.Error(err),
		)
		return "", err
	}

	logger.Info("Process started",
		zap.String("processInstanceId", resp.ID),
		zap.String("processKey", processKey),
		zap.String("businessKey", businessKey),
	)

	return resp.ID, nil
}

func (c *FlowableClient) GetTask(taskId string) (*FlowableTask, error) {
	path := "/runtime/tasks/" + taskId

	var task FlowableTask
	if err := c.doRequest(http.MethodGet, path, nil, &task); err != nil {
		logger.Error("Get task failed",
			zap.String("taskId", taskId),
			logger.Error(err),
		)
		return nil, err
	}

	return &task, nil
}

func (c *FlowableClient) CompleteTask(taskId string, variables map[string]interface{}) error {
	path := "/runtime/tasks/" + taskId

	flowVars := make([]FlowableVariable, 0, len(variables))
	for k, v := range variables {
		flowVars = append(flowVars, FlowableVariable{
			Name:  k,
			Value: v,
			Type:  detectVariableType(v),
		})
	}

	body := map[string]interface{}{
		"action":    "complete",
		"variables": flowVars,
	}

	if err := c.doRequest(http.MethodPost, path, body, nil); err != nil {
		logger.Error("Complete task failed",
			zap.String("taskId", taskId),
			logger.Error(err),
		)
		return err
	}

	logger.Info("Task completed", zap.String("taskId", taskId))
	return nil
}

func (c *FlowableClient) RejectTask(taskId string, reason string) error {
	path := "/runtime/tasks/" + taskId

	body := map[string]interface{}{
		"action": "complete",
		"variables": []FlowableVariable{
			{Name: "approved", Value: false, Type: "boolean"},
			{Name: "rejectReason", Value: reason, Type: "string"},
			{Name: "action", Value: "reject", Type: "string"},
		},
	}

	if err := c.doRequest(http.MethodPost, path, body, nil); err != nil {
		logger.Error("Reject task failed",
			zap.String("taskId", taskId),
			zap.String("reason", reason),
			logger.Error(err),
		)
		return err
	}

	if err := c.AddComment(taskId, reason); err != nil {
		logger.Warn("Add reject comment failed", zap.String("taskId", taskId), logger.Error(err))
	}

	logger.Info("Task rejected", zap.String("taskId", taskId), zap.String("reason", reason))
	return nil
}

func (c *FlowableClient) DelegateTask(taskId string, userId string) error {
	path := "/runtime/tasks/" + taskId

	body := map[string]interface{}{
		"action":   "delegate",
		"assignee": userId,
	}

	if err := c.doRequest(http.MethodPost, path, body, nil); err != nil {
		logger.Error("Delegate task failed",
			zap.String("taskId", taskId),
			zap.String("userId", userId),
			logger.Error(err),
		)
		return err
	}

	logger.Info("Task delegated", zap.String("taskId", taskId), zap.String("userId", userId))
	return nil
}

func (c *FlowableClient) AddMultiInstance(taskId string, assignee string) error {
	task, err := c.GetTask(taskId)
	if err != nil {
		return err
	}

	path := "/runtime/tasks/" + taskId + "/identitylinks"

	body := FlowableIdentityLink{
		UserID: assignee,
		Type:   "candidate",
	}

	if err := c.doRequest(http.MethodPost, path, body, nil); err != nil {
		logger.Error("Add multi instance failed",
			zap.String("taskId", taskId),
			zap.String("assignee", assignee),
			logger.Error(err),
		)
		return err
	}

	logger.Info("Multi instance added",
		zap.String("taskId", taskId),
		zap.String("assignee", assignee),
		zap.String("processInstanceId", task.ProcessInstanceID),
	)
	return nil
}

type TaskListResponse struct {
	Data  []*FlowableTask `json:"data"`
	Total int             `json:"total"`
}

func (c *FlowableClient) GetProcessInstanceTasks(processInstanceId string) ([]*FlowableTask, error) {
	path := "/runtime/tasks?processInstanceId=" + processInstanceId

	var resp TaskListResponse
	if err := c.doRequest(http.MethodGet, path, nil, &resp); err != nil {
		logger.Error("Get process tasks failed",
			zap.String("processInstanceId", processInstanceId),
			logger.Error(err),
		)
		return nil, err
	}

	return resp.Data, nil
}

type HistoryTaskResponse struct {
	Data  []*HistoryTask `json:"data"`
	Total int            `json:"total"`
}

type HistoryTask struct {
	ID                     string    `json:"id"`
	ProcessInstanceID      string    `json:"processInstanceId"`
	TaskDefinitionKey      string    `json:"taskDefinitionKey"`
	Name                   string    `json:"name"`
	Description            string    `json:"description"`
	StartTime              time.Time `json:"startTime"`
	EndTime                time.Time `json:"endTime"`
	DurationInMillis       int64     `json:"durationInMillis"`
	Assignee               string    `json:"assignee"`
	Owner                  string    `json:"owner"`
	EndTimeDefined         string    `json:"endTimeDefined"`
	DueDate                time.Time `json:"dueDate"`
	DeleteReason           string    `json:"deleteReason"`
	Priority               int       `json:"priority"`
	Category               string    `json:"category"`
	ProcessDefinitionID    string    `json:"processDefinitionId"`
	ProcessDefinitionKey   string    `json:"processDefinitionKey"`
	ParentTaskID           string    `json:"parentTaskId"`
	FormKey                string    `json:"formKey"`
}

type HistoryActivityResponse struct {
	Data  []*HistoryActivity `json:"data"`
	Total int                `json:"total"`
}

func (c *FlowableClient) GetProcessStatus(processInstanceId string) (*ProcessStatus, error) {
	status := &ProcessStatus{}

	tasks, err := c.GetProcessInstanceTasks(processInstanceId)
	if err != nil {
		return nil, err
	}
	status.CurrentTasks = tasks

	historyPath := "/history/historic-task-instances?processInstanceId=" + processInstanceId + "&finished=true"
	var historyTaskResp HistoryTaskResponse
	if err := c.doRequest(http.MethodGet, historyPath, nil, &historyTaskResp); err != nil {
		logger.Warn("Get history tasks failed",
			zap.String("processInstanceId", processInstanceId),
			logger.Error(err),
		)
	}

	status.CompletedTasks = make([]*FlowableTask, 0, len(historyTaskResp.Data))
	for _, ht := range historyTaskResp.Data {
		status.CompletedTasks = append(status.CompletedTasks, &FlowableTask{
			ID:                ht.ID,
			Name:              ht.Name,
			Assignee:          ht.Assignee,
			ProcessInstanceID: ht.ProcessInstanceID,
			ProcessDefinitionID: ht.ProcessDefinitionID,
			CreateTime:        ht.StartTime,
			DueDate:           ht.DueDate,
			Priority:          ht.Priority,
			Owner:             ht.Owner,
			ParentTaskID:      ht.ParentTaskID,
			TaskDefinitionKey: ht.TaskDefinitionKey,
			Description:       ht.Description,
			Category:          ht.Category,
		})
	}

	activityPath := "/history/historic-activity-instances?processInstanceId=" + processInstanceId
	var activityResp HistoryActivityResponse
	if err := c.doRequest(http.MethodGet, activityPath, nil, &activityResp); err != nil {
		logger.Warn("Get history activities failed",
			zap.String("processInstanceId", processInstanceId),
			logger.Error(err),
		)
	}
	status.History = activityResp.Data

	if len(status.CurrentTasks) == 0 && len(status.CompletedTasks) > 0 {
		status.Status = "completed"
	} else if len(status.CurrentTasks) > 0 {
		status.Status = "active"
	} else {
		status.Status = "unknown"
	}

	return status, nil
}

func (c *FlowableClient) AddComment(taskId string, message string) error {
	path := "/runtime/tasks/" + taskId + "/comments"

	body := FlowableCommentRequest{
		Message:              message,
		SaveProcessInstanceId: true,
	}

	if err := c.doRequest(http.MethodPost, path, body, nil); err != nil {
		logger.Error("Add comment failed",
			zap.String("taskId", taskId),
			logger.Error(err),
		)
		return err
	}

	return nil
}

func (c *FlowableClient) ClaimTask(taskId string, userId string) error {
	path := "/runtime/tasks/" + taskId

	body := map[string]interface{}{
		"action":   "claim",
		"assignee": userId,
	}

	if err := c.doRequest(http.MethodPost, path, body, nil); err != nil {
		logger.Error("Claim task failed",
			zap.String("taskId", taskId),
			zap.String("userId", userId),
			logger.Error(err),
		)
		return err
	}

	logger.Info("Task claimed", zap.String("taskId", taskId), zap.String("userId", userId))
	return nil
}

func detectVariableType(v interface{}) string {
	switch v.(type) {
	case string:
		return "string"
	case int, int32, int64:
		return "integer"
	case float32, float64:
		return "double"
	case bool:
		return "boolean"
	case time.Time:
		return "date"
	default:
		return "json"
	}
}
