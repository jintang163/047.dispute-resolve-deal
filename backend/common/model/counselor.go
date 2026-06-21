package model

import "time"

type Counselor struct {
	BaseModel
	CounselorNo        string    `gorm:"size:50;uniqueIndex;not null" json:"counselorNo"`
	UserID             int64     `gorm:"index" json:"userId"`
	RealName           string    `gorm:"size:50;not null" json:"realName"`
	Gender             int       `gorm:"default:0" json:"gender"`
	Phone              string    `gorm:"size:20" json:"phone"`
	Email              string    `gorm:"size:100" json:"email"`
	Avatar             string    `gorm:"size:255" json:"avatar"`
	Title              string    `gorm:"size:100" json:"title"`
	LicenseNo          string    `gorm:"size:100" json:"licenseNo"`
	Specialty          string    `gorm:"size:500" json:"specialty"`
	SpecialtyTags      string    `gorm:"size:500" json:"specialtyTags"`
	YearsOfExperience  int       `json:"yearsOfExperience"`
	Education          string    `gorm:"size:100" json:"education"`
	Introduction       string    `gorm:"type:text" json:"introduction"`
	ConsultationTypes  string    `gorm:"size:200" json:"consultationTypes"`
	WorkDays           string    `gorm:"size:100" json:"workDays"`
	WorkStartTime      string    `gorm:"type:time" json:"workStartTime"`
	WorkEndTime        string    `gorm:"type:time" json:"workEndTime"`
	SessionDuration    int       `gorm:"default:50" json:"sessionDuration"`
	Price              float64   `gorm:"type:decimal(10,2);default:0" json:"price"`
	OrganizationID     int64     `gorm:"column:org_id;index" json:"organizationId"`
	OrganizationName   string    `gorm:"column:org_name;size:100" json:"organizationName"`
	RatingAvg          float64   `gorm:"type:decimal(3,2);default:0" json:"ratingAvg"`
	RatingCount        int       `gorm:"default:0" json:"ratingCount"`
	AppointmentCount   int       `gorm:"default:0" json:"appointmentCount"`
	CompletedCount     int       `gorm:"default:0" json:"completedCount"`
	IsEmergencyAvailable int     `gorm:"default:0;index" json:"isEmergencyAvailable"`
	Status             int32     `gorm:"default:1;index" json:"status"`
	SortOrder          int       `gorm:"default:0" json:"sortOrder"`
	CreatedBy          int64     `json:"createdBy"`

	OrgID   int64  `gorm:"-" json:"orgId"`
	OrgName string `gorm:"-" json:"orgName"`
}

func (c *Counselor) AfterFind(tx interface{}) error {
	c.OrgID = c.OrganizationID
	c.OrgName = c.OrganizationName
	return nil
}

func (Counselor) TableName() string {
	return "counselor"
}

