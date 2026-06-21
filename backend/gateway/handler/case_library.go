package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/gateway/middleware"
	"github.com/dispute-resolve/gateway/service"

	"github.com/cloudwego/hertz/pkg/app"
	"go.uber.org/zap"
)

func GetCaseLibraryList(ctx context.Context, c *app.RequestContext) {
	var req model.BaseQuery
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	disputeType := c.Query("disputeType")
	difficultyStr := c.Query("difficultyLevel")
	statusStr := c.DefaultQuery("status", "-1")

	difficultyLevel, _ := strconv.Atoi(difficultyStr)
	status, _ := strconv.Atoi(statusStr)

	svc := service.CaseLibraryServiceInst()
	cases, total, err := svc.ListCases(ctx, req.Page, req.PageSize, req.Keyword, disputeType, difficultyLevel, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Page(cases, total, req.Page, req.PageSize))
}

func GetCaseLibraryDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.CaseLibraryServiceInst()
	caseLib, err := svc.GetCase(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, response.NotFound(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(caseLib))
}

func CreateCaseLibrary(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Title            string `json:"title" binding:"required"`
		Description      string `json:"description"`
		DisputeType      string `json:"disputeType"`
		TypeID           int64  `json:"typeId"`
		MediationTactics string `json:"mediationTactics"`
		KeyPoints        string `json:"keyPoints"`
		ResultSummary    string `json:"resultSummary"`
		DifficultyLevel  int    `json:"difficultyLevel"`
		IsSuccess        int32  `json:"isSuccess"`
		MediatorName     string `json:"mediatorName"`
		MediatorID       int64  `json:"mediatorId"`
		OrgName          string `json:"orgName"`
		OrgID            int64  `json:"orgId"`
		SourceCaseID     int64  `json:"sourceCaseId"`
		Keywords         string `json:"keywords"`
		Tags             string `json:"tags"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	var createdBy int64
	if userInfo != nil {
		createdBy = userInfo.UserID
	}

	if req.DifficultyLevel <= 0 {
		req.DifficultyLevel = constants.CaseLibraryDifficultyNormal
	}
	if req.IsSuccess <= 0 {
		req.IsSuccess = 1
	}

	caseLib := &model.CaseLibrary{
		Title:            req.Title,
		Description:      req.Description,
		DisputeType:      req.DisputeType,
		TypeID:           req.TypeID,
		MediationTactics: req.MediationTactics,
		KeyPoints:        req.KeyPoints,
		ResultSummary:    req.ResultSummary,
		DifficultyLevel:  req.DifficultyLevel,
		IsSuccess:        req.IsSuccess,
		MediatorName:     req.MediatorName,
		MediatorID:       req.MediatorID,
		OrgName:          req.OrgName,
		OrgID:            req.OrgID,
		SourceCaseID:     req.SourceCaseID,
		Keywords:         req.Keywords,
		Tags:             req.Tags,
		Status:           constants.CaseLibraryStatusActive,
		CreatedBy:        createdBy,
	}

	svc := service.CaseLibraryServiceInst()
	if err := svc.CreateCase(ctx, caseLib); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id": caseLib.ID,
	}, "案例创建成功"))
}

func UpdateCaseLibrary(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Title            string `json:"title"`
		Description      string `json:"description"`
		DisputeType      string `json:"disputeType"`
		TypeID           int64  `json:"typeId"`
		MediationTactics string `json:"mediationTactics"`
		KeyPoints        string `json:"keyPoints"`
		ResultSummary    string `json:"resultSummary"`
		DifficultyLevel  int    `json:"difficultyLevel"`
		IsSuccess        int32  `json:"isSuccess"`
		MediatorName     string `json:"mediatorName"`
		MediatorID       int64  `json:"mediatorId"`
		OrgName          string `json:"orgName"`
		OrgID            int64  `json:"orgId"`
		Keywords         string `json:"keywords"`
		Tags             string `json:"tags"`
		Status           int32  `json:"status"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	caseLib := &model.CaseLibrary{
		BaseModel:        model.BaseModel{ID: id},
		Title:            req.Title,
		Description:      req.Description,
		DisputeType:      req.DisputeType,
		TypeID:           req.TypeID,
		MediationTactics: req.MediationTactics,
		KeyPoints:        req.KeyPoints,
		ResultSummary:    req.ResultSummary,
		DifficultyLevel:  req.DifficultyLevel,
		IsSuccess:        req.IsSuccess,
		MediatorName:     req.MediatorName,
		MediatorID:       req.MediatorID,
		OrgName:          req.OrgName,
		OrgID:            req.OrgID,
		Keywords:         req.Keywords,
		Tags:             req.Tags,
		Status:           req.Status,
	}

	svc := service.CaseLibraryServiceInst()
	if err := svc.UpdateCase(ctx, caseLib); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "案例更新成功"))
}

