package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/amap"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type PatrolServiceImpl struct {
	pointsService service.PointsService
}

func NewPatrolService() service.PatrolService {
	return &PatrolServiceImpl{
		pointsService: NewPointsService(),
	}
}

func (s *PatrolServiceImpl) CreateTask(ctx context.Context, req map[string]interface{}, assignerID int64, assignerName string) (int64, error) {
	tx := database.GetDB().Begin()

	taskNo := fmt.Sprintf("PT%s", utils.GenerateID())
	taskType := int(req["taskType"].(float64))
	pointsReward := s.calculateTaskPoints(taskType)

	task := model.PatrolTask{
		TaskNo:       taskNo,
		Title:        req["title"].(string),
		Description:  req["description"].(string),
		TaskType:     taskType,
		TypeName:     getTaskTypeName(taskType),
		Priority:     int(req["priority"].(float64)),
		PriorityName: getPriorityName(int(req["priority"].(float64))),
		AssigneeID:   int64(req["assigneeId"].(float64)),
		Status:       10,
		StatusName:   "待执行",
		PointsReward: pointsReward,
		StartTime:    parseTime(req["startTime"]),
		EndTime:      parseTime(req["endTime"]),
		AssignerID:   assignerID,
		AssignerName: assignerName,
	}

	if orgID, ok := req["orgId"].(float64); ok {
		task.OrganizationID = int64(orgID)
	}

	if gridCodes, ok := req["gridCodes"].(string); ok {
		task.GridCodes = gridCodes
	}

	if err := tx.Create(&task).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	if pointsList, ok := req["points"].([]interface{}); ok {
		for i, p := range pointsList {
			point := p.(map[string]interface{})
			taskPoint := model.PatrolTaskPoint{
				TaskID:      task.ID,
				PointNo:     fmt.Sprintf("PTP%s%d", utils.GenerateID(), i+1),
				PointName:   point["pointName"].(string),
				Address:     point["address"].(string),
				Longitude:   point["longitude"].(float64),
				Latitude:    point["latitude"].(float64),
				PointType:   int(point["pointType"].(float64)),
				CheckinType: int(point["checkinType"].(float64)),
				CheckinRadius: floatOrDefault(point, "checkinRadius", 200),
				RequiredPhotos: intOrDefault(point, "requiredPhotos", 3),
				SortOrder:   i + 1,
			}
			if err := tx.Create(&taskPoint).Error; err != nil {
				tx.Rollback()
				return 0, err
			}
		}
	}

	var member model.GridMember
	tx.Where("id = ?", task.AssigneeID).First(&member)
	if member.Phone != "" {
		sendTaskNotification(member.Phone, task.Title)
	}

	tx.Commit()
	return task.ID, nil
}

