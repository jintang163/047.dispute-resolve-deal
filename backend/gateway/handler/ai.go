package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"
	"github.com/dispute-resolve/gateway/service"

	"github.com/cloudwego/hertz/pkg/app"
)

type AIConsultRequest struct {
	Question string `json:"question" binding:"required"`
	UserType int32  `json:"userType"`
}

type AIResponse struct {
	Summary    string   `json:"summary"`
	Suggestion string   `json:"suggestion"`
	RiskLevel  int32    `json:"riskLevel"`
	RiskPoints string   `json:"riskPoints"`
	LawRefs    []LawRef `json:"lawRefs"`
	TokensUsed int      `json:"tokensUsed"`
	CostTime   int      `json:"costTime"`
}

type LawRef struct {
	ID            int64  `json:"id"`
	LawName       string `json:"lawName"`
	ArticleNo     string `json:"articleNo"`
	ArticleTitle  string `json:"articleTitle"`
	ArticleContent string `json:"articleContent"`
	Relevance     float64 `json:"relevance"`
}

func AIConsult(ctx context.Context, c *app.RequestContext) {
	var req AIConsultRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	var userID int64
	var userName string
	var userType int32 = 1

	if userInfo != nil {
		userID = userInfo.UserID
		userName = userInfo.RealName
		userType = req.UserType
	}

	startTime := time.Now()

	questionType := service.ClassifyQuestion(req.Question)

	lawRefs, err := service.SearchRelevantLaw(req.Question)
	if err != nil {
		logger.Error("Search relevant law failed", logger.Error(err))
	}

	aiResp, err := service.GenerateLegalAdvice(req.Question, lawRefs)
	if err != nil {
		logger.Error("Generate legal advice failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("生成法律建议失败，请稍后重试"))
		return
	}

	costTime := int(time.Since(startTime).Milliseconds())

	consultNo := utils.GenerateConsultNo()

	lawRefIDs := make([]string, len(lawRefs))
	for i, ref := range lawRefs {
		lawRefIDs[i] = strconv.FormatInt(ref.ID, 10)
	}

	consult := map[string]interface{}{
		"id":                  utils.GenerateID(),
		"consult_no":          consultNo,
		"user_type":           userType,
		"user_id":             userID,
		"user_name":           userName,
		"question":            req.Question,
		"question_type":       questionType,
		"related_law_articles": strings.Join(lawRefIDs, ","),
		"ai_answer":           aiResp.Summary,
		"ai_model":            "deepseek",
		"reference_cases":     "",
		"tokens_used":         aiResp.TokensUsed,
		"cost_time":           costTime,
	}
	database.GetDB().Table("ai_law_consult").Create(consult)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"consultNo":    consultNo,
		"answer":       aiResp.Summary,
		"suggestion":   aiResp.Suggestion,
		"lawRefs":      lawRefs,
		"questionType": questionType,
		"tokensUsed":   aiResp.TokensUsed,
		"costTime":     costTime,
	}))
}

func GetLawArticles(ctx context.Context, c *app.RequestContext) {
	var req common.BaseQuery
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	db := database.GetDB().Table("ai_law_article").
		Where("status = 1")

	if req.Keyword != "" {
		db = db.Where("law_name LIKE ? OR article_title LIKE ? OR article_content LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("id DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func CreateLawArticle(ctx context.Context, c *app.RequestContext) {
	var req struct {
		LawCode        string `json:"lawCode"`
		LawName        string `json:"lawName" binding:"required"`
		ArticleNo      string `json:"articleNo"`
		ArticleTitle   string `json:"articleTitle"`
		ArticleContent string `json:"articleContent" binding:"required"`
		Category       string `json:"category"`
		Tags           string `json:"tags"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	id := utils.GenerateID()

	article := map[string]interface{}{
		"id":              id,
		"law_code":        req.LawCode,
		"law_name":        req.LawName,
		"article_no":      req.ArticleNo,
		"article_title":   req.ArticleTitle,
		"article_content": req.ArticleContent,
		"category":        req.Category,
		"tags":            req.Tags,
		"status":          1,
	}

	database.GetDB().Table("ai_law_article").Create(article)

	go func() {
		service.VectorizeLawArticle(id, req.ArticleContent)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id": id,
	}, "法条创建成功"))
}

func UpdateLawArticle(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		LawCode        string `json:"lawCode"`
		LawName        string `json:"lawName"`
		ArticleNo      string `json:"articleNo"`
		ArticleTitle   string `json:"articleTitle"`
		ArticleContent string `json:"articleContent"`
		Category       string `json:"category"`
		Tags           string `json:"tags"`
		Status         int32  `json:"status"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	updates := make(map[string]interface{})
	if req.LawCode != "" {
		updates["law_code"] = req.LawCode
	}
	if req.LawName != "" {
		updates["law_name"] = req.LawName
	}
	if req.ArticleNo != "" {
		updates["article_no"] = req.ArticleNo
	}
	if req.ArticleTitle != "" {
		updates["article_title"] = req.ArticleTitle
	}
	if req.ArticleContent != "" {
		updates["article_content"] = req.ArticleContent
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.Status > 0 {
		updates["status"] = req.Status
	}

	database.GetDB().Table("ai_law_article").
		Where("id = ?", id).
		Updates(updates)

	if req.ArticleContent != "" {
		go func() {
			service.VectorizeLawArticle(id, req.ArticleContent)
		}()
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "法条更新成功"))
}

func DeleteLawArticle(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	database.GetDB().Table("ai_law_article").
		Where("id = ?", id).
		Update("status", 0)

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "法条删除成功"))
}

func VectorizeLawArticles(ctx context.Context, c *app.RequestContext) {
	go func() {
		var articles []struct {
			ID              int64  `gorm:"column:id"`
			ArticleContent  string `gorm:"column:article_content"`
		}

		database.GetDB().Table("ai_law_article").
			Select("id, article_content").
			Where("status = 1").
			Find(&articles)

		for _, article := range articles {
			service.VectorizeLawArticle(article.ID, article.ArticleContent)
			time.Sleep(time.Millisecond * 100)
		}
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "已开始向量化处理"))
}

func GetAIConfig(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"model":       "deepseek-chat",
		"maxTokens":   4096,
		"temperature": 0.7,
		"enabled":     true,
	}))
}

func UpdateAIConfig(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Model       string  `json:"model"`
		MaxTokens   int     `json:"maxTokens"`
		Temperature float64 `json:"temperature"`
		Enabled     bool    `json:"enabled"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "AI配置更新成功"))
}
