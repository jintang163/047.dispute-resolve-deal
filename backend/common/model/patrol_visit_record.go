package model

import "time"

type PatrolVisitRecord struct {
	BaseModel
	RecordNo        string     `gorm:"size:32;uniqueIndex" json:"recordNo"`
	TaskID          int64      `gorm:"index" json:"taskId"`
	PointID         int64      `gorm:"index" json:"pointId"`
	CheckinID       int64      `json:"checkinId"`
	MemberID        int64      `gorm:"index;not null" json:"memberId"`
	MemberName      string     `gorm:"size:64" json:"memberName"`
	OrganizationID  int64      `gorm:"column:organization_id;index" json:"organizationId"`
	VisitType       int        `gorm:"default:1;index" json:"visitType"`
	VisitObject     string     `gorm:"size:128" json:"visitObject"`
	VisitObjectPhone string    `gorm:"size:20" json:"visitObjectPhone"`
	VisitAddress    string     `gorm:"size:256" json:"visitAddress"`
	Longitude       float64    `json:"longitude"`
	Latitude        float64    `json:"latitude"`
	VisitTime       *time.Time `gorm:"default:CURRENT_TIMESTAMP;index" json:"visitTime"`
	VisitDuration   int        `json:"visitDuration"`
	Content         string     `gorm:"type:text" json:"content"`
	Situation       string     `gorm:"type:text" json:"situation"`
	ProblemDesc     string     `gorm:"type:text" json:"problemDesc"`
	HandleSituation string     `gorm:"type:text" json:"handleSituation"`
	NextPlan        string     `gorm:"type:text" json:"nextPlan"`
	PhotoURLs       string     `gorm:"type:text" json:"photoUrls"`
	VideoURL        string     `gorm:"size:512" json:"videoUrl"`
	AudioURL        string     `gorm:"size:512" json:"audioUrl"`
	HasDanger       int        `gorm:"default:0;index" json:"hasDanger"`
	DangerID        int64      `json:"dangerId"`
	HasDispute      int        `gorm:"default:0;index" json:"hasDispute"`
	DisputeCaseID   int64      `json:"disputeCaseId"`
	Status          int32      `gorm:"default:1;index" json:"status"`
	AuditorID       int64      `json:"auditorId"`
	AuditTime       *time.Time `json:"auditTime"`
	AuditRemark     string     `gorm:"size:512" json:"auditRemark"`
	PointsReward    int        `json:"pointsReward"`
}

func (PatrolVisitRecord) TableName() string {
	return "patrol_visit_record"
}

type PatrolVisitRecordQuery struct {
	BaseQuery
	MemberID       int64  `form:"memberId" json:"memberId"`
	TaskID         int64  `form:"taskId" json:"taskId"`
	VisitType      int    `form:"visitType" json:"visitType"`
	Status         int    `form:"status" json:"status"`
	OrgID          int64  `form:"orgId" json:"orgId"`
	HasDanger      int    `form:"hasDanger" json:"hasDanger"`
	HasDispute     int    `form:"hasDispute" json:"hasDispute"`
	RecordNo       string `form:"recordNo" json:"recordNo"`
	VisitObject    string `form:"visitObject" json:"visitObject"`
	DateRangeQuery
}

type CreateVisitRecordRequest struct {
	TaskID           int64    `json:"taskId"`
	PointID          int64    `json:"pointId"`
	CheckinID        int64    `json:"checkinId"`
	VisitType        int      `json:"visitType" binding:"required"`
	VisitObject      string   `json:"visitObject"`
	VisitObjectPhone string   `json:"visitObjectPhone"`
	VisitAddress     string   `json:"visitAddress"`
	Longitude        float64  `json:"longitude"`
	Latitude         float64  `json:"latitude"`
	VisitTime        string   `json:"visitTime"`
	VisitDuration    int      `json:"visitDuration"`
	Content          string   `json:"content" binding:"required"`
	Situation        string   `json:"situation"`
	ProblemDesc      string   `json:"problemDesc"`
	HandleSituation  string   `json:"handleSituation"`
	NextPlan         string   `json:"nextPlan"`
	PhotoURLs        []string `json:"photoUrls"`
	VideoURL         string   `json:"videoUrl"`
	AudioURL         string   `json:"audioUrl"`
	HasDanger        int      `json:"hasDanger"`
	DangerInfo       *DangerInfo `json:"dangerInfo"`
	HasDispute       int      `json:"hasDispute"`
	DisputeInfo      *DisputeInfo `json:"disputeInfo"`
	Status           int32    `json:"status"`
}

type DangerInfo struct {
	DangerType      int      `json:"dangerType" binding:"required"`
	DangerLevel     int      `json:"dangerLevel"`
	Title           string   `json:"title" binding:"required"`
	Description     string   `json:"description"`
	InvolvedPerson  string   `json:"involvedPerson"`
	InvolvedPhone   string   `json:"involvedPhone"`
	PhotoURLs       []string `json:"photoUrls"`
	VideoURL        string   `json:"videoUrl"`
}

type DisputeInfo struct {
	Title          string   `json:"title" binding:"required"`
	TypeID         int64    `json:"typeId" binding:"required"`
	TypeName       string   `json:"typeName"`
	Description    string   `json:"description"`
	CaseLevel      int      `json:"caseLevel"`
	RespondentName string   `json:"respondentName"`
	RespondentPhone string  `json:"respondentPhone"`
	Expectation    string   `json:"expectation"`
	PhotoURLs      []string `json:"photoUrls"`
}

type AuditVisitRecordRequest struct {
	ID          int64  `json:"id" binding:"required"`
	Status      int32  `json:"status" binding:"required"`
	AuditRemark string `json:"auditRemark"`
}

type VisitRecordStatisticsResponse struct {
	TodayVisitCount    int `json:"todayVisitCount"`
	WeekVisitCount     int `json:"weekVisitCount"`
	MonthVisitCount    int `json:"monthVisitCount"`
	TotalVisitCount    int `json:"totalVisitCount"`
	TodayDangerCount   int `json:"todayDangerCount"`
	TotalDangerCount   int `json:"totalDangerCount"`
	TodayDisputeCount  int `json:"todayDisputeCount"`
	TotalDisputeCount  int `json:"totalDisputeCount"`
}
