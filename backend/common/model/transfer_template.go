package model

import "time"

type TransferTemplate struct {
	BaseModel
	TemplateName    string `gorm:"size:128;not null" json:"templateName"`
	DeptCode        string `gorm:"size:64;uniqueIndex;not null" json:"deptCode"`
	DeptName        string `gorm:"size:128;not null" json:"deptName"`
	DeptType        string `gorm:"size:32;not null" json:"deptType"`
	ContactPerson   string `gorm:"size:64" json:"contactPerson"`
	ContactPhone    string `gorm:"size:20" json:"contactPhone"`
	ContactEmail    string `gorm:"size:128" json:"contactEmail"`
	Description     string `gorm:"size:512" json:"description"`
	ApplicableTypes string `gorm:"size:512" json:"applicableTypes"`
	SortOrder       int    `gorm:"default:0" json:"sortOrder"`
	Status          int32  `gorm:"default:1;index" json:"status"`
}

func (TransferTemplate) TableName() string {
	return "dispute_transfer_template"
}

type DisputeTransfer struct {
	BaseModel
	TransferNo      string     `gorm:"size:32;uniqueIndex;not null" json:"transferNo"`
	CaseID          int64      `gorm:"index;not null" json:"caseId"`
	CaseNo          string     `gorm:"size:32;index" json:"caseNo"`
	CaseTitle       string     `gorm:"size:256" json:"caseTitle"`
	TemplateID      int64      `gorm:"default:0" json:"templateId"`
	FromDeptID      int64      `gorm:"index;not null" json:"fromDeptId"`
	FromDeptName    string     `gorm:"size:128;not null" json:"fromDeptName"`
	FromUserID      int64      `gorm:"not null" json:"fromUserId"`
	FromUserName    string     `gorm:"size:64;not null" json:"fromUserName"`
	ToDeptCode      string     `gorm:"size:64;index;not null" json:"toDeptCode"`
	ToDeptName      string     `gorm:"size:128;not null" json:"toDeptName"`
	ToDeptType      string     `gorm:"size:32;not null" json:"toDeptType"`
	ToContactPerson string     `gorm:"size:64" json:"toContactPerson"`
	ToContactPhone  string     `gorm:"size:20" json:"toContactPhone"`
	TransferReason  string     `gorm:"size:1000;not null" json:"transferReason"`
	TransferRemark  string     `gorm:"size:500" json:"transferRemark"`
	AttachIDs       string     `gorm:"size:512" json:"attachIds"`
	Status          int32      `gorm:"default:10;index" json:"status"`
	ReceiveTime     *time.Time `json:"receiveTime"`
	ReceiveUserID   int64      `gorm:"default:0" json:"receiveUserId"`
	ReceiveUserName string     `gorm:"size:64" json:"receiveUserName"`
	ReceiveRemark   string     `gorm:"size:500" json:"receiveRemark"`
	RejectReason    string     `gorm:"size:500" json:"rejectReason"`
	RejectTime      *time.Time `json:"rejectTime"`
	ProcessStartTime *time.Time `json:"processStartTime"`
	ProcessEndTime  *time.Time `json:"processEndTime"`
	ProcessResult   string     `gorm:"size:1000" json:"processResult"`
	ProcessDuration int        `gorm:"default:0" json:"processDuration"`
	UrgeCount       int        `gorm:"default:0" json:"urgeCount"`
	LastUrgeTime    *time.Time `json:"lastUrgeTime"`
	FirstUrgeTime   *time.Time `json:"firstUrgeTime"`
	TimeoutHours    int        `gorm:"default:72" json:"timeoutHours"`
	IsTimeout       int32      `gorm:"default:0;index" json:"isTimeout"`
	ClosedAt        *time.Time `json:"closedAt"`
}

func (DisputeTransfer) TableName() string {
	return "dispute_transfer"
}

type DisputeTransferUrge struct {
	BaseModel
	TransferID    int64  `gorm:"index;not null" json:"transferId"`
	TransferNo    string `gorm:"size:32;not null" json:"transferNo"`
	UrgeType      int    `gorm:"index;not null" json:"urgeType"`
	UrgeSource    int    `gorm:"default:2" json:"urgeSource"`
	OperatorID    int64  `gorm:"default:0" json:"operatorId"`
	OperatorName  string `gorm:"size:64" json:"operatorName"`
	UrgencyLevel  int    `gorm:"default:2" json:"urgencyLevel"`
	UrgeContent   string `gorm:"size:500;not null" json:"urgeContent"`
	NotifyType    string `gorm:"size:32;default:'app,sms'" json:"notifyType"`
}

func (DisputeTransferUrge) TableName() string {
	return "dispute_transfer_urge"
}