func (s *PatrolServiceImpl) GetTaskList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("patrol_task pt").
		Select("pt.*, gm.real_name as assignee_name, gm.member_no as assignee_member_no")

	if status, ok := query["status"].(float64); ok && int(status) > 0 {
		db = db.Where("pt.status = ?", int(status))
	}
	if assigneeID, ok := query["assigneeId"].(float64); ok && int64(assigneeID) > 0 {
		db = db.Where("pt.assignee_id = ?", int64(assigneeID))
	}
	if taskType, ok := query["taskType"].(float64); ok && int(taskType) > 0 {
		db = db.Where("pt.task_type = ?", int(taskType))
	}
	if priority, ok := query["priority"].(float64); ok && int(priority) > 0 {
		db = db.Where("pt.priority = ?", int(priority))
	}
	if keyword, ok := query["keyword"].(string); ok && keyword != "" {
		db = db.Where("pt.title LIKE ? OR pt.task_no LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if orgID, ok := query["orgId"].(float64); ok && int64(orgID) > 0 {
		db = db.Where("pt.organization_id = ?", int64(orgID))
	}

	db = db.Joins("LEFT JOIN grid_member gm ON gm.id = pt.assignee_id")

	var total int64
	db.Count(&total)

	page := intOrDefault(query, "page", 1)
	pageSize := intOrDefault(query, "pageSize", 20)
	offset := (page - 1) * pageSize

	var tasks []map[string]interface{}
	result := db.Order("pt.created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks)

	return tasks, total, result.Error
}

func (s *PatrolServiceImpl) GetTaskDetail(ctx context.Context, taskID int64) (map[string]interface{}, error) {
	var task model.PatrolTask
	result := database.GetDB().Where("id = ?", taskID).Preload("Points").First(&task)
	if result.Error != nil {
		return nil, result.Error
	}

	var points []map[string]interface{}
	database.GetDB().Table("patrol_task_point").
		Where("task_id = ?", taskID).
		Order("sort_order ASC").
		Find(&points)

	for i, p := range points {
		pointID := p["id"].(int64)
		var checkinCount int64
		database.GetDB().Table("patrol_checkin").
			Where("task_id = ? AND task_point_id = ? AND is_valid = 1", taskID, pointID).
			Count(&checkinCount)
		points[i]["isChecked"] = checkinCount > 0
	}

	taskMap := structToMap(task)
	taskMap["points"] = points

	var visitCount int64
	database.GetDB().Table("patrol_visit_record").
		Where("task_id = ?", taskID).
		Count(&visitCount)
	taskMap["visitCount"] = visitCount

	var dangerCount int64
	database.GetDB().Table("hidden_danger").
		Where("task_id = ?", taskID).
		Count(&dangerCount)
	taskMap["dangerCount"] = dangerCount

	return taskMap, nil
}

func (s *PatrolServiceImpl) UpdateTask(ctx context.Context, taskID int64, req map[string]interface{}) error {
	updates := make(map[string]interface{})
	for k, v := range req {
		if k == "points" {
			continue
		}
		if k == "taskType" {
			taskType := int(v.(float64))
			updates["task_type"] = taskType
			updates["type_name"] = getTaskTypeName(taskType)
			updates["points_reward"] = s.calculateTaskPoints(taskType)
		} else if k == "priority" {
			priority := int(v.(float64))
			updates["priority"] = priority
			updates["priority_name"] = getPriorityName(priority)
		} else if k == "startTime" || k == "endTime" {
			updates[utils.CamelToSnake(k)] = parseTime(v)
		} else {
			updates[utils.CamelToSnake(k)] = v
		}
	}
	return database.GetDB().Table("patrol_task").Where("id = ?", taskID).Updates(updates).Error
}

func (s *PatrolServiceImpl) DeleteTask(ctx context.Context, taskID int64) error {
	tx := database.GetDB().Begin()
	tx.Where("task_id = ?", taskID).Delete(&model.PatrolTaskPoint{})
	tx.Delete(&model.PatrolTask{}, taskID)
	tx.Commit()
	return nil
}

func (s *PatrolServiceImpl) CancelTask(ctx context.Context, taskID int64, reason string) error {
	return database.GetDB().Table("patrol_task").Where("id = ?", taskID).Updates(map[string]interface{}{
		"status":      50,
		"status_name": "已取消",
		"cancel_reason": reason,
	}).Error
}

func (s *PatrolServiceImpl) StartTask(ctx context.Context, taskID int64, memberID int64) error {
	now := time.Now()
	return database.GetDB().Table("patrol_task").
		Where("id = ? AND assignee_id = ? AND status = 10", taskID, memberID).
		Updates(map[string]interface{}{
			"status":       20,
			"status_name":  "进行中",
			"started_at":   now,
			"started_by":   memberID,
		}).Error
}

func (s *PatrolServiceImpl) CompleteTask(ctx context.Context, taskID int64, memberID int64) error {
	var task model.PatrolTask
	result := database.GetDB().Where("id = ?", taskID).First(&task)
	if result.Error != nil {
		return result.Error
	}

	var totalPoints int
	var uncheckedCount int64
	database.GetDB().Table("patrol_task_point").
		Where("task_id = ?", taskID).
		Count(&uncheckedCount)

	var checkedCount int64
	database.GetDB().Table("patrol_checkin pc").
		Where("pc.task_id = ? AND pc.is_valid = 1", taskID).
		Group("pc.task_point_id").
		Count(&checkedCount)

	totalPoints = task.PointsReward

	tx := database.GetDB().Begin()
	now := time.Now()

	err := tx.Table("patrol_task").
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":       30,
			"status_name":  "已完成",
			"completed_at": now,
			"completed_by": memberID,
			"actual_points": totalPoints,
		}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	if totalPoints > 0 {
		err = s.pointsService.AddPoints(ctx, memberID, totalPoints,
			"task_complete", task.TaskNo,
			fmt.Sprintf("完成排查任务: %s", task.Title))
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	tx.Commit()
	return nil
}

func (s *PatrolServiceImpl) PlanRoute(ctx context.Context, startLng, startLat float64, points []map[string]interface{}, strategy int) (map[string]interface{}, error) {
	if strategy <= 0 || strategy > 13 {
		strategy = 10
	}

	routePoints := make([]utils.RoutePoint, 0, len(points))
	for i, p := range points {
		pointType := intOrDefault(p, "pointType", 1)
		priority := intOrDefault(p, "priority", 3)
		routePoints = append(routePoints, utils.RoutePoint{
			Index:     i,
			Lng:       p["longitude"].(float64),
			Lat:       p["latitude"].(float64),
			Name:      p["pointName"].(string),
			Priority:  priority,
			PointType: pointType,
		})
	}

	localResult := utils.OptimizeRoute(startLng, startLat, routePoints, strategy)

	amapClient := amap.GetAmapClient()
	var amapOrderedPoints []amap.OrderedRoutePoint
	for sortedIdx, rp := range localResult.OrderedPoints {
		amapOrderedPoints = append(amapOrderedPoints, amap.OrderedRoutePoint{
			OriginalIndex: rp.Index,
			SortedIndex:   sortedIdx,
			Longitude:     rp.Lng,
			Latitude:      rp.Lat,
			Name:          rp.Name,
			SortOrder:     sortedIdx + 1,
		})
	}

	totalDuration := 0
	var paths []amap.Path
	amapErr := error(nil)

	amapResult, err := amapClient.PlanDrivingRoute(startLng, startLat, amapOrderedPoints, 0)
	if err != nil {
		amapErr = err
	} else {
		totalDuration = amapResult.TotalDuration
		paths = amapResult.Paths
	}

	orderedPoints := make([]map[string]interface{}, 0, len(localResult.OrderedPoints))
	prevLng, prevLat := startLng, startLat
	var distanceFromPrev float64

	for sortedIdx, rp := range localResult.OrderedPoints {
		originalPoint := points[rp.Index]

		if sortedIdx == 0 {
			distanceFromPrev = utils.HaversineDistance(prevLng, prevLat, rp.Lng, rp.Lat)
		} else {
			distanceFromPrev = utils.HaversineDistance(prevLng, prevLat, rp.Lng, rp.Lat)
		}

		durationFromPrev := 0
		if amapResult != nil && sortedIdx < len(amapResult.OrderedPoints) {
			durationFromPrev = amapResult.OrderedPoints[sortedIdx].Duration
		}

		orderedPoints = append(orderedPoints, map[string]interface{}{
			"originalIndex":    rp.Index,
			"sortedIndex":      sortedIdx,
			"pointName":        originalPoint["pointName"],
			"address":          originalPoint["address"],
			"longitude":        rp.Lng,
			"latitude":         rp.Lat,
			"distanceFromPrev": distanceFromPrev,
			"durationFromPrev": durationFromPrev,
		})

		prevLng, prevLat = rp.Lng, rp.Lat
	}

	totalDistance := localResult.TotalDistance
	if amapResult != nil {
		totalDistance = float64(amapResult.TotalDistance)
	}
	if totalDuration == 0 {
		totalDuration = int(totalDistance / 1.39)
	}

	totalTaxiCost := 0.0
	if totalDistance > 0 {
		totalTaxiCost = 13.0 + (totalDistance/1000.0-3)*2.3
		if totalTaxiCost < 13 {
			totalTaxiCost = 13
		}
	}

	routeMap := map[string]interface{}{
		"totalDistance":     totalDistance,
		"totalDuration":     totalDuration,
		"totalTaxiCost":     totalTaxiCost,
		"strategy":          localResult.Strategy,
		"strategyName":      localResult.StrategyName,
		"points":            orderedPoints,
		"paths":             paths,
		"localOptimization": true,
		"amapAvailable":     amapErr == nil,
	}

	if amapErr != nil {
		routeMap["amapError"] = amapErr.Error()
	}

	return routeMap, nil
}

func (s *PatrolServiceImpl) GetMemberTasks(ctx context.Context, memberID int64, status int, page, pageSize int) ([]map[string]interface{}, int64, error) {
	query := make(map[string]interface{})
	query["assigneeId"] = float64(memberID)
	if status > 0 {
		query["status"] = float64(status)
	}
	query["page"] = page
	query["pageSize"] = pageSize
	return s.GetTaskList(ctx, query)
}

func (s *PatrolServiceImpl) GetTaskPoints(ctx context.Context, taskID int64) ([]map[string]interface{}, error) {
	var points []map[string]interface{}
	result := database.GetDB().Table("patrol_task_point").
		Where("task_id = ?", taskID).
		Order("sort_order ASC").
		Find(&points)
	return points, result.Error
}

func (s *PatrolServiceImpl) Checkin(ctx context.Context, req map[string]interface{}, memberID int64, memberName string, ipAddress string) (map[string]interface{}, error) {
	taskID := int64(req["taskId"].(float64))
	taskPointID := int64(req["taskPointId"].(float64))
	longitude := req["longitude"].(float64)
	latitude := req["latitude"].(float64)

	var taskPoint model.PatrolTaskPoint
	result := database.GetDB().Where("id = ?", taskPointID).First(&taskPoint)
	if result.Error != nil {
		return nil, fmt.Errorf("task point not found")
	}

	distance := amap.CalculateDistance(
		longitude, latitude,
		taskPoint.Longitude, taskPoint.Latitude,
	)

	checkinRadius := taskPoint.CheckinRadius
	if checkinRadius <= 0 {
		checkinRadius = 200
	}

	isValid := 1
	invalidReason := ""
	if distance > checkinRadius {
		isValid = 0
		invalidReason = fmt.Sprintf("超出签到范围，距离目标位置%.0f米", distance)
	}

	liveVerifyScore := float64(0)
	isLiveVerified := 0
	if req["livePhotoUrl"] != nil && req["livePhotoUrl"].(string) != "" {
		liveVerifyScore = 95.0
		isLiveVerified = 1
	}

	checkinNo := fmt.Sprintf("CK%s", utils.GenerateID())
	now := time.Now()

	checkin := model.PatrolCheckin{
		CheckinNo:      checkinNo,
		TaskID:         taskID,
		TaskPointID:    taskPointID,
		MemberID:       memberID,
		MemberName:     memberName,
		Longitude:      longitude,
		Latitude:       latitude,
		LocationAccuracy: floatOrDefault(req, "locationAccuracy", 0),
		Address:        stringOrDefault(req, "address", ""),
		PhotoURL:       stringOrDefault(req, "photoUrl", ""),
		LivePhotoURL:   stringOrDefault(req, "livePhotoUrl", ""),
		IsLiveVerified: isLiveVerified,
		LiveVerifyScore: liveVerifyScore,
		CheckinDistance: distance,
		CheckinRadius:  checkinRadius,
		IsValid:        isValid,
		InvalidReason:  invalidReason,
		IPAddress:      ipAddress,
		DeviceInfo:     stringOrDefault(req, "deviceInfo", ""),
		CheckinTime:    &now,
		Remark:         stringOrDefault(req, "remark", ""),
	}

	tx := database.GetDB().Begin()
	if err := tx.Create(&checkin).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var pointsEarned int
	if isValid == 1 {
		pointsEarned = 10
		err := s.pointsService.AddPoints(ctx, memberID, pointsEarned,
			"checkin", checkinNo,
			fmt.Sprintf("签到打卡: %s", taskPoint.PointName))
		if err != nil {
			logger.Error("Failed to add checkin points", logger.Error(err))
		}

		var account model.GridMemberPointsAccount
		tx.Where("member_id = ?", memberID).First(&account)
		if account.ID > 0 {
			today := time.Now().Format("2006-01-02")
			var todayCheckins int64
			tx.Table("patrol_checkin").
				Where("member_id = ? AND DATE(created_at) = ? AND is_valid = 1", memberID, today).
				Count(&todayCheckins)
			if todayCheckins == 1 {
				tx.Model(&account).Update("checkin_days", account.CheckinDays+1)

				if account.CheckinDays+1 >= 7 {
					err = s.pointsService.AddPoints(ctx, memberID, 50,
						"weekly_checkin_bonus", checkinNo,
						"连续7天签到奖励")
				}
				if account.CheckinDays+1 >= 30 {
					err = s.pointsService.AddPoints(ctx, memberID, 200,
						"monthly_checkin_bonus", checkinNo,
						"连续30天签到奖励")
				}
			}
		}
	}

	tx.Commit()

	return map[string]interface{}{
		"id":             checkin.ID,
		"checkinNo":      checkinNo,
		"isValid":        isValid,
		"distance":       distance,
		"pointsEarned":   pointsEarned,
		"isLiveVerified": isLiveVerified,
		"invalidReason":  invalidReason,
		"checkinTime":    now,
	}, nil
}

func (s *PatrolServiceImpl) GetCheckinRecords(ctx context.Context, memberID int64, page, pageSize int) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("patrol_checkin pc").
		Select("pc.*, ptp.point_name, pt.title as task_name").
		Joins("LEFT JOIN patrol_task_point ptp ON ptp.id = pc.task_point_id").
		Joins("LEFT JOIN patrol_task pt ON pt.id = pc.task_id").
		Where("pc.member_id = ?", memberID)

	var total int64
	db.Count(&total)

	offset := (page - 1) * pageSize
	var records []map[string]interface{}
	result := db.Order("pc.created_at DESC").Offset(offset).Limit(pageSize).Find(&records)

	return records, total, result.Error
}

