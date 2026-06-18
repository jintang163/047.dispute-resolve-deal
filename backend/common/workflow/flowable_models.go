package workflow

import "time"

type FlowableTask struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Assignee          string    `json:"assignee"`
	ProcessInstanceID string    `json:"processInstanceId"`
	ProcessDefinitionID string  `json:"processDefinitionId"`
	CreateTime        time.Time `json:"createTime"`
	DueDate           time.Time `json:"dueDate"`
	Priority          int       `json:"priority"`
	Owner             string    `json:"owner"`
	ParentTaskID      string    `json:"parentTaskId"`
	TaskDefinitionKey string    `json:"taskDefinitionKey"`
	Description       string    `json:"description"`
	Category          string    `json:"category"`
}

type FlowableProcessInstance struct {
	ID                  string    `json:"id"`
	ProcessDefinitionID string    `json:"processDefinitionId"`
	BusinessKey         string    `json:"businessKey"`
	StartTime           time.Time `json:"startTime"`
	EndTime             time.Time `json:"endTime"`
	StartUserID         string    `json:"startUserId"`
	State               string    `json:"state"`
	Name                string    `json:"name"`
	StartedActivityID   string    `json:"startedActivityId"`
}

type ProcessStatus struct {
	Status         string              `json:"status"`
	CurrentTasks   []*FlowableTask     `json:"currentTasks"`
	CompletedTasks []*FlowableTask     `json:"completedTasks"`
	History        []*HistoryActivity  `json:"history"`
}

type FlowableVariable struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
	Type  string      `json:"type"`
}

type HistoryActivity struct {
	ID                   string    `json:"id"`
	ActivityID           string    `json:"activityId"`
	ActivityName         string    `json:"activityName"`
	ActivityType         string    `json:"activityType"`
	ProcessInstanceID    string    `json:"processInstanceId"`
	StartTime            time.Time `json:"startTime"`
	EndTime              time.Time `json:"endTime"`
	DurationInMillis     int64     `json:"durationInMillis"`
	Assignee             string    `json:"assignee"`
	TaskID               string    `json:"taskId"`
	DeleteReason         string    `json:"deleteReason"`
}

type ApprovalNode struct {
	NodeName   string    `json:"nodeName"`
	ApproverID int64     `json:"approverId"`
	Approver   string    `json:"approver"`
	Status     int32     `json:"status"`
	StatusText string    `json:"statusText"`
	Action     int       `json:"action"`
	ActionText string    `json:"actionText"`
	Remark     string    `json:"remark"`
	CreateTime time.Time `json:"createTime"`
	HandleTime time.Time `json:"handleTime"`
	SortOrder  int       `json:"sortOrder"`
}

type FlowableTaskRequest struct {
	Action   string                 `json:"action"`
	Assignee string                 `json:"assignee,omitempty"`
	Variables []FlowableVariable    `json:"variables,omitempty"`
	TransientVariables []FlowableVariable `json:"transientVariables,omitempty"`
}

type FlowableCommentRequest struct {
	Message string `json:"message"`
	SaveProcessInstanceId bool `json:"saveProcessInstanceId"`
}

type FlowableIdentityLink struct {
	UserID string `json:"userId"`
	GroupID string `json:"groupId"`
	Type    string `json:"type"`
}
