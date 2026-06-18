package model

import "time"

const (
	ActionTypeSubmit       = 10
	ActionTypeCourtAccept  = 20
	ActionTypePass         = 30
	ActionTypeReject       = 40
	ActionTypeSeal         = 50
	ActionTypeDeliver      = 60
	ActionTypePerformRemind = 70
	ActionTypeExpireRemind = 80
	ActionTypePerformed    = 90
	ActionTypeExpired      = 99

	OperatorTypeSystem   = 1
	OperatorTypeAdmin    = 2
	OperatorTypeCourt    = 3
	OperatorTypeParty    = 4
)

type JudicialConfirmLog struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	ConfirmID       int64      `gorm:"index;not null" json:"confirmId"`
	ConfirmNo       string     `gorm:"size:32;not null" json:"confirmNo"`
	ActionType      int        `gorm:"index;not null" json:"actionType"`
	ActionName      string     `gorm:"size:64;not null" json:"actionName"`
	ActionDetail    string     `gorm:"type:text" json:"actionDetail"`
	OperatorID      int64      `json:"operatorId"`
	OperatorName    string     `gorm:"size:64" json:"operatorName"`
	OperatorType    int32      `gorm:"default:1" json:"operatorType"`

	CourtRemark     string     `gorm:"size:512" json:"courtRemark"`
	CourtOperator   string     `gorm:"size:64" json:"courtOperator"`
	CourtHandleTime *time.Time `json:"courtHandleTime"`

	CreatedAt       *time.Time `gorm:"index" json:"createdAt"`
}

func (JudicialConfirmLog) TableName() string {
	return "judicial_confirm_log"
}
