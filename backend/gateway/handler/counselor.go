package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

type CounselorListRequest struct {
	model.BaseQuery
	Status              int32  `form:"status"`
	Specialty           string `form:"specialty"`
	IsEmergencyAvailable int32  `form:"isEmergencyAvailable"`
	ConsultationType    int    `form:"consultationType"`
	OrganizationID      int64  `form:"organizationId"`
}

type CounselorCreateRequest struct {
	RealName             string  `json:"realName" binding:"required"`
	Gender               int     `json:"gender"`
	Phone                string  `json:"phone"`
	Email                string  `json:"email"`
	Avatar               string  `json:"avatar"`
	Title                string  `json:"title"`
	LicenseNo            string  `json:"licenseNo"`
	Specialty            string  `json:"specialty"`
	SpecialtyTags        string  `json:"specialtyTags"`
	YearsOfExperience    int     `json:"yearsOfExperience"`
	Education            string  `json:"education"`
	Introduction         string  `json:"introduction"`
	ConsultationTypes    string  `json:"consultationTypes"`
	WorkDays             string  `json:"workDays"`
	WorkStartTime        string  `json:"workStartTime"`
	WorkEndTime          string  `json:"workEndTime"`
	SessionDuration      int     `json:"sessionDuration"`
	Price                float64 `json:"price"`
	OrganizationID       int64   `json:"organizationId"`
	OrganizationName     string  `json:"organizationName"`
	IsEmergencyAvailable int32   `json:"isEmergencyAvailable"`
	Status               int32   `json:"status"`
	SortOrder            int     `json:"sortOrder"`
}

type CounselorUpdateRequest struct {
	RealName             string  `json:"realName"`
	Gender               int     `json:"gender"`
	Phone                string  `json:"phone"`
	Email                string  `json:"email"`
	Avatar               string  `json:"avatar"`
	Title                string  `json:"title"`
	LicenseNo            string  `json:"licenseNo"`
	Specialty            string  `json:"specialty"`
	SpecialtyTags        string  `json:"specialtyTags"`
	YearsOfExperience    int     `json:"yearsOfExperience"`
	Education            string  `json:"education"`
	Introduction         string  `json:"introduction"`
	ConsultationTypes    string  `json:"consultationTypes"`
	WorkDays             string  `json:"workDays"`
	WorkStartTime        string  `json:"workStartTime"`
	WorkEndTime          string  `json:"workEndTime"`
	SessionDuration      int     `json:"sessionDuration"`
	Price                float64 `json:"price"`
	OrganizationID       int64   `json:"organizationId"`
	OrganizationName     string  `json:"organizationName"`
	IsEmergencyAvailable int32   `json:"isEmergencyAvailable"`
	Status               int32   `json:"status"`
	SortOrder            int     `json:"sortOrder"`
}

type CounselorRecommendRequest struct {
	CaseID       int64  `json:"caseId"`
	DisputeType  string `json:"disputeType"`
	Keywords     string `json:"keywords"`
	Description  string `json:"description"`
	IsEmergency  int32  `json:"isEmergency"`
}

type AppointmentListRequest struct {
	model.BaseQuery
	Status          int32  `form:"status"`
	CounselorID     int64  `form:"counselorId"`
	CaseID          int64  `form:"caseId"`
	PartyName       string `form:"partyName"`
	IsEmergency     int32  `form:"isEmergency"`
	IsAnonymous     int32  `form:"isAnonymous"`
	StartDate       string `form:"startDate"`
	EndDate         string `form:"endDate"`
	AppointmentDate string `form:"appointmentDate"`
}

type AppointmentCreateRequest struct {
	CounselorID        int64  `json:"counselorId" binding:"required"`
	CaseID             int64  `json:"caseId"`
	PartyID            int64  `json:"partyId"`
	PartyName          string `json:"partyName"`
	PartyPhone         string `json:"partyPhone"`
	PartyIDCard        string `json:"partyIdCard"`
	IsAnonymous        int32  `json:"isAnonymous"`
	AppointmentDate    string `json:"appointmentDate" binding:"required"`
	StartTime          string `json:"startTime" binding:"required"`
	EndTime            string `json:"endTime" binding:"required"`
	ConsultationType   int    `json:"consultationType"`
	ConcernType        string `json:"concernType"`
	ConcernDescription string `json:"concernDescription"`
	IsEmergency        int32  `json:"isEmergency"`
}

type AppointmentUpdateRequest struct {
	Status               int32  `json:"status"`
	CounselorID          int64  `json:"counselorId"`
	AppointmentDate      string `json:"appointmentDate"`
	StartTime            string `json:"startTime"`
	EndTime              string `json:"endTime"`
	ConsultationType     int    `json:"consultationType"`
	ConcernType          string `json:"concernType"`
	ConcernDescription   string `json:"concernDescription"`
	ConsultationSummary  string `json:"consultationSummary"`
	FollowUpSuggestion   string `json:"followUpSuggestion"`
	Location             string `json:"location"`
	CancelReason         string `json:"cancelReason"`
}

