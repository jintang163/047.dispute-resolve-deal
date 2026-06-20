package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"
	"github.com/dispute-resolve/gateway/service"

	"github.com/cloudwego/hertz/pkg/app"
)

type EscalationListRequest struct {
	model.BaseQuery
	ToLevel int   `form:"toLevel" json:"toLevel"`
	Status  int32 `form:"status" json:"status"`
}

type HandleEscalationRequest struct {
	Remark string `json:"remark"`
}

type CloseEscalationRequest struct {
	Remark string `json:"remark" binding:"required"`
}

func GetEscalationListHandler(ctx context.Context, c *app.RequestContext) {
	var req EscalationListRequest
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid request parameters"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	orgID := userInfo.OrganizationID
	if userInfo.Role == constants.RoleAdmin {
		orgID = 0
	}

	svc := service.TimeoutUrgeServiceInst()
	list, total, err := svc.GetEscalationList(ctx, orgID, req.ToLevel, req.Status, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"list":  list,
		"total": total,
		"page":  req.Page,
		"size":  req.PageSize,
	}))
}

func GetEscalationDetailHandler(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid escalation id"))
		return
	}

	svc := service.TimeoutUrgeServiceInst()
	detail, err := svc.GetEscalationDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(detail))
}

func GetCaseEscalationListHandler(ctx context.Context, c *app.RequestContext) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseInt(caseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid case id"))
		return
	}

	svc := service.TimeoutUrgeServiceInst()
	list, err := svc.GetCaseEscalationList(ctx, caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(list))
}

func GetCaseUrgeListHandler(ctx context.Context, c *app.RequestContext) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseInt(caseIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid case id"))
		return
	}

	svc := service.TimeoutUrgeServiceInst()
	list, err := svc.GetCaseUrgeList(ctx, caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(list))
}

func HandleEscalationHandler(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid escalation id"))
		return
	}

	var req HandleEscalationRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	svc := service.TimeoutUrgeServiceInst()
	err = svc.HandleEscalation(ctx, id, userInfo.UserID, userInfo.RealName, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "处理成功"))
}

func CloseEscalationHandler(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("invalid escalation id"))
		return
	}

	var req CloseEscalationRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限关闭升级记录"))
		return
	}

	svc := service.TimeoutUrgeServiceInst()
	err = svc.CloseEscalation(ctx, id, userInfo.UserID, userInfo.RealName, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "关闭成功"))
}
