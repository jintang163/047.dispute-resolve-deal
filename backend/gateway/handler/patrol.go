package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"
	"github.com/dispute-resolve/gateway/service/impl"
	"gorm.io/gorm"
)

var patrolService = impl.NewPatrolService()

func CreatePatrolTask(ctx context.Context, c *app.RequestContext) {
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限下发任务"))
		return
	}

	id, err := patrolService.CreateTask(ctx, req, userInfo.UserID, userInfo.RealName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(id))
}

func GetPatrolTaskList(ctx context.Context, c *app.RequestContext) {
	query := make(map[string]interface{})

	if status := c.Query("status"); status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query["status"] = float64(s)
		}
	}
	if assigneeId := c.Query("assigneeId"); assigneeId != "" {
		if id, err := strconv.ParseInt(assigneeId, 10, 64); err == nil {
			query["assigneeId"] = float64(id)
		}
	}
	if taskType := c.Query("taskType"); taskType != "" {
		if t, err := strconv.Atoi(taskType); err == nil {
			query["taskType"] = float64(t)
		}
	}
	if priority := c.Query("priority"); priority != "" {
		if p, err := strconv.Atoi(priority); err == nil {
			query["priority"] = float64(p)
		}
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query["keyword"] = keyword
	}
	if orgId := c.Query("orgId"); orgId != "" {
		if id, err := strconv.ParseInt(orgId, 10, 64); err == nil {
			query["orgId"] = float64(id)
		}
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			query["page"] = p
		}
	}
	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			query["pageSize"] = ps
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	list, total, err := patrolService.GetTaskList(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetPatrolTaskDetail(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的任务ID"))
		return
	}

	detail, err := patrolService.GetTaskDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(detail))
}