func (s *PatrolServiceImpl) GetCheckinStatistics(ctx context.Context, memberID int64) (map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	monthStart := time.Now().Format("2006-01-01")
	weekStart := time.Now().AddDate(0, 0, -int(time.Now().Weekday())).Format("2006-01-02")

	var todayCount int64
	database.GetDB().Table("patrol_checkin").
		Where("member_id = ? AND DATE(created_at) = ? AND is_valid = 1", memberID, today).
		Count(&todayCount)

	var weekCount int64
	database.GetDB().Table("patrol_checkin").
		Where("member_id = ? AND DATE(created_at) >= ? AND is_valid = 1", memberID, weekStart).
		Count(&weekCount)

	var monthCount int64
	database.GetDB().Table("patrol_checkin").
		Where("member_id = ? AND DATE(created_at) >= ? AND is_valid = 1", memberID, monthStart).
		Count(&monthCount)

	var totalCount int64
	database.GetDB().Table("patrol_checkin").
		Where("member_id = ? AND is_valid = 1", memberID).
		Count(&totalCount)

	return map[string]interface{}{
		"todayCount":  todayCount,
		"weekCount":   weekCount,
		"monthCount":  monthCount,
		"totalCount":  totalCount,
	}, nil
}

func (s *PatrolServiceImpl) CreateVisitRecord(ctx context.Context, req map[string]interface{}, memberID int64, memberName string, orgID int64) (int64, error) {
	recordNo := fmt.Sprintf("VR%s", utils.GenerateID())

	record := model.PatrolVisitRecord{
		RecordNo:       recordNo,
		MemberID:       memberID,
		MemberName:     memberName,
		TaskID:         int64OrDefault(req, "taskId", 0),
		TaskPointID:    int64OrDefault(req, "taskPointId", 0),
		VisitType:      int(req["visitType"].(float64)),
		VisitTypeName:  getVisitTypeName(int(req["visitType"].(float64))),
		VisitObject:    req["visitObject"].(string),
		VisitContent:   req["visitContent"].(string),
		VisitResult:    stringOrDefault(req, "visitResult", ""),
		Longitude:      floatOrDefault(req, "longitude", 0),
		Latitude:       floatOrDefault(req, "latitude", 0),
		Address:        stringOrDefault(req, "address", ""),
		PhotoURLs:      stringOrDefault(req, "photoUrls", ""),
		OrganizationID: orgID,
		Status:         10,
		StatusName:     "待审核",
		Remark:         stringOrDefault(req, "remark", ""),
	}

	if residentID, ok := req["residentId"].(float64); ok {
		record.ResidentID = int64(residentID)
	}

	if caseID, ok := req["disputeCaseId"].(float64); ok {
		record.DisputeCaseID = int64(caseID)
	}

	result := database.GetDB().Create(&record)
	if result.Error != nil {
		return 0, result.Error
	}

	s.pointsService.AddPoints(ctx, memberID, 5,
		"visit_record", recordNo,
		fmt.Sprintf("提交走访记录: %s", record.VisitObject))

	return record.ID, nil
}

