package handler

import (
	"context"
	"strconv"

	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"
	"github.com/dispute-resolve/gateway/service"

	"github.com/cloudwego/hertz/pkg/app"
)

type ImprovementListRequest struct {
	model.BaseQuery
	MediatorID int64 `form:"mediatorId" json:"mediatorId"`
	Status     int   `form:"status" json:"status"`
}

type RectificationRequest struct {
	Content string `json:"content" binding:"required"`
	Result  string `json:"result" binding:"required"`
}

type ReviewRequest struct {
	Opinion  string `json:"opinion" binding:"required"`
	Approved bool   `json:"approved"`
}

type SatisfactionStatsRequest struct {
	OrgID     int64  `form:"orgId" json:"orgId"`
	StartDate string `form:"startDate" json:"startDate"`
	EndDate   string `form:"endDate" json:"endDate"`
}

func AnalyzeSatisfactionHandler(ctx context.Context, c *app.RequestContext) {
	caseIDStr := c.Param("caseId")
	caseID, err := strconv.ParseInt(caseIDStr, 10, 64)
	if err != nil {
		response.Fail(c, "invalid case id")
		return
	}

	svc := service.GetSatisfactionService()
	order, err := svc.AnalyzeSatisfaction(ctx, caseID)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	if order != nil {
		response.Success(c, map[string]interface{}{
			"sentimentAnalyzed": true,
			"emotion":           "negative",
			"improvementOrder":  order,
		})
	} else {
		response.Success(c, map[string]interface{}{
			"sentimentAnalyzed": true,
			"emotion":           "positive_or_neutral",
			"improvementOrder":  nil,
		})
	}
}

func GetImprovementOrderListHandler(ctx context.Context, c *app.RequestContext) {
	var req ImprovementListRequest
	if err := c.Bind(&req); err != nil {
		response.Fail(c, "invalid request parameters")
		return
	}

	svc := service.GetSatisfactionService()
	orders, total, err := svc.GetImprovementOrderList(ctx, req.MediatorID, req.Status, req.Page, req.PageSize)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"list":     orders,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	})
}

func GetImprovementOrderDetailHandler(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "invalid id")
		return
	}

	svc := service.GetSatisfactionService()
	order, err := svc.GetImprovementOrderDetail(ctx, id)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, order)
}

func SubmitRectificationHandler(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "invalid id")
		return
	}

	var req RectificationRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, "invalid request parameters")
		return
	}

	svc := service.GetSatisfactionService()
	if err := svc.SubmitRectification(ctx, id, req.Content, req.Result); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func ReviewRectificationHandler(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "invalid id")
		return
	}

	var req ReviewRequest
	if err := c.BindJSON(&req); err != nil {
		response.Fail(c, "invalid request parameters")
		return
	}

	userInfo := middleware.GetUserInfo(c)

	svc := service.GetSatisfactionService()
	if err := svc.ReviewRectification(ctx, id, userInfo.UserID, userInfo.RealName, req.Opinion, req.Approved); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func CloseImprovementOrderHandler(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "invalid id")
		return
	}

	remark := c.Query("remark")

	svc := service.GetSatisfactionService()
	if err := svc.CloseImprovementOrder(ctx, id, remark); err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func GetSatisfactionSentimentStatsHandler(ctx context.Context, c *app.RequestContext) {
	var req SatisfactionStatsRequest
	if err := c.Bind(&req); err != nil {
		response.Fail(c, "invalid request parameters")
		return
	}

	svc := service.GetSatisfactionService()
	stats, err := svc.GetSatisfactionSentimentStats(ctx, req.OrgID, req.StartDate, req.EndDate)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}

	response.Success(c, stats)
}
