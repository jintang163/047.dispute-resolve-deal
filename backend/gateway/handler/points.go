package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"
	"github.com/dispute-resolve/gateway/service/impl"
)

var pointsService = impl.NewPointsService()

func GetPointsSummary(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	summary, err := pointsService.GetPointsSummary(ctx, userInfo.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(summary))
}

func AddPoints(ctx context.Context, c *app.RequestContext) {
	var req struct {
		MemberID     int64  `json:"memberId" binding:"required"`
		Points       int    `json:"points" binding:"required"`
		BusinessType string `json:"businessType"`
		BusinessNo   string `json:"businessNo"`
		Description  string `json:"description"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限发放积分"))
		return
	}

	err := pointsService.AddPoints(ctx, req.MemberID, req.Points, req.BusinessType, req.BusinessNo, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeductPoints(ctx context.Context, c *app.RequestContext) {
	var req struct {
		MemberID     int64  `json:"memberId" binding:"required"`
		Points       int    `json:"points" binding:"required"`
		BusinessType string `json:"businessType"`
		BusinessNo   string `json:"businessNo"`
		Description  string `json:"description"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限扣减积分"))
		return
	}

	err := pointsService.DeductPoints(ctx, req.MemberID, req.Points, req.BusinessType, req.BusinessNo, req.Description)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetPointsRecords(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	page := 1
	if p := c.Query("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}
	pageSize := 20
	if ps := c.Query("pageSize"); ps != "" {
		pageSize, _ = strconv.Atoi(ps)
	}

	records, total, err := pointsService.GetPointsRecords(ctx, userInfo.UserID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Page(records, total, page, pageSize))
}

func GetPointsRules(ctx context.Context, c *app.RequestContext) {
	ruleType := c.Query("ruleType")

	rules, err := pointsService.GetPointsRules(ctx, ruleType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(rules))
}

func CreatePointsRule(ctx context.Context, c *app.RequestContext) {
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleAdmin {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限创建积分规则"))
		return
	}

	err := pointsService.CreatePointsRule(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func UpdatePointsRule(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的规则ID"))
		return
	}

	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleAdmin {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限编辑积分规则"))
		return
	}

	err = pointsService.UpdatePointsRule(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeletePointsRule(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的规则ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleAdmin {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限删除积分规则"))
		return
	}

	err = pointsService.DeletePointsRule(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func ExchangeGift(ctx context.Context, c *app.RequestContext) {
	var req struct {
		GiftID          int    `json:"giftId" binding:"required"`
		Quantity        int    `json:"quantity" binding:"required"`
		ReceiverName    string `json:"receiverName"`
		ReceiverPhone   string `json:"receiverPhone"`
		ReceiverAddress string `json:"receiverAddress"`
		Remark          string `json:"remark"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	id, err := pointsService.ExchangeGift(ctx, userInfo.UserID, int64(req.GiftID), req.Quantity,
		req.ReceiverName, req.ReceiverPhone, req.ReceiverAddress, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(id))
}

func ProcessExpiredPoints(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role > constants.RoleAdmin {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限执行此操作"))
		return
	}

	err := pointsService.ProcessExpiredPoints(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