type RatingCreateRequest struct {
	AppointmentID     int64  `json:"appointmentId" binding:"required"`
	CounselorID       int64  `json:"counselorId" binding:"required"`
	IsAnonymousRating int32  `json:"isAnonymousRating"`
	OverallScore      int    `json:"overallScore" binding:"required"`
	ProfessionalScore int    `json:"professionalScore"`
	AttitudeScore     int    `json:"attitudeScore"`
	EmpathyScore      int    `json:"empathyScore"`
	HelpfulScore      int    `json:"helpfulScore"`
	Content           string `json:"content"`
	Tags              string `json:"tags"`
}

type ScheduleListRequest struct {
	model.BaseQuery
	CounselorID   int64  `form:"counselorId"`
	ScheduleDate  string `form:"scheduleDate"`
	StartDate     string `form:"startDate"`
	EndDate       string `form:"endDate"`
}

type ScheduleCreateRequest struct {
	CounselorID  int64  `json:"counselorId" binding:"required"`
	ScheduleDate string `json:"scheduleDate" binding:"required"`
	StartTime    string `json:"startTime" binding:"required"`
	EndTime      string `json:"endTime" binding:"required"`
	ScheduleType int    `json:"scheduleType"`
	Title        string `json:"title"`
	Remark       string `json:"remark"`
}

var emergencyKeywords = []string{
	"自杀", "想死", "不想活", "活不下去", "结束生命", "自残", "割腕",
	"跳楼", "跳河", "自伤", "自我伤害", "不想活了", "活够了",
}

func detectEmergency(text string) (bool, []string) {
	if text == "" {
		return false, nil
	}
	found := []string{}
	for _, kw := range emergencyKeywords {
		if strings.Contains(text, kw) {
			found = append(found, kw)
		}
	}
	return len(found) > 0, found
}

func generateAnonymousCode(id int64) string {
	return fmt.Sprintf("ANON-%06d", id%1000000)
}