func (s *PatrolServiceImpl) GetVisitRecords(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("patrol_visit_record pvr").
		Select("pvr.*, gm.real_name as member_name")

	if memberID, ok := query["memberId"].(float64); ok && int64(memberID) > 0 {
		db = db.Where("pvr.member_id = ?", int64(memberID))
	}
	if status, ok := query["status"].(float64); ok && int(status) > 0 {
		db = db.Where("pvr.status = ?", int(status))
	}
	if visitType, ok := query["visitType"].(float64); ok && int(visitType) > 0 {
		db = db.Where("pvr.visit_type = ?", int(visitType))
	}
	if orgID, ok := query["orgId"].(float64); ok && int64(orgID) > 0 {
		db = db.Where("pvr.organization_id = ?", int64(orgID))
	}
	if startDate, ok := query["startDate"].(string); ok && startDate != "" {
		db = db.Where("DATE(pvr.created_at) >= ?", startDate)
	}
	if endDate, ok := query["endDate"].(string); ok && endDate != "" {
		db = db.Where("DATE(pvr.created_at) <= ?", endDate)
	}

	db = db.Joins("LEFT JOIN grid_member gm ON gm.id = pvr.member_id")

	var total int64
	db.Count(&total)

	page := intOrDefault(query, "page", 1)
	pageSize := intOrDefault(query, "pageSize", 20)
	offset := (page - 1) * pageSize

	var records []map[string]interface{}
	result := db.Order("pvr.created_at DESC").Offset(offset).Limit(pageSize).Find(&records)

	return records, total, result.Error
}

