package model

type AIAssistRecord struct {
	BaseModel
	CaseID     int64  `gorm:"index" json:"caseId"`
	AssistType int    `gorm:"index" json:"assistType"`
	Input      string `gorm:"type:text" json:"input"`
	Output     string `gorm:"type:text" json:"output"`
	TokenUsage int64  `json:"tokenUsage"`
	CreatedBy  int64  `json:"createdBy"`
}

func (AIAssistRecord) TableName() string {
	return "ai_assist_record"
}