func GetCounselorList(ctx context.Context, c *app.RequestContext) {
	var req CounselorListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("counselor").Where("deleted_at IS NULL")

	if req.Status > 0 {
		db = db.Where("status = ?", req.Status)
	}
	if req.Specialty != "" {
		db = db.Where("specialty LIKE ? OR specialty_tags LIKE ?",
			"%"+req.Specialty+"%", "%"+req.Specialty+"%")
	}
	if req.IsEmergencyAvailable > 0 {
		db = db.Where("is_emergency_available = ?", req.IsEmergencyAvailable)
	}
	if req.ConsultationType > 0 {
		db = db.Where("consultation_types LIKE ?", "%"+strconv.Itoa(req.ConsultationType)+"%")
	}
	if req.OrganizationID > 0 {
		db = db.Where("org_id = ?", req.OrganizationID)
	}
	if req.Keyword != "" {
		db = db.Where("real_name LIKE ? OR title LIKE ? OR specialty_tags LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order(req.GetSort()).
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	for _, item := range list {
		if status, ok := item["status"].(int32); ok {
			item["status_name"] = constants.CounselorStatusMap[int(status)]
		}
		if specialty, ok := item["specialty"].(string); ok && specialty != "" {
			item["specialty_list"] = strings.Split(specialty, ",")
		}
		if tags, ok := item["specialty_tags"].(string); ok && tags != "" {
			item["specialty_tag_list"] = strings.Split(tags, ",")
		}
		if ctypes, ok := item["consultation_types"].(string); ok && ctypes != "" {
			typeList := []map[string]interface{}{}
			for _, t := range strings.Split(ctypes, ",") {
				ti, _ := strconv.Atoi(t)
				if ti > 0 {
					typeList = append(typeList, map[string]interface{}{
						"type": ti,
						"name": constants.CounselorConsultTypeMap[ti],
					})
				}
			}
			item["consultation_type_list"] = typeList
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetCounselorDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var counselor map[string]interface{}
	database.GetDB().Table("counselor").
		Where("id = ? AND deleted_at IS NULL", id).
		Find(&counselor)

	if counselor == nil {
		c.JSON(http.StatusNotFound, response.NotFound("心理咨询师不存在"))
		return
	}

	if status, ok := counselor["status"].(int32); ok {
		counselor["status_name"] = constants.CounselorStatusMap[int(status)]
	}
	if specialty, ok := counselor["specialty"].(string); ok && specialty != "" {
		counselor["specialty_list"] = strings.Split(specialty, ",")
	}
	if tags, ok := counselor["specialty_tags"].(string); ok && tags != "" {
		counselor["specialty_tag_list"] = strings.Split(tags, ",")
	}
	if ctypes, ok := counselor["consultation_types"].(string); ok && ctypes != "" {
		typeList := []map[string]interface{}{}
		for _, t := range strings.Split(ctypes, ",") {
			ti, _ := strconv.Atoi(t)
			if ti > 0 {
				typeList = append(typeList, map[string]interface{}{
					"type": ti,
					"name": constants.CounselorConsultTypeMap[ti],
				})
			}
		}
		counselor["consultation_type_list"] = typeList
	}
	if workDays, ok := counselor["work_days"].(string); ok && workDays != "" {
		counselor["work_day_list"] = strings.Split(workDays, ",")
	}

	var ratings []map[string]interface{}
	database.GetDB().Table("counselor_rating").
		Where("counselor_id = ? AND status = 1 AND deleted_at IS NULL", id).
		Order("created_at DESC").
		Limit(10).
		Find(&ratings)

	for _, r := range ratings {
		if isAnon, ok := r["is_anonymous_rating"].(int32); ok && isAnon == 1 {
			r["rater_name"] = "匿名用户"
		}
		if tags, ok := r["tags"].(string); ok && tags != "" {
			r["tag_list"] = strings.Split(tags, ",")
		}
	}
	counselor["recent_ratings"] = ratings

	c.JSON(http.StatusOK, response.Success(counselor))
}

func CreateCounselor(ctx context.Context, c *app.RequestContext) {
	var req CounselorCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	counselorNo := fmt.Sprintf("CS-%s-%03d", time.Now().Format("2006"), utils.GenerateID()%1000)

	counselorID := utils.GenerateID()
	counselorData := map[string]interface{}{
		"id":                    counselorID,
		"counselor_no":          counselorNo,
		"real_name":             req.RealName,
		"gender":                req.Gender,
		"phone":                 req.Phone,
		"email":                 req.Email,
		"avatar":                req.Avatar,
		"title":                 req.Title,
		"license_no":            req.LicenseNo,
		"specialty":             req.Specialty,
		"specialty_tags":        req.SpecialtyTags,
		"years_of_experience":   req.YearsOfExperience,
		"education":             req.Education,
		"introduction":          req.Introduction,
		"consultation_types":    req.ConsultationTypes,
		"work_days":             req.WorkDays,
		"work_start_time":       req.WorkStartTime,
		"work_end_time":         req.WorkEndTime,
		"session_duration":      req.SessionDuration,
		"price":                 req.Price,
		"org_id":                req.OrganizationID,
		"org_name":              req.OrganizationName,
		"is_emergency_available": req.IsEmergencyAvailable,
		"status":                req.Status,
		"sort_order":            req.SortOrder,
		"created_by":            userInfo.UserID,
	}

	if err := database.GetDB().Table("counselor").Create(counselorData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("创建心理咨询师失败"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id":          counselorID,
		"counselorNo": counselorNo,
	}, "创建成功"))
}

func UpdateCounselor(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req CounselorUpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	updates := map[string]interface{}{}
	if req.RealName != "" {
		updates["real_name"] = req.RealName
	}
	if req.Gender > 0 {
		updates["gender"] = req.Gender
	}
	if req.Phone != "" {
		updates["phone"] = req.Phone
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.LicenseNo != "" {
		updates["license_no"] = req.LicenseNo
	}
	if req.Specialty != "" {
		updates["specialty"] = req.Specialty
	}
	if req.SpecialtyTags != "" {
		updates["specialty_tags"] = req.SpecialtyTags
	}
	if req.YearsOfExperience > 0 {
		updates["years_of_experience"] = req.YearsOfExperience
	}
	if req.Education != "" {
		updates["education"] = req.Education
	}
	if req.Introduction != "" {
		updates["introduction"] = req.Introduction
	}
	if req.ConsultationTypes != "" {
		updates["consultation_types"] = req.ConsultationTypes
	}
	if req.WorkDays != "" {
		updates["work_days"] = req.WorkDays
	}
	if req.WorkStartTime != "" {
		updates["work_start_time"] = req.WorkStartTime
	}
	if req.WorkEndTime != "" {
		updates["work_end_time"] = req.WorkEndTime
	}
	if req.SessionDuration > 0 {
		updates["session_duration"] = req.SessionDuration
	}
	if req.Price >= 0 {
		updates["price"] = req.Price
	}
	if req.OrganizationID > 0 {
		updates["org_id"] = req.OrganizationID
	}
	if req.OrganizationName != "" {
		updates["org_name"] = req.OrganizationName
	}
	updates["is_emergency_available"] = req.IsEmergencyAvailable
	if req.Status > 0 {
		updates["status"] = req.Status
	}
	updates["sort_order"] = req.SortOrder

	if err := database.GetDB().Table("counselor").
		Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("更新失败"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "更新成功"))
}

func DeleteCounselor(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := database.GetDB().Table("counselor").
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("删除失败"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "删除成功"))
}

func RecommendCounselors(ctx context.Context, c *app.RequestContext) {
	var req CounselorRecommendRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("counselor").
		Where("status = ? AND deleted_at IS NULL", constants.CounselorStatusEnabled)

	if req.IsEmergency == 1 {
		db = db.Where("is_emergency_available = 1")
	}

	matchKeywords := []string{}
	if req.DisputeType != "" {
		matchKeywords = append(matchKeywords, req.DisputeType)
	}
	if req.Keywords != "" {
		matchKeywords = append(matchKeywords, strings.Split(req.Keywords, ",")...)
	}

	if len(matchKeywords) > 0 {
		conditions := []string{}
		params := []interface{}{}
		for _, kw := range matchKeywords {
			kw = strings.TrimSpace(kw)
			if kw != "" {
				conditions = append(conditions, "(specialty LIKE ? OR specialty_tags LIKE ?)")
				params = append(params, "%"+kw+"%", "%"+kw+"%")
			}
		}
		if len(conditions) > 0 {
			db = db.Where(strings.Join(conditions, " OR "), params...)
		}
	}

	var list []map[string]interface{}
	db.Order("rating_avg DESC, completed_count DESC, sort_order ASC").
		Limit(10).
		Find(&list)

	for _, item := range list {
		if tags, ok := item["specialty_tags"].(string); ok && tags != "" {
			item["specialty_tag_list"] = strings.Split(tags, ",")
		}
		if specialty, ok := item["specialty"].(string); ok && specialty != "" {
			item["specialty_list"] = strings.Split(specialty, ",")
		}
	}

	c.JSON(http.StatusOK, response.Success(list))
}

func GetCounselorAvailableSlots(ctx context.Context, c *app.RequestContext) {
	counselorID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	var counselor struct {
		WorkStartTime   string `gorm:"column:work_start_time"`
		WorkEndTime     string `gorm:"column:work_end_time"`
		WorkDays        string `gorm:"column:work_days"`
		SessionDuration int    `gorm:"column:session_duration"`
	}
	database.GetDB().Table("counselor").
		Select("work_start_time, work_end_time, work_days, session_duration").
		Where("id = ?", counselorID).
		First(&counselor)

	if counselor.WorkStartTime == "" {
		counselor.WorkStartTime = "09:00:00"
		counselor.WorkEndTime = "18:00:00"
		counselor.SessionDuration = 50
	}

	weekday := int(time.Now().Weekday())
	if weekday == 0 {
		weekday = 7
	}
	workDayAllowed := true
	if counselor.WorkDays != "" {
		days := strings.Split(counselor.WorkDays, ",")
		workDayAllowed = false
		for _, d := range days {
			di, _ := strconv.Atoi(d)
			if di == weekday {
				workDayAllowed = true
				break
			}
		}
	}

	var existingAppointments []map[string]interface{}
	database.GetDB().Table("counselor_appointment").
		Select("start_time, end_time").
		Where("counselor_id = ? AND appointment_date = ? AND status IN ? AND deleted_at IS NULL",
			counselorID, date, []int32{
				constants.CounselorAppointmentStatusPending,
				constants.CounselorAppointmentStatusConfirmed,
				constants.CounselorAppointmentStatusOngoing,
			}).
		Find(&existingAppointments)

	var existingSchedules []map[string]interface{}
	database.GetDB().Table("counselor_schedule").
		Select("start_time, end_time, schedule_type").
		Where("counselor_id = ? AND schedule_date = ? AND deleted_at IS NULL", counselorID, date).
		Find(&existingSchedules)

	isTimeOverlap := func(start, end, existStart, existEnd string) bool {
		return start < existEnd && end > existStart
	}

	slots := []map[string]interface{}{}
	if workDayAllowed {
		startH, _ := strconv.Atoi(strings.Split(counselor.WorkStartTime, ":")[0])
		startM, _ := strconv.Atoi(strings.Split(counselor.WorkStartTime, ":")[1])
		endH, _ := strconv.Atoi(strings.Split(counselor.WorkEndTime, ":")[0])
		endM, _ := strconv.Atoi(strings.Split(counselor.WorkEndTime, ":")[1])

		startMinutes := startH*60 + startM
		endMinutes := endH*60 + endM
		duration := counselor.SessionDuration
		if duration <= 0 {
			duration = 50
		}
		breakMinutes := 10

		for cur := startMinutes; cur+duration <= endMinutes; cur += duration + breakMinutes {
			sh := cur / 60
			sm := cur % 60
			eh := (cur + duration) / 60
			em := (cur + duration) % 60
			slotStart := fmt.Sprintf("%02d:%02d:00", sh, sm)
			slotEnd := fmt.Sprintf("%02d:%02d:00", eh, em)

			available := true
			for _, apt := range existingAppointments {
				es, _ := apt["start_time"].(string)
				ee, _ := apt["end_time"].(string)
				if isTimeOverlap(slotStart, slotEnd, es, ee) {
					available = false
					break
				}
			}
			if available {
				for _, sch := range existingSchedules {
					es, _ := sch["start_time"].(string)
					ee, _ := sch["end_time"].(string)
					if isTimeOverlap(slotStart, slotEnd, es, ee) {
						available = false
						break
					}
				}
			}

			slots = append(slots, map[string]interface{}{
				"startTime": slotStart,
				"endTime":   slotEnd,
				"available": available,
			})
		}
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"date":  date,
		"slots": slots,
	}))
}