func DeleteCaseLibrary(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.CaseLibraryServiceInst()
	if err := svc.DeleteCase(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "案例删除成功"))
}

func SearchSimilarCases(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Query  string `json:"query"`
		CaseID int64  `json:"caseId"`
		TopK   int    `json:"topK"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	if req.Query == "" && req.CaseID <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("query或caseId参数不能同时为空"))
		return
	}

	if req.TopK <= 0 {
		req.TopK = 5
	}

	svc := service.CaseLibraryServiceInst()
	results, err := svc.SearchSimilarCases(ctx, req.Query, req.CaseID, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(results))
}

func ScoreCaseLibrary(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Score        int    `json:"score" binding:"required"`
		SourceCaseID int64  `json:"sourceCaseId"`
		Comment      string `json:"comment"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	if req.Score < 1 || req.Score > 5 {
		c.JSON(http.StatusBadRequest, response.BadRequest("评分必须在1-5之间"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	var userID int64
	var userName string
	if userInfo != nil {
		userID = userInfo.UserID
		userName = userInfo.RealName
	}

	score := &model.CaseLibraryScore{
		CaseID:       id,
		UserID:       userID,
		UserName:     userName,
		Score:        req.Score,
		SourceCaseID: req.SourceCaseID,
		Comment:      req.Comment,
	}

	svc := service.CaseLibraryServiceInst()
	if err := svc.ScoreCase(ctx, score); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "评分成功"))
}

func QuoteCaseLibrary(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		SourceCaseID      int64  `json:"sourceCaseId" binding:"required"`
		QuoteType         int32  `json:"quoteType"`
		QuoteContent      string `json:"quoteContent"`
		MediationRecordID int64  `json:"mediationRecordId"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	if req.QuoteType <= 0 {
		req.QuoteType = constants.CaseLibraryQuoteTypeTactics
	}

	userInfo := middleware.GetUserInfo(c)
	var userID int64
	var userName string
	if userInfo != nil {
		userID = userInfo.UserID
		userName = userInfo.RealName
	}

	quote := &model.CaseLibraryQuote{
		SourceCaseID:      req.SourceCaseID,
		LibraryCaseID:     id,
		QuoteType:         req.QuoteType,
		QuoteContent:      req.QuoteContent,
		UserID:            userID,
		UserName:          userName,
		MediationRecordID: req.MediationRecordID,
	}

	svc := service.CaseLibraryServiceInst()
	if err := svc.QuoteCase(ctx, quote); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"quoteContent":      quote.QuoteContent,
		"mediationRecordId": quote.MediationRecordID,
	}, "引用成功"))
}

func GetCaseQuoteList(ctx context.Context, c *app.RequestContext) {
	sourceCaseIDStr := c.Query("sourceCaseId")
	sourceCaseID, _ := strconv.ParseInt(sourceCaseIDStr, 10, 64)

	if sourceCaseID <= 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("sourceCaseId参数错误"))
		return
	}

	svc := service.CaseLibraryServiceInst()
	quotes, err := svc.GetQuoteList(ctx, sourceCaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(quotes))
}

func ArchiveCaseLibrary(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Reason int32 `json:"reason"`
	}
	c.BindAndValidate(&req)
	if req.Reason <= 0 {
		req.Reason = constants.CaseLibraryArchiveReasonManual
	}

	userInfo := middleware.GetUserInfo(c)
	var archivedBy int64
	if userInfo != nil {
		archivedBy = userInfo.UserID
	}

	svc := service.CaseLibraryServiceInst()
	if err := svc.ArchiveCase(ctx, id, archivedBy, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "归档成功"))
}

func RestoreCaseLibrary(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.CaseLibraryServiceInst()
	if err := svc.RestoreFromArchive(ctx, id); err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "恢复成功"))
}

func VectorizeCaseLibrary(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	svc := service.CaseLibraryServiceInst()
	go func() {
		if err := svc.VectorizeCase(context.Background(), id); err != nil {
			logger.Error("Vectorize case failed", logger.Error(err))
		}
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "已开始向量化处理"))
}

func VectorizeAllCaseLibrary(ctx context.Context, c *app.RequestContext) {
	svc := service.CaseLibraryServiceInst()
	go func() {
		count, err := svc.VectorizeAllCases(context.Background())
		if err != nil {
			logger.Error("Vectorize all cases failed", logger.Error(err))
			return
		}
		logger.Info("Vectorize all cases completed", zap.Int("processedCount", count))
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "已开始批量向量化处理"))
}

func GetCaseLibraryArchiveList(ctx context.Context, c *app.RequestContext) {
	var req model.BaseQuery
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	svc := service.CaseLibraryServiceInst()
	archives, total, err := svc.GetArchiveList(ctx, req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Page(archives, total, req.Page, req.PageSize))
}
