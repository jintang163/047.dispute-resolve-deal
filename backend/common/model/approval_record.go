package model

import "time"

type ApprovalRecord struct {
	BaseModel
	CaseID           int64      `gorm:"index;not null" json:"caseId"`
	WorkflowID       int64      `gorm:"index" json:"workflowId"`
	WorkflowName     string     `gorm:"size:100" json:"workflowName"`
	NodeType         int        `gorm:"index" json:"nodeType"`
	NodeName         string     `gorm:"size:100" json:"nodeName"`
	ApproverID       int64      `gorm:"index" json:"approverId"`
	ApproverName     string     `gorm:"size:50" json:"approverName"`
	Status           int32      `gorm:"index" json:"status"`
	Remark           string     `gorm:"size:500" json:"remark"`
	ApproveAction    int        `json:"approveAction"`
	ActionName       string     `gorm:"size:50" json:"actionName"`
	ApprovedAt       *time.Time `json:"approvedAt"`
	SortOrder        int        `gorm:"default:0" json:"sortOrder"`
	SignUserID       int64      `json:"signUserId"`
	SignUserName     string     `gorm:"size:50" json:"signUserName"`
	TransferUserID   int64      `json:"transferUserId"`
	TransferUserName string     `gorm:"size:50" json:"transferUserName"`
	Level            int        `json:"level"`
	Deadline         *time.Time `json:"deadline"`
	TimeoutLevel     int        `json:"timeoutLevel"`
}

func (ApprovalRecord) TableName() string {
	return "approval_record"
}