func GetAppointmentList(ctx context.Context, c *app.RequestContext) {
	var req AppointmentListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	db := database.GetDB().Table("counselor_appointment ca").
		Select("ca.*, c.real_name as counselor_real_name, c.title as counselor_title, c.avatar as counselor_avatar").
		Joins("LEFT JOIN counselor c ON ca.counselor_id = c.id").
		Where("ca.deleted_at IS NULL")

	if userInfo.Role == constants.RoleMediator {
		db = db.Where("ca.created_by = ?", userInfo.UserID)
	}

	if req.Status > 0 {
		db = db.Where("ca.status = ?", req.Status)
	}
	if req.CounselorID > 0 {
		db = db.Where("ca.counselor_id = ?", req.CounselorID)
	}
	if req.CaseID > 0 {
		db = db.Where("ca.case_id = ?", req.CaseID)
	}
	if req.PartyName != "" {
		db = db.Where("ca.party_name LIKE ?", "%"+req.PartyName+"%")
	}
	if req.IsEmergency > 0 {
		db = db.Where("ca.is_emergency = ?", req.IsEmergency)
	}
	if req.AppointmentDate != "" {
		db = db.Where("ca.appointment_date = ?", req.AppointmentDate)
	}
	if req.StartDate != "" {
		db = db.Where("ca.appointment_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		db = db.Where("ca.appointment_date <= ?", req.EndDate)
	}
	if req.Keyword != "" {
		db = db.Where("ca.appointment_no LIKE ? OR ca.party_name LIKE ? OR ca.concern_description LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order(req.GetSort()).
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	for _, item := range list {
		if status, ok := item["status"].(int32); ok {
			item["status_name"] = constants.CounselorAppointmentStatusMap[int(status)]
		}
		if ctype, ok := item["consultation_type"].(int); ok {
			item["consultation_type_name"] = constants.CounselorConsultTypeMap[ctype]
		} else if ctype, ok := item["consultation_type"].(int32); ok {
			item["consultation_type_name"] = constants.CounselorConsultTypeMap[int(ctype)]
		}
		if elevel, ok := item["emergency_level"].(int); ok {
			item["emergency_level_name"] = constants.CounselorEmergencyLevelMap[elevel]
		} else if elevel, ok := item["emergency_level"].(int32); ok {
			item["emergency_level_name"] = constants.CounselorEmergencyLevelMap[int(elevel)]
		}
		if isAnon, ok := item["is_anonymous"].(int32); ok && isAnon == 1 {
			if code, ok := item["anonymous_code"].(string); ok && code != "" {
				item["party_name_display"] = code
			} else {
				item["party_name_display"] = "匿名用户"
			}
			item["party_phone_display"] = "***"
			item["party_id_card_display"] = "***"
		} else {
			if pname, ok := item["party_name"].(string); ok {
				item["party_name_display"] = pname
			}
			if pphone, ok := item["party_phone"].(string); ok {
				item["party_phone_display"] = pphone
			}
			if pidcard, ok := item["party_id_card"].(string); ok {
				item["party_id_card_display"] = pidcard
			}
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetAppointmentDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var apt map[string]interface{}
	database.GetDB().Table("counselor_appointment ca").
		Select("ca.*, c.real_name as counselor_real_name, c.title as counselor_title, c.avatar as counselor_avatar, c.phone as counselor_phone").
		Joins("LEFT JOIN counselor c ON ca.counselor_id = c.id").
		Where("ca.id = ? AND ca.deleted_at IS NULL", id).
		Find(&apt)

	if apt == nil {
		c.JSON(http.StatusNotFound, response.NotFound("预约不存在"))
		return
	}

	if status, ok := apt["status"].(int32); ok {
		apt["status_name"] = constants.CounselorAppointmentStatusMap[int(status)]
	}
	if ctype, ok := apt["consultation_type"].(int32); ok {
		apt["consultation_type_name"] = constants.CounselorConsultTypeMap[int(ctype)]
	}
	if elevel, ok := apt["emergency_level"].(int32); ok {
		apt["emergency_level_name"] = constants.CounselorEmergencyLevelMap[int(elevel)]
	}
	if isAnon, ok := apt["is_anonymous"].(int32); ok && isAnon == 1 {
		if code, ok := apt["anonymous_code"].(string); ok && code != "" {
			apt["party_name_display"] = code
		} else {
			apt["party_name_display"] = "匿名用户"
		}
		apt["party_phone_display"] = "***"
		apt["party_id_card_display"] = "***"
	} else {
		apt["party_name_display"] = apt["party_name"]
		apt["party_phone_display"] = apt["party_phone"]
		apt["party_id_card_display"] = apt["party_id_card"]
	}

	c.JSON(http.StatusOK, response.Success(apt))
}

func CreateAppointment(ctx context.Context, c *app.RequestContext) {
	var req AppointmentCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var counselor struct {
		RealName string `gorm:"column:real_name"`
		Status   int32  `gorm:"column:status"`
	}
	database.GetDB().Table("counselor").
		Select("real_name, status").
		Where("id = ? AND deleted_at IS NULL", req.CounselorID).
		First(&counselor)

	if counselor.Status != constants.CounselorStatusEnabled {
		c.JSON(http.StatusBadRequest, response.BadRequest("该心理咨询师暂不可预约"))
		return
	}

	var existingCount int64
	database.GetDB().Table("counselor_appointment").
		Where("counselor_id = ? AND appointment_date = ? AND status IN ? AND deleted_at IS NULL AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?) OR (start_time >= ? AND end_time <= ?))",
			req.CounselorID, req.AppointmentDate,
			[]int32{constants.CounselorAppointmentStatusPending, constants.CounselorAppointmentStatusConfirmed, constants.CounselorAppointmentStatusOngoing},
			req.StartTime, req.StartTime,
			req.EndTime, req.EndTime,
			req.StartTime, req.EndTime).
		Count(&existingCount)

	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该时间段已被预约，请选择其他时间"))
		return
	}

	isEmergency := req.IsEmergency
	emergencyLevel := 0
	emergencyWords := ""
	if isEmergency == 0 {
		isEm, found := detectEmergency(req.ConcernDescription)
		if isEm {
			isEmergency = 1
			emergencyWords = strings.Join(found, ",")
			if len(found) >= 3 {
				emergencyLevel = constants.CounselorEmergencyLevelHighRisk
			} else if len(found) >= 2 {
				emergencyLevel = constants.CounselorEmergencyLevelUrgent
			} else {
				emergencyLevel = constants.CounselorEmergencyLevelAttention
			}
		}
	} else if isEmergency == 1 {
		emergencyLevel = constants.CounselorEmergencyLevelUrgent
	}

	aptID := utils.GenerateID()
	aptNo := fmt.Sprintf("APT-%s-%06d", time.Now().Format("20060102"), aptID%1000000)

	anonymousCode := ""
	if req.IsAnonymous == 1 {
		anonymousCode = generateAnonymousCode(aptID)
	}

	aptData := map[string]interface{}{
		"id":                      aptID,
		"appointment_no":          aptNo,
		"counselor_id":            req.CounselorID,
		"counselor_name":          counselor.RealName,
		"case_id":                 req.CaseID,
		"party_id":                req.PartyID,
		"party_name":              req.PartyName,
		"party_phone":             req.PartyPhone,
		"party_id_card":           req.PartyIDCard,
		"is_anonymous":            req.IsAnonymous,
		"anonymous_code":          anonymousCode,
		"appointment_date":        req.AppointmentDate,
		"start_time":              req.StartTime,
		"end_time":                req.EndTime,
		"consultation_type":       req.ConsultationType,
		"appointment_source":      constants.CounselorAppointmentSourceAdmin,
		"is_emergency":            isEmergency,
		"emergency_trigger_words": emergencyWords,
		"emergency_level":         emergencyLevel,
		"concern_type":            req.ConcernType,
		"concern_description":     req.ConcernDescription,
		"status":                  constants.CounselorAppointmentStatusPending,
		"created_by":              userInfo.UserID,
		"created_by_name":         userInfo.RealName,
	}

	tx := database.GetDB().Begin()

	if err := tx.Table("counselor_appointment").Create(aptData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ServerError("创建预约失败"))
		return
	}

	if err := tx.Table("counselor_schedule").Create(map[string]interface{}{
		"counselor_id":   req.CounselorID,
		"schedule_date":  req.AppointmentDate,
		"start_time":     req.StartTime,
		"end_time":       req.EndTime,
		"schedule_type":  constants.CounselorScheduleTypeAppointment,
		"title":          fmt.Sprintf("心理咨询预约 - %s", req.PartyName),
		"appointment_id": aptID,
		"created_by":     userInfo.UserID,
	}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ServerError("创建预约失败"))
		return
	}

	tx.Table("counselor").
		Where("id = ?", req.CounselorID).
		UpdateColumn("appointment_count", gorm.Expr("appointment_count + 1"))

	tx.Commit()

	result := map[string]interface{}{
		"id":              aptID,
		"appointmentNo":   aptNo,
		"isEmergency":     isEmergency,
		"emergencyLevel":  emergencyLevel,
		"emergencyWords":  emergencyWords,
		"anonymousCode":   anonymousCode,
	}
	if isEmergency == 1 {
		result["warning"] = "检测到紧急心理风险，已标记为紧急预约，请优先处理！"
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(result, "预约创建成功"))
}

func UpdateAppointment(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req AppointmentUpdateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	now := time.Now()

	var oldApt struct {
		Status          int32  `gorm:"column:status"`
		CounselorID     int64  `gorm:"column:counselor_id"`
		AppointmentDate string `gorm:"column:appointment_date"`
		StartTime       string `gorm:"column:start_time"`
		EndTime         string `gorm:"column:end_time"`
	}
	database.GetDB().Table("counselor_appointment").
		Select("status, counselor_id, appointment_date, start_time, end_time").
		Where("id = ?", id).
		First(&oldApt)

	updates := map[string]interface{}{}

	if req.Status > 0 {
		updates["status"] = req.Status
		switch req.Status {
		case constants.CounselorAppointmentStatusConfirmed:
			updates["confirmed_by"] = userInfo.UserID
			updates["confirmed_at"] = now
		case constants.CounselorAppointmentStatusOngoing:
			updates["started_at"] = now
		case constants.CounselorAppointmentStatusCompleted:
			updates["completed_at"] = now
			database.GetDB().Table("counselor").
				Where("id = ?", oldApt.CounselorID).
				UpdateColumn("completed_count", gorm.Expr("completed_count + 1"))
		case constants.CounselorAppointmentStatusCancelled:
			updates["cancelled_by"] = userInfo.UserID
			updates["cancelled_at"] = now
			updates["cancel_reason"] = req.CancelReason
			database.GetDB().Table("counselor_schedule").
				Where("appointment_id = ?", id).
				Update("deleted_at", now)
		}
	}

	if req.CounselorID > 0 {
		updates["counselor_id"] = req.CounselorID
	}
	if req.AppointmentDate != "" {
		updates["appointment_date"] = req.AppointmentDate
	}
	if req.StartTime != "" {
		updates["start_time"] = req.StartTime
	}
	if req.EndTime != "" {
		updates["end_time"] = req.EndTime
	}
	if req.ConsultationType > 0 {
		updates["consultation_type"] = req.ConsultationType
	}
	if req.ConcernType != "" {
		updates["concern_type"] = req.ConcernType
	}
	if req.ConcernDescription != "" {
		updates["concern_description"] = req.ConcernDescription
	}
	if req.ConsultationSummary != "" {
		updates["consultation_summary"] = req.ConsultationSummary
	}
	if req.FollowUpSuggestion != "" {
		updates["follow_up_suggestion"] = req.FollowUpSuggestion
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}

	if err := database.GetDB().Table("counselor_appointment").
		Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("更新失败"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "更新成功"))
}

func CancelAppointment(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	now := time.Now()

	tx := database.GetDB().Begin()

	tx.Table("counselor_appointment").
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":        constants.CounselorAppointmentStatusCancelled,
			"cancel_reason": req.Reason,
			"cancelled_by":  userInfo.UserID,
			"cancelled_at":  now,
		})

	var counselorID int64
	database.GetDB().Table("counselor_appointment").
		Select("counselor_id").
		Where("id = ?", id).
		Row().Scan(&counselorID)

	tx.Table("counselor_schedule").
		Where("appointment_id = ?", id).
		Update("deleted_at", now)

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "取消成功"))
}

