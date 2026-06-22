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

var giftService = impl.NewGiftService()

func GetGiftList(ctx context.Context, c *app.RequestContext) {
	query := make(map[string]interface{})

	if categoryId := c.Query("categoryId"); categoryId != "" {
		if id, err := strconv.ParseInt(categoryId, 10, 64); err == nil {
			query["categoryId"] = float64(id)
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
	if isHot := c.Query("isHot"); isHot != "" {
		if h, err := strconv.Atoi(isHot); err == nil {
			query["isHot"] = float64(h)
		}
	}
	if isNew := c.Query("isNew"); isNew != "" {
		if n, err := strconv.Atoi(isNew); err == nil {
			query["isNew"] = float64(n)
		}
	}
	if minPoints := c.Query("minPoints"); minPoints != "" {
		if mp, err := strconv.Atoi(minPoints); err == nil {
			query["minPoints"] = float64(mp)
		}
	}
	if maxPoints := c.Query("maxPoints"); maxPoints != "" {
		if mp, err := strconv.Atoi(maxPoints); err == nil {
			query["maxPoints"] = float64(mp)
		}
	}
	if sortBy := c.Query("sortBy"); sortBy != "" {
		query["sortBy"] = sortBy
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

	list, total, err := giftService.GetGiftList(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetGiftDetail(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的礼品ID"))
		return
	}

	detail, err := giftService.GetGiftDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(detail))
}

func CreateGift(ctx context.Context, c *app.RequestContext) {
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限添加礼品"))
		return
	}

	id, err := giftService.CreateGift(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(id))
}

func UpdateGift(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的礼品ID"))
		return
	}

	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限编辑礼品"))
		return
	}

	err = giftService.UpdateGift(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeleteGift(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的礼品ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限删除礼品"))
		return
	}

	err = giftService.DeleteGift(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetGiftCategories(ctx context.Context, c *app.RequestContext) {
	categories, err := giftService.GetGiftCategories(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(categories))
}

func CreateGiftCategory(ctx context.Context, c *app.RequestContext) {
	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限添加分类"))
		return
	}

	id, err := giftService.CreateGiftCategory(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(id))
}

func UpdateGiftCategory(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的分类ID"))
		return
	}

	var req map[string]interface{}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限编辑分类"))
		return
	}

	err = giftService.UpdateGiftCategory(ctx, id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func DeleteGiftCategory(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的分类ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限删除分类"))
		return
	}

	err = giftService.DeleteGiftCategory(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetExchangeList(ctx context.Context, c *app.RequestContext) {
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
	if giftId := c.Query("giftId"); giftId != "" {
		if id, err := strconv.ParseInt(giftId, 10, 64); err == nil {
			query["giftId"] = float64(id)
		}
	}
	if startDate := c.Query("startDate"); startDate != "" {
		query["startDate"] = startDate
	}
	if endDate := c.Query("endDate"); endDate != "" {
		query["endDate"] = endDate
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

	list, total, err := giftService.GetExchangeList(ctx, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetExchangeDetail(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的兑换ID"))
		return
	}

	detail, err := giftService.GetExchangeDetail(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(detail))
}

func AuditExchange(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的兑换ID"))
		return
	}

	var req struct {
		Status int32  `json:"status" binding:"required"`
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

	err = giftService.AuditExchange(ctx, id, req.Status, req.Remark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func ShipExchange(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的兑换ID"))
		return
	}

	var req struct {
		ExpressCompany string `json:"expressCompany" binding:"required"`
		ExpressNo      string `json:"expressNo" binding:"required"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo.Role != constants.RoleAdmin && userInfo.Role != constants.RoleDirector && userInfo.Role != constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限发货"))
		return
	}

	err = giftService.ShipExchange(ctx, id, req.ExpressCompany, req.ExpressNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func ReceiveExchange(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的兑换ID"))
		return
	}

	err = giftService.ReceiveExchange(ctx, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func CancelExchange(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的兑换ID"))
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	err = giftService.CancelExchange(ctx, id, req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func GetMemberExchanges(ctx context.Context, c *app.RequestContext) {
	userInfo := middleware.GetUserInfo(c)

	page := 1
	if p := c.Query("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}
	pageSize := 20
	if ps := c.Query("pageSize"); ps != "" {
		pageSize, _ = strconv.Atoi(ps)
	}

	list, total, err := giftService.GetMemberExchanges(ctx, userInfo.UserID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Page(list, total, page, pageSize))
}

func GetGiftStatistics(ctx context.Context, c *app.RequestContext) {
	stats, err := giftService.GetGiftStatistics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(stats))
}