func (s *PatrolServiceImpl) GetVisitRecordDetail(ctx context.Context, id int64) (map[string]interface{}, error) {
	var record map[string]interface{}
	result := database.GetDB().Table("patrol_visit_record pvr").
		Select("pvr.*, gm.real_name as member_name, pt.title as task_name").
		Joins("LEFT JOIN grid_member gm ON gm.id = pvr.member_id").
		Joins("LEFT JOIN patrol_task pt ON pt.id = pvr.task_id").
		Where("pvr.id = ?", id).
		First(&record)
	return record, result.Error
}

func (s *PatrolServiceImpl) UpdateVisitRecord(ctx context.Context, id int64, req map[string]interface{}) error {
	updates := make(map[string]interface{})
	for k, v := range req {
		if k == "visitType" {
			vt := int(v.(float64))
			updates["visit_type"] = vt
			updates["visit_type_name"] = getVisitTypeName(vt)
		} else {
			updates[utils.CamelToSnake(k)] = v
		}
	}
	return database.GetDB().Table("patrol_visit_record").Where("id = ?", id).Updates(updates).Error
}

func (s *PatrolServiceImpl) AuditVisitRecord(ctx context.Context, id int64, status int32, remark string, auditorID int64) error {
	updates := map[string]interface{}{
		"status":       status,
		"status_name":  getVisitStatusName(int(status)),
		"audit_remark": remark,
		"audit_time":   time.Now(),
		"auditor_id":   auditorID,
	}

	if status == 20 {
		var record model.PatrolVisitRecord
		database.GetDB().Where("id = ?", id).First(&record)
		if record.ID > 0 {
			s.pointsService.AddPoints(ctx, record.MemberID, 10,
				"visit_audit_pass", record.RecordNo,
				"走访记录审核通过")
		}
	}

	return database.GetDB().Table("patrol_visit_record").Where("id = ?", id).Updates(updates).Error
}

func (s *PatrolServiceImpl) DeleteVisitRecord(ctx context.Context, id int64) error {
	return database.GetDB().Delete(&model.PatrolVisitRecord{}, id).Error
}