type CounselorAppointment struct {
	BaseModel
	AppointmentNo           string     `gorm:"size:50;uniqueIndex;not null" json:"appointmentNo"`
	CounselorID             int64      `gorm:"not null;index" json:"counselorId"`
	CounselorName           string     `gorm:"size:50" json:"counselorName"`
	CaseID                  int64      `gorm:"index" json:"caseId"`
	PartyID                 int64      `gorm:"index" json:"partyId"`
	PartyName               string     `gorm:"size:50" json:"partyName"`
	PartyPhone              string     `gorm:"size:20" json:"partyPhone"`
	PartyIDCard             string     `gorm:"size:18" json:"partyIdCard"`
	IsAnonymous             int32      `gorm:"default:0;index" json:"isAnonymous"`
	AnonymousCode           string     `gorm:"size:50" json:"anonymousCode"`
	AppointmentDate         string     `gorm:"type:date;not null;index" json:"appointmentDate"`
	StartTime               string     `gorm:"type:time;not null" json:"startTime"`
	EndTime                 string     `gorm:"type:time;not null" json:"endTime"`
	ConsultationType        int        `gorm:"default:1" json:"consultationType"`
	AppointmentSource       int        `gorm:"default:1" json:"appointmentSource"`
	IsEmergency             int32      `gorm:"default:0;index" json:"isEmergency"`
	EmergencyTriggerWords   string     `gorm:"size:500" json:"emergencyTriggerWords"`
	EmergencyLevel          int        `gorm:"default:0" json:"emergencyLevel"`
	ConcernType             string     `gorm:"size:100" json:"concernType"`
	ConcernDescription      string     `gorm:"type:text" json:"concernDescription"`
	Status                  int32      `gorm:"default:10;index" json:"status"`
	CancelReason            string     `gorm:"size:500" json:"cancelReason"`
	CancelledBy             int64      `json:"cancelledBy"`
	CancelledAt             *time.Time `json:"cancelledAt"`
	ConfirmedBy             int64      `json:"confirmedBy"`
	ConfirmedAt             *time.Time `json:"confirmedAt"`
	StartedAt               *time.Time `json:"startedAt"`
	CompletedAt             *time.Time `json:"completedAt"`
	ConsultationSummary     string     `gorm:"type:text" json:"consultationSummary"`
	FollowUpSuggestion      string     `gorm:"type:text" json:"followUpSuggestion"`
	NextAppointmentSuggestion string   `gorm:"size:200" json:"nextAppointmentSuggestion"`
	RoomID                  string     `gorm:"size:100" json:"roomId"`
	RoomURL                 string     `gorm:"size:255" json:"roomUrl"`
	Location                string     `gorm:"size:255" json:"location"`
	ReminderSent            int32      `gorm:"default:0" json:"reminderSent"`
	RatingSubmitted         int32      `gorm:"default:0" json:"ratingSubmitted"`
	CreatedBy               int64      `json:"createdBy"`
	CreatedByName           string     `gorm:"size:50" json:"createdByName"`
}

func (CounselorAppointment) TableName() string {
	return "counselor_appointment"
}

type CounselorRating struct {
	BaseModel
	AppointmentID     int64      `gorm:"not null;index" json:"appointmentId"`
	CounselorID       int64      `gorm:"not null;index" json:"counselorId"`
	RaterID           int64      `gorm:"index" json:"raterId"`
	RaterName         string     `gorm:"size:50" json:"raterName"`
	IsAnonymousRating int32      `gorm:"default:0" json:"isAnonymousRating"`
	OverallScore      int        `gorm:"not null" json:"overallScore"`
	ProfessionalScore int        `gorm:"default:0" json:"professionalScore"`
	AttitudeScore     int        `gorm:"default:0" json:"attitudeScore"`
	EmpathyScore      int        `gorm:"default:0" json:"empathyScore"`
	HelpfulScore      int        `gorm:"default:0" json:"helpfulScore"`
	Content           string     `gorm:"type:text" json:"content"`
	Tags              string     `gorm:"size:500" json:"tags"`
	CounselorReply    string     `gorm:"type:text" json:"counselorReply"`
	CounselorReplyAt  *time.Time `json:"counselorReplyAt"`
	IsHelpful         int        `gorm:"default:0" json:"isHelpful"`
	Status            int32      `gorm:"default:1" json:"status"`
}

func (CounselorRating) TableName() string {
	return "counselor_rating"
}

type CounselorSchedule struct {
	BaseModel
	CounselorID   int64  `gorm:"not null;index" json:"counselorId"`
	ScheduleDate  string `gorm:"type:date;not null" json:"scheduleDate"`
	StartTime     string `gorm:"type:time;not null" json:"startTime"`
	EndTime       string `gorm:"type:time;not null" json:"endTime"`
	ScheduleType  int    `gorm:"default:1" json:"scheduleType"`
	Title         string `gorm:"size:200" json:"title"`
	Remark        string `gorm:"size:500" json:"remark"`
	AppointmentID int64  `gorm:"index" json:"appointmentId"`
	CreatedBy     int64  `json:"createdBy"`
}

func (CounselorSchedule) TableName() string {
	return "counselor_schedule"
}
