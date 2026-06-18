package model

import "time"

type DisputeHistory struct {
	BaseModel
	CaseID        int64     `gorm:"index;not null" json:"caseId"`
	ActionType    int       `gorm:"index;not null" json:"actionType"`
	ActionName    string    `gorm:"size:100;not null" json:"actionName"`
	Remark        string    `gorm:"size:500" json:"remark"`
	OperatorID    int64     `json:"operatorId"`
	OperatorName  string    `gorm:"size:50" json:"operatorName"`
	OperatorRole  string    `gorm:"size:50" json:"operatorRole"`
	CreatedAt     time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (DisputeHistory) TableName() string {
	return "dispute_history"
}