func (s *PatrolServiceImpl) GetVisitStatistics(ctx context.Context, memberID int64) (map[string]interface{}, error) {
	today := time.Now().Format("2006-01-02")
	monthStart := time.Now().Format("2006-01-01")

	var todayCount int64
	database.GetDB().Table("patrol_visit_record").
		Where("member_id = ? AND DATE(created_at) = ?", memberID, today).
		Count(&todayCount)

	var monthCount int64
	database.GetDB().Table("patrol_visit_record").
		Where("member_id = ? AND DATE(created_at) >= ?", memberID, monthStart).
		Count(&monthCount)

	var totalCount int64
	database.GetDB().Table("patrol_visit_record").
		Where("member_id = ?", memberID).
		Count(&totalCount)

	typeCount := make(map[string]int64)
	var typeResults []struct {
		VisitType int   `gorm:"column:visit_type"`
		Count     int64 `gorm:"column:count"`
	}
	database.GetDB().Table("patrol_visit_record").
		Select("visit_type, COUNT(*) as count").
		Where("member_id = ?", memberID).
		Group("visit_type").
		Scan(&typeResults)

	for _, r := range typeResults {
		typeCount[getVisitTypeName(r.VisitType)] = r.Count
	}

	return map[string]interface{}{
		"todayCount": todayCount,
		"monthCount": monthCount,
		"totalCount": totalCount,
		"typeCount":  typeCount,
	}, nil
}

func (s *PatrolServiceImpl) ReportDanger(ctx context.Context, req map[string]interface{}, reporterID int64, reporterName string, orgID int64) (int64, error) {
	dangerNo := fmt.Sprintf("HD%s", utils.GenerateID())

	danger := model.HiddenDanger{
		DangerNo:       dangerNo,
		ReporterID:     reporterID,
		ReporterName:   reporterName,
		TaskID:         int64OrDefault(req, "taskId", 0),
		TaskPointID:    int64OrDefault(req, "taskPointId", 0),
		DangerType:     int(req["dangerType"].(float64)),
		DangerTypeName: getDangerTypeName(int(req["dangerType"].(float64))),
		Level:          int(req["level"].(float64)),
		LevelName:      getDangerLevelName(int(req["level"].(float64))),
		Title:          req["title"].(string),
		Description:    req["description"].(string),
		Longitude:      floatOrDefault(req, "longitude", 0),
		Latitude:       floatOrDefault(req, "latitude", 0),
		Address:        stringOrDefault(req, "address", ""),
		PhotoURLs:      stringOrDefault(req, "photoUrls", ""),
		VideoURL:       stringOrDefault(req, "videoUrl", ""),
		OrganizationID: orgID,
		Status:         10,
		StatusName:     "待处理",
		Source:         2,
		SourceName:     "网格员上报",
	}

	if residentID, ok := req["involvedPerson"].(string); ok {
		danger.InvolvedPerson = residentID
	}

	result := database.GetDB().Create(&danger)
	if result.Error != nil {
		return 0, result.Error
	}

	points := s.calculateDangerPoints(danger.Level)
	s.pointsService.AddPoints(ctx, reporterID, points,
		"danger_report", dangerNo,
		fmt.Sprintf("隐患上报: %s", danger.Title))

	return danger.ID, nil
}

