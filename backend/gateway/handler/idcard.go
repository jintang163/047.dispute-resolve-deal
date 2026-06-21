package handler

import (
	"context"
	"net/http"

	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/population"
	"github.com/dispute-resolve/common/response"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

type IDCardQueryRequest struct {
	IDCard string `json:"idCard" binding:"required"`
}

func QueryPopulationByIDCard(ctx context.Context, c *app.RequestContext) {
	var req IDCardQueryRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("身份证号不能为空"))
		return
	}

	client := population.GetPopulationClient()
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, response.ServerError("人口库服务未配置"))
		return
	}

	info, err := client.QueryByIDCard(req.IDCard)
	if err != nil {
		logger.Error("Population query population info failed",
			zap.String("idcard", maskIDCard(req.IDCard)),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("人口信息查询失败："+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(info))
}

func maskIDCard(idCard string) string {
	if len(idCard) < 8 {
		return "****"
	}
	return idCard[:6] + "********" + idCard[14:]
}