func GetCounselorRatingList(ctx context.Context, c *app.RequestContext) {
	counselorID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	db := database.GetDB().Table("counselor_rating").
		Where("counselor_id = ? AND status = 1 AND deleted_at IS NULL", counselorID)

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&list)

	for _, item := range list {
		if isAnon, ok := item["is_anonymous_rating"].(int32); ok && isAnon == 1 {
			item["rater_name"] = "匿名用户"
		}
		if tags, ok := item["tags"].(string); ok && tags != "" {
			item["tag_list"] = strings.Split(tags, ",")
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func CreateRating(ctx context.Context, c *app.RequestContext) {
	var req RatingCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var aptStatus int32
	database.GetDB().Table("counselor_appointment").
		Select("status").
		Where("id = ?", req.AppointmentID).
		Row().Scan(&aptStatus)

	if aptStatus != constants.CounselorAppointmentStatusCompleted {
		c.JSON(http.StatusBadRequest, response.BadRequest("只有已完成的咨询才能评价"))
		return
	}

	var existingCount int64
	database.GetDB().Table("counselor_rating").
		Where("appointment_id = ? AND deleted_at IS NULL", req.AppointmentID).
		Count(&existingCount)

	if existingCount > 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("该预约已评价，请勿重复提交"))
		return
	}

	ratingID := utils.GenerateID()
	raterName := userInfo.RealName
	if req.IsAnonymousRating == 1 {
		raterName = ""
	}

	ratingData := map[string]interface{}{
		"id":                  ratingID,
		"appointment_id":      req.AppointmentID,
		"counselor_id":        req.CounselorID,
		"rater_id":            userInfo.UserID,
		"rater_name":          raterName,
		"is_anonymous_rating": req.IsAnonymousRating,
		"overall_score":       req.OverallScore,
		"professional_score":  req.ProfessionalScore,
		"attitude_score":      req.AttitudeScore,
		"empathy_score":       req.EmpathyScore,
		"helpful_score":       req.HelpfulScore,
		"content":             req.Content,
		"tags":                req.Tags,
	}

	tx := database.GetDB().Begin()

	if err := tx.Table("counselor_rating").Create(ratingData).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, response.ServerError("提交评价失败"))
		return
	}

	tx.Table("counselor_appointment").
		Where("id = ?", req.AppointmentID).
		Update("rating_submitted", 1)

	var avgRating float64
	var ratingCount int64
	tx.Table("counselor_rating").
		Select("AVG(overall_score), COUNT(*)").
		Where("counselor_id = ? AND status = 1 AND deleted_at IS NULL", req.CounselorID).
		Row().Scan(&avgRating, &ratingCount)

	tx.Table("counselor").
		Where("id = ?", req.CounselorID).
		Updates(map[string]interface{}{
			"rating_avg":   avgRating,
			"rating_count": ratingCount,
		})

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id": ratingID,
	}, "评价提交成功"))
}