func (s *PatrolServiceImpl) GetDangerList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("hidden_danger hd").
		Select("hd.*, gm.real_name as reporter_name")

	if reporterID, ok := query["reporterId"].(float64); ok && int64(reporterID) > 0 {
		db = db.Where("hd.reporter_id = ?", int64(reporterID))
	}
	if status, ok := query["status"].(float64); ok && int(status) > 0 {
		db = db.Where("hd.status = ?", int(status))
	}
	if dangerType, ok := query["dangerType"].(float64); ok && int(dangerType) > 0 {
		db = db.Where("hd.danger_type = ?", int(dangerType))
	}
	if level, ok := query["level"].(float64); ok && int(level) > 0 {
		db = db.Where("hd.level = ?", int(level))
	}
	if orgID, ok := query["orgId"].(float64); ok && int64(orgID) > 0 {
		db = db.Where("hd.organization_id = ?", int64(orgID))
	}
	if keyword, ok := query["keyword"].(string); ok && keyword != "" {
		db = db.Where("hd.title LIKE ? OR hd.danger_no LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	db = db.Joins("LEFT JOIN grid_member gm ON gm.id = hd.reporter_id")

	var total int64
	db.Count(&total)

	page := intOrDefault(query, "page", 1)
	pageSize := intOrDefault(query, "pageSize", 20)
	offset := (page - 1) * pageSize

	var list []map[string]interface{}
	result := db.Order("hd.created_at DESC").Offset(offset).Limit(pageSize).Find(&list)

	return list, total, result.Error
}

func (s *PatrolServiceImpl) GetDangerDetail(ctx context.Context, id int64) (map[string]interface{}, error) {
	var detail map[string]interface{}
	result := database.GetDB().Table("hidden_danger hd").
		Select("hd.*, gm.real_name as reporter_name").
		Joins("LEFT JOIN grid_member gm ON gm.id = hd.reporter_id").
		Where("hd.id = ?", id).
		First(&detail)
	return detail, result.Error
}

func (s *PatrolServiceImpl) HandleDanger(ctx context.Context, id int64, status int32, handlerID int64, handlerName string, result string) error {
	updates := map[string]interface{}{
		"status":        status,
		"status_name":   getDangerStatusName(int(status)),
		"handler_id":    handlerID,
		"handler_name":  handlerName,
		"handle_result": result,
		"handled_at":    time.Now(),
	}
	return database.GetDB().Table("hidden_danger").Where("id = ?", id).Updates(updates).Error
}

func (s *PatrolServiceImpl) GetDangerStatistics(ctx context.Context) (map[string]interface{}, error) {
	var total int64
	database.GetDB().Table("hidden_danger").Count(&total)

	statusStats := make(map[string]int64)
	var statusResults []struct {
		Status int   `gorm:"column:status"`
		Count  int64 `gorm:"column:count"`
	}
	database.GetDB().Table("hidden_danger").
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusResults)

	for _, r := range statusResults {
		statusStats[getDangerStatusName(r.Status)] = r.Count
	}

	typeStats := make(map[string]int64)
	var typeResults []struct {
		DangerType int   `gorm:"column:danger_type"`
		Count      int64 `gorm:"column:count"`
	}
	database.GetDB().Table("hidden_danger").
		Select("danger_type, COUNT(*) as count").
		Group("danger_type").
		Scan(&typeResults)

	for _, r := range typeResults {
		typeStats[getDangerTypeName(r.DangerType)] = r.Count
	}

	return map[string]interface{}{
		"total":       total,
		"statusStats": statusStats,
		"typeStats":   typeStats,
	}, nil
}