func UpdatePatrolTask(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的任务ID"))
		return
	}

	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限编辑任务"))
		return
	}

	err = patrolService.UpdateTask(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeletePatrolTask(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的任务ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限删除任务"))
		return
	}

	err = patrolService.DeleteTask(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func CancelPatrolTask(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的任务ID"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限取消任务"))
		return
	}

	err = patrolService.CancelTask(ctx, id, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func StartPatrolTask(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的任务ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	err = patrolService.StartTask(ctx, id, userInfo.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func CompletePatrolTask(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的任务ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	err = patrolService.CompleteTask(ctx, id, userInfo.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func PlanRoute(ctx context.Context, c *app.RequestContext) {
	var req struct {
		StartLng float64                   `json:"startLng"`
		StartLat float64                   `json:"startLat"`
		Points   []map[string]interface{} `json:"points"`
		Strategy int                      `json:"strategy"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	if req.Strategy == 0 {
		req.Strategy = 10
	}

	result, err := patrolService.PlanRoute(ctx, req.StartLng, req.StartLat, req.Points, req.Strategy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func GetMemberTasks(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	status := 0
	if s := c.Query("status"); s != "" {
		status, _ = strconv.Atoi(s)
	}
	page := 1
	if p := c.Query("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}
	pageSize := 20
	if ps := c.Query("pageSize"); ps != "" {
		pageSize, _ = strconv.Atoi(ps)
	}

	list, total, err := patrolService.GetMemberTasks(ctx, userInfo.UserID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetTaskPoints(ctx context.Context, c *app.RequestContext) {
	taskID, err := strconv.ParseInt(c.Param("taskId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的任务ID"))
		return
	}

	points, err := patrolService.GetTaskPoints(ctx, taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(points))
}

func Checkin(ctx context.Context, c *app.RequestContext) {
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	ipAddress := c.ClientIP()

	result, err := patrolService.Checkin(ctx, req, userInfo.UserID, userInfo.RealName, ipAddress)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func GetCheckinRecords(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	page := 1
	if p := c.Query("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}
	pageSize := 20
	if ps := c.Query("pageSize"); ps != "" {
		pageSize, _ = strconv.Atoi(ps)
	}

	list, total, err := patrolService.GetCheckinRecords(ctx, userInfo.UserID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetCheckinStatistics(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	stats, err := patrolService.GetCheckinStatistics(ctx, userInfo.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(stats))
}

func CreateVisitRecord(ctx context.Context, c *app.RequestContext) {
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	id, err := patrolService.CreateVisitRecord(ctx, req, userInfo.UserID, userInfo.RealName, userInfo.OrgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(id))
}

func GetVisitRecords(ctx context.Context, c *app.RequestContext) {
	query := make(map[string]interface{})

	if memberId := c.Query("memberId"); memberId != "" {
		if id, err := strconv.ParseInt(memberId, 10, 64); err == nil {
			query["memberId"] = float64(id)
		}
	}
	if status := c.Query("status"); status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query["status"] = float64(s)
		}
	}
	if visitType := c.Query("visitType"); visitType != "" {
		if t, err := strconv.Atoi(visitType); err == nil {
			query["visitType"] = float64(t)
		}
	}
	if orgId := c.Query("orgId"); orgId != "" {
		if id, err := strconv.ParseInt(orgId, 10, 64); err == nil {
			query["orgId"] = float64(id)
		}
	}
	if startDate := c.Query("startDate"); startDate != "" {
		query["startDate"] = startDate
	}
	if endDate := c.Query("endDate"); endDate != "" {
		query["endDate"] = endDate
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			query["page"] = p
		}
	}
	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			query["pageSize"] = ps
		}
	}

	list, total, err := patrolService.GetVisitRecords(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetVisitRecordDetail(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的记录ID"))
		return
	}

	detail, err := patrolService.GetVisitRecordDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(detail))
}

func UpdateVisitRecord(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的记录ID"))
		return
	}

	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	err = patrolService.UpdateVisitRecord(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func AuditVisitRecord(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的记录ID"))
		return
	}

	var req struct {
		Status int32  `json:"status"`
		Remark string `json:"remark"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限审核"))
		return
	}

	err = patrolService.AuditVisitRecord(ctx, id, req.Status, req.Remark, userInfo.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeleteVisitRecord(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的记录ID"))
		return
	}

	err = patrolService.DeleteVisitRecord(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetVisitStatistics(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	stats, err := patrolService.GetVisitStatistics(ctx, userInfo.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(stats))
}

func ReportDanger(ctx context.Context, c *app.RequestContext) {
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	id, err := patrolService.ReportDanger(ctx, req, userInfo.UserID, userInfo.RealName, userInfo.OrgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(id))
}

func GetDangerList(ctx context.Context, c *app.RequestContext) {
	query := make(map[string]interface{})

	if reporterId := c.Query("reporterId"); reporterId != "" {
		if id, err := strconv.ParseInt(reporterId, 10, 64); err == nil {
			query["reporterId"] = float64(id)
		}
	}
	if status := c.Query("status"); status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query["status"] = float64(s)
		}
	}
	if dangerType := c.Query("dangerType"); dangerType != "" {
		if t, err := strconv.Atoi(dangerType); err == nil {
			query["dangerType"] = float64(t)
		}
	}
	if level := c.Query("level"); level != "" {
		if l, err := strconv.Atoi(level); err == nil {
			query["level"] = float64(l)
		}
	}
	if orgId := c.Query("orgId"); orgId != "" {
		if id, err := strconv.ParseInt(orgId, 10, 64); err == nil {
			query["orgId"] = float64(id)
		}
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query["keyword"] = keyword
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			query["page"] = p
		}
	}
	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			query["pageSize"] = ps
		}
	}

	list, total, err := patrolService.GetDangerList(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetDangerDetail(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的隐患ID"))
		return
	}

	detail, err := patrolService.GetDangerDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(detail))
}

func HandleDanger(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的隐患ID"))
		return
	}

	var req struct {
		Status int32  `json:"status"`
		Result string `json:"result"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	err = patrolService.HandleDanger(ctx, id, req.Status, userInfo.UserID, userInfo.RealName, req.Result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetDangerStatistics(ctx context.Context, c *app.RequestContext) {
	stats, err := patrolService.GetDangerStatistics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(stats))
}

func GetMemberList(ctx context.Context, c *app.RequestContext) {
	query := make(map[string]interface{})

	if orgId := c.Query("orgId"); orgId != "" {
		if id, err := strconv.ParseInt(orgId, 10, 64); err == nil {
			query["orgId"] = float64(id)
		}
	}
	if status := c.Query("status"); status != "" {
		if s, err := strconv.Atoi(status); err == nil {
			query["status"] = float64(s)
		}
	}
	if keyword := c.Query("keyword"); keyword != "" {
		query["keyword"] = keyword
	}
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			query["page"] = p
		}
	}
	if pageSize := c.Query("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			query["pageSize"] = ps
		}
	}

	list, total, err := patrolService.GetMemberList(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetMemberDetail(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的网格员ID"))
		return
	}

	detail, err := patrolService.GetMemberDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(detail))
}

func CreateMember(ctx context.Context, c *app.RequestContext) {
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限添加网格员"))
		return
	}

	id, err := patrolService.CreateMember(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(id))
}

func UpdateMember(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的网格员ID"))
		return
	}

	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限编辑网格员"))
		return
	}

	err = patrolService.UpdateMember(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeleteMember(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的网格员ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限删除网格员"))
		return
	}

	err = patrolService.DeleteMember(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetMemberMe(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	detail, err := patrolService.GetMemberByUserID(ctx, userInfo.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, response.Success(nil))
			return
		}
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(detail))
}