func GetScheduleList(ctx context.Context, c *app.RequestContext) {
	var req ScheduleListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("counselor_schedule").Where("deleted_at IS NULL")

	if req.CounselorID > 0 {
		db = db.Where("counselor_id = ?", req.CounselorID)
	}
	if req.ScheduleDate != "" {
		db = db.Where("schedule_date = ?", req.ScheduleDate)
	}
	if req.StartDate != "" {
		db = db.Where("schedule_date >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		db = db.Where("schedule_date <= ?", req.EndDate)
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("schedule_date ASC, start_time ASC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	for _, item := range list {
		if stype, ok := item["schedule_type"].(int32); ok {
			item["schedule_type_name"] = constants.CounselorScheduleTypeMap[int(stype)]
		} else if stype, ok := item["schedule_type"].(int); ok {
			item["schedule_type_name"] = constants.CounselorScheduleTypeMap[stype]
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func CreateSchedule(ctx context.Context, c *app.RequestContext) {
	var req ScheduleCreateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	scheduleID := utils.GenerateID()
	scheduleData := map[string]interface{}{
		"id":            scheduleID,
		"counselor_id":  req.CounselorID,
		"schedule_date": req.ScheduleDate,
		"start_time":    req.StartTime,
		"end_time":      req.EndTime,
		"schedule_type": req.ScheduleType,
		"title":         req.Title,
		"remark":        req.Remark,
		"created_by":    userInfo.UserID,
	}

	if err := database.GetDB().Table("counselor_schedule").Create(scheduleData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("创建日程失败"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id": scheduleID,
	}, "创建成功"))
}

func DeleteSchedule(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := database.GetDB().Table("counselor_schedule").
		Where("id = ?", id).
		Update("deleted_at", time.Now()).Error; err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("删除失败"))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "删除成功"))
}

func GetCounselorStats(ctx context.Context, c *app.RequestContext) {
	counselorID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var totalAppointments int64
	var completedAppointments int64
	var avgRating float64
	var totalRatings int64

	database.GetDB().Table("counselor_appointment").
		Where("counselor_id = ? AND deleted_at IS NULL", counselorID).
		Count(&totalAppointments)

	database.GetDB().Table("counselor_appointment").
		Where("counselor_id = ? AND status = ? AND deleted_at IS NULL",
			counselorID, constants.CounselorAppointmentStatusCompleted).
		Count(&completedAppointments)

	database.GetDB().Table("counselor_rating").
		Select("AVG(overall_score), COUNT(*)").
		Where("counselor_id = ? AND status = 1 AND deleted_at IS NULL", counselorID).
		Row().Scan(&avgRating, &totalRatings)

	ratingDistribution := []map[string]interface{}{}
	for i := 5; i >= 1; i-- {
		var cnt int64
		database.GetDB().Table("counselor_rating").
			Where("counselor_id = ? AND overall_score = ? AND status = 1 AND deleted_at IS NULL", counselorID, i).
			Count(&cnt)
		ratingDistribution = append(ratingDistribution, map[string]interface{}{
			"score": i,
			"count": cnt,
		})
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"totalAppointments":     totalAppointments,
		"completedAppointments": completedAppointments,
		"avgRating":             avgRating,
		"totalRatings":          totalRatings,
		"ratingDistribution":    ratingDistribution,
	}))
}