func (s *PatrolServiceImpl) GetMemberList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("grid_member gm").
		Select("gm.*, u.username, u.phone")

	if orgID, ok := query["orgId"].(float64); ok && int64(orgID) > 0 {
		db = db.Where("gm.organization_id = ?", int64(orgID))
	}
	if status, ok := query["status"].(float64); ok && int(status) > 0 {
		db = db.Where("gm.status = ?", int(status))
	}
	if keyword, ok := query["keyword"].(string); ok && keyword != "" {
		db = db.Where("gm.real_name LIKE ? OR gm.member_no LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	db = db.Joins("LEFT JOIN user u ON u.id = gm.user_id")

	var total int64
	db.Count(&total)

	page := intOrDefault(query, "page", 1)
	pageSize := intOrDefault(query, "pageSize", 20)
	offset := (page - 1) * pageSize

	var list []map[string]interface{}
	result := db.Order("gm.created_at DESC").Offset(offset).Limit(pageSize).Find(&list)

	return list, total, result.Error
}

func (s *PatrolServiceImpl) GetMemberDetail(ctx context.Context, id int64) (map[string]interface{}, error) {
	var detail map[string]interface{}
	result := database.GetDB().Table("grid_member gm").
		Select("gm.*, u.username, u.phone").
		Joins("LEFT JOIN user u ON u.id = gm.user_id").
		Where("gm.id = ?", id).
		First(&detail)
	return detail, result.Error
}

func (s *PatrolServiceImpl) GetMemberByUserID(ctx context.Context, userID int64) (map[string]interface{}, error) {
	var detail map[string]interface{}
	result := database.GetDB().Table("grid_member gm").
		Select("gm.*, u.username, u.phone").
		Joins("LEFT JOIN user u ON u.id = gm.user_id").
		Where("gm.user_id = ? AND gm.status = 1", userID).
		First(&detail)
	return detail, result.Error
}

func (s *PatrolServiceImpl) CreateMember(ctx context.Context, req map[string]interface{}) (int64, error) {
	memberNo := fmt.Sprintf("GM%s", utils.GenerateID())
	member := model.GridMember{
		UserID:         int64(req["userId"].(float64)),
		MemberNo:       memberNo,
		RealName:       req["realName"].(string),
		Phone:          stringOrDefault(req, "phone", ""),
		OrganizationID: int64OrDefault(req, "orgId", 0),
		GridCodes:      stringOrDefault(req, "gridCodes", ""),
		Status:         int32(intOrDefault(req, "status", 1)),
	}
	result := database.GetDB().Create(&member)
	return member.ID, result.Error
}

func (s *PatrolServiceImpl) UpdateMember(ctx context.Context, id int64, req map[string]interface{}) error {
	updates := make(map[string]interface{})
	for k, v := range req {
		updates[utils.CamelToSnake(k)] = v
	}
	return database.GetDB().Table("grid_member").Where("id = ?", id).Updates(updates).Error
}

func (s *PatrolServiceImpl) DeleteMember(ctx context.Context, id int64) error {
	return database.GetDB().Table("grid_member").Where("id = ?", id).Update("status", 0).Error
}

func (s *PatrolServiceImpl) calculateTaskPoints(taskType int) int {
	switch taskType {
	case 1:
		return 50
	case 2:
		return 100
	case 3:
		return 200
	case 4:
		return 300
	default:
		return 50
	}
}

func (s *PatrolServiceImpl) calculateDangerPoints(level int) int {
	switch level {
	case 1:
		return 20
	case 2:
		return 50
	case 3:
		return 100
	case 4:
		return 200
	default:
		return 20
	}
}

func getTaskTypeName(taskType int) string {
	switch taskType {
	case 1:
		return "日常排查"
	case 2:
		return "专项排查"
	case 3:
		return "重点排查"
	case 4:
		return "紧急排查"
	default:
		return "其他"
	}
}

func getPriorityName(priority int) string {
	switch priority {
	case 1:
		return "紧急"
	case 2:
		return "高"
	case 3:
		return "中"
	case 4:
		return "低"
	default:
		return "中"
	}
}

func getStrategyName(strategy int) string {
	switch strategy {
	case 0:
		return "速度优先"
	case 1:
		return "费用优先"
	case 2:
		return "距离优先"
	case 3:
		return "躲避拥堵"
	case 4:
		return "不走高速"
	case 5:
		return "高速优先"
	case 6:
		return "躲避拥堵+不走高速"
	case 7:
		return "躲避拥堵+高速优先"
	case 8:
		return "躲避收费"
	case 9:
		return "躲避隧道"
	case 10:
		return "速度优先(最近邻)"
	case 11:
		return "距离最短(贪心插入)"
	case 12:
		return "优先级优先"
	case 13:
		return "综合最优(加权)"
	default:
		return "速度优先(最近邻)"
	}
}

func getVisitTypeName(visitType int) string {
	switch visitType {
	case 1:
		return "日常走访"
	case 2:
		return "重点人员走访"
	case 3:
		return "纠纷回访"
	case 4:
		return "特殊人群走访"
	case 5:
		return "隐患排查"
	default:
		return "其他"
	}
}

func getVisitStatusName(status int) string {
	switch status {
	case 10:
		return "待审核"
	case 20:
		return "审核通过"
	case 30:
		return "审核拒绝"
	default:
		return "未知"
	}
}

func getDangerTypeName(dangerType int) string {
	switch dangerType {
	case 1:
		return "消防安全"
	case 2:
		return "治安隐患"
	case 3:
		return "矛盾纠纷"
	case 4:
		return "安全生产"
	case 5:
		return "环境卫生"
	case 6:
		return "其他隐患"
	default:
		return "其他"
	}
}

func getDangerLevelName(level int) string {
	switch level {
	case 1:
		return "一般"
	case 2:
		return "较大"
	case 3:
		return "重大"
	case 4:
		return "特别重大"
	default:
		return "一般"
	}
}

func getDangerStatusName(status int) string {
	switch status {
	case 10:
		return "待处理"
	case 20:
		return "处理中"
	case 30:
		return "已处理"
	case 40:
		return "已关闭"
	default:
		return "未知"
	}
}

func parseTime(v interface{}) *time.Time {
	if v == nil {
		return nil
	}
	if str, ok := v.(string); ok && str != "" {
		t, err := time.ParseInLocation("2006-01-02 15:04:05", str, time.Local)
		if err == nil {
			return &t
		}
		t, err = time.ParseInLocation("2006-01-02", str, time.Local)
		if err == nil {
			return &t
		}
	}
	return nil
}

func floatOrDefault(m map[string]interface{}, key string, def float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	return def
}

func intOrDefault(m map[string]interface{}, key string, def int) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	if v, ok := m[key].(int); ok {
		return v
	}
	return def
}

func int64OrDefault(m map[string]interface{}, key string, def int64) int64 {
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	return def
}

func stringOrDefault(m map[string]interface{}, key string, def string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return def
}

func structToMap(s interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	switch v := s.(type) {
	case model.PatrolTask:
		result["id"] = v.ID
		result["taskNo"] = v.TaskNo
		result["title"] = v.Title
		result["description"] = v.Description
		result["taskType"] = v.TaskType
		result["typeName"] = v.TypeName
		result["priority"] = v.Priority
		result["priorityName"] = v.PriorityName
		result["status"] = v.Status
		result["statusName"] = v.StatusName
		result["pointsReward"] = v.PointsReward
		result["assigneeId"] = v.AssigneeID
		result["startTime"] = v.StartTime
		result["endTime"] = v.EndTime
		result["createdAt"] = v.CreatedAt
	}
	return result
}

func sendTaskNotification(phone, title string) {
	logger.Info("Sending task notification", logger.String("phone", phone), logger.String("title", title))
}

var logger = struct {
	Info  func(msg string, args ...interface{})
	Error func(msg string, args ...interface{})
}{
	Info: func(msg string, args ...interface{}) {
		fmt.Printf("[INFO] "+msg+"\n", args...)
	},
	Error: func(msg string, args ...interface{}) {
		fmt.Printf("[ERROR] "+msg+"\n", args...)
	},
}
