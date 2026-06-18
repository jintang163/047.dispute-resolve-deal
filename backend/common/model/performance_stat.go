package model

import "time"

type PerformanceStat struct {
	BaseModel
	UserID            int64     `gorm:"index;not null" json:"userId"`
	UserName          string    `gorm:"size:50;not null" json:"userName"`
	OrgID             int64     `gorm:"index" json:"orgId"`
	OrgName           string    `gorm:"size:100" json:"orgName"`
	PeriodType        int       `gorm:"index;not null" json:"periodType"`
	PeriodDate        string    `gorm:"size:20;index" json:"periodDate"`
	CaseCount         int       `gorm:"default:0" json:"caseCount"`
	CloseCount        int       `gorm:"default:0" json:"closeCount"`
	CloseRate         float64   `gorm:"default:0" json:"closeRate"`
	SuccessCount      int       `gorm:"default:0" json:"successCount"`
	SuccessRate       float64   `gorm:"default:0" json:"successRate"`
	AvgDays           float64   `gorm:"default:0" json:"avgDays"`
	TotalSatisfaction int       `gorm:"default:0" json:"totalSatisfaction"`
	AvgSatisfaction   float64   `gorm:"default:0" json:"avgSatisfaction"`
	Score             float64   `gorm:"default:0" json:"score"`
	Grade             string    `gorm:"size:10" json:"grade"`
	CreatedAt         time.Time `gorm:"autoCreateTime;index" json:"createdAt"`
}

func (PerformanceStat) TableName() string {
	return "performance_stat"
}
