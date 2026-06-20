package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/ai"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"
	"github.com/dispute-resolve/gateway/service"

	common "github.com/dispute-resolve/common/model"

	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
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

func ExtractKeywords(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Text        string `json:"text" binding:"required"`
		Title       string `json:"title"`
		CaseTypeID  int64  `json:"typeId"`
		MaxKeywords int    `json:"maxKeywords"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	if req.MaxKeywords <= 0 || req.MaxKeywords > 20 {
		req.MaxKeywords = 8
	}

	input := req.Title
	if input == "" {
		input = req.Text
	} else {
		input = req.Title + "\n" + req.Text
	}

	var typeName string
	if req.CaseTypeID > 0 {
		database.GetDB().Table("dispute_type").
			Select("type_name").
			Where("id = ?", req.CaseTypeID).
			Scan(&typeName)
	}

	var allTypes []map[string]interface{}
	database.GetDB().Table("dispute_type").
		Select("id, type_code, type_name, parent_id, level").
		Where("status = 1").
		Order("level ASC, sort_order ASC, id ASC").
		Find(&allTypes)

	typeTreeText := buildDisputeTypeTreeText(allTypes)

	client := ai.GetDeepSeekClient()
	prompt := buildKeywordExtractionPromptV2(input, typeName, typeTreeText, req.MaxKeywords)

	result, err := client.ChatCompletion([]ai.ChatMessage{}, prompt)
	if err != nil {
		logger.Error("DeepSeek keyword extraction failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("关键词提取失败"))
		return
	}

	result = strings.TrimSpace(result)
	result = strings.TrimPrefix(result, "```json")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimSpace(result)

	keywords := []string{}
	suggestedTypeID := int64(0)
	suggestedTypeName := ""
	suggestionReason := ""

	var parsedResult map[string]interface{}
	if err := json.Unmarshal([]byte(result), &parsedResult); err == nil {
		if kws, ok := parsedResult["keywords"].([]interface{}); ok {
			for _, kw := range kws {
				if s, ok := kw.(string); ok {
					s = strings.TrimSpace(s)
					if s != "" {
						keywords = append(keywords, s)
					}
				}
			}
		}
		if tid, ok := parsedResult["suggestedTypeId"]; ok {
			switch v := tid.(type) {
			case float64:
				suggestedTypeID = int64(v)
			case int64:
				suggestedTypeID = v
			case string:
				suggestedTypeID, _ = strconv.ParseInt(v, 10, 64)
			}
		}
		if tname, ok := parsedResult["suggestedTypeName"].(string); ok {
			suggestedTypeName = tname
		}
		if reason, ok := parsedResult["reason"].(string); ok {
			suggestionReason = reason
		}
	} else {
		logger.Warn("Parse structured keyword result failed, fallback to legacy list",
			logger.String("raw", result[:min(len(result), 200)]), logger.Error(err))
		keywords = extractKeywordsFallback(result)
	}

	if len(keywords) > req.MaxKeywords {
		keywords = keywords[:req.MaxKeywords]
	}

	if suggestedTypeID > 0 {
		var checkName string
		database.GetDB().Table("dispute_type").
			Select("type_name").
			Where("id = ?", suggestedTypeID).
			Scan(&checkName)
		if checkName == "" {
			suggestedTypeID = 0
		} else {
			suggestedTypeName = checkName
		}
	}

	go updateKeywordDict(keywords)

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"keywords":          keywords,
		"count":             len(keywords),
		"suggestedTypeId":   suggestedTypeID,
		"suggestedTypeName": suggestedTypeName,
		"reason":            suggestionReason,
	}))
}

func buildDisputeTypeTreeText(types []map[string]interface{}) string {
	typeMap := make(map[int64]map[string]interface{})
	for _, t := range types {
		id := toInt64(t["id"])
		typeMap[id] = t
		t["children"] = []map[string]interface{}{}
	}
	root := []map[string]interface{}{}
	for _, t := range types {
		pid := toInt64(t["parent_id"])
		if pid == 0 {
			root = append(root, t)
		} else if parent, ok := typeMap[pid]; ok {
			ch := parent["children"].([]map[string]interface{})
			parent["children"] = append(ch, t)
		}
	}
	var sb strings.Builder
	writeDisputeTypeTree(&sb, root, 0)
	return sb.String()
}

func writeDisputeTypeTree(sb *strings.Builder, list []map[string]interface{}, indent int) {
	for _, t := range list {
		id := toInt64(t["id"])
		name := t["type_name"].(string)
		sb.WriteString(strings.Repeat("  ", indent))
		sb.WriteString(fmt.Sprintf("- id=%d | %s\n", id, name))
		if children, ok := t["children"].([]map[string]interface{}); ok && len(children) > 0 {
			writeDisputeTypeTree(sb, children, indent+1)
		}
	}
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int32:
		return int64(val)
	case int:
		return int64(val)
	case float64:
		return int64(val)
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	default:
		return 0
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func buildKeywordExtractionPromptV2(text, typeNameHint, typeTree string, maxKeywords int) string {
	typeHint := ""
	if typeNameHint != "" {
		typeHint = fmt.Sprintf("\n用户已初选的纠纷大类为「%s」，请在此大类范围内优先推断。", typeNameHint)
	}

	return fmt.Sprintf(`你是矛盾纠纷分类与打标签专家。请从纠纷描述中：
1) 提取8个以内有区分度的核心关键词
2) 从下方的纠纷类型三级分类树中，推断最匹配的**三级子分类的 ID**（务必选择一个叶子节点）

要求：
- 关键词：具体，如"噪音扰民""欠薪3个月""漏水赔偿"，不要"纠纷""问题"等宽泛词
- 分类：严格从树形中选存在的叶子节点的id；不能编造id；若拿不准选最接近的
- 返回纯JSON，格式：{"keywords": [...], "suggestedTypeId": 数字id, "suggestedTypeName": "名称", "reason": "匹配理由（一句话）"}

纠纷类型三级分类树（id | 名称）：
%s%s

纠纷描述：
%s`, typeTree, typeHint, text)
}


func GetKeywordDict(ctx context.Context, c *app.RequestContext) {
	category := c.Query("category")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	db := database.GetDB().Table("dispute_keyword_dict").
		Where("status = 1")

	if category != "" {
		db = db.Where("category = ?", category)
	}

	var list []map[string]interface{}
	db.Order("frequency DESC").
		Limit(limit).
		Find(&list)

	c.JSON(http.StatusOK, response.Success(list))
}

func GetHotKeywords(ctx context.Context, c *app.RequestContext) {
	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 30
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	var list []map[string]interface{}
	database.GetDB().Table("dispute_keyword_dict").
		Select("keyword, category, frequency").
		Where("status = 1 AND updated_at >= ?", startDate).
		Order("frequency DESC").
		Limit(limit).
		Find(&list)

	c.JSON(http.StatusOK, response.Success(list))
}

func buildKeywordExtractionPrompt(text string, typeName string, maxKeywords int) string {
	typeHint := ""
	if typeName != "" {
		typeHint = fmt.Sprintf("\n已知纠纷类型为「%s」，请提取与该类型相关且具体的关键词。", typeName)
	}

	return fmt.Sprintf(`你是一个专业的矛盾纠纷分析助手。请从以下纠纷描述中提取核心关键词标签。

要求：
1. 提取%d个以内的关键词
2. 关键词要具体、有区分度，如"噪音扰民""欠薪3个月""漏水赔偿""物业费纠纷""邻里围墙争议"等
3. 不要提取过于宽泛的词如"纠纷""矛盾""问题"等
4. 关键词应覆盖：纠纷性质(如噪音/欠薪/漏水)、行为(如拖欠/侵占/骚扰)、对象(如物业/房东/雇主)、程度(如3个月/2万元/长期)
5. 返回纯JSON数组格式，不要有其他文字
6. 每个关键词2-8个字
%s

纠纷描述：
%s`, maxKeywords, typeHint, text)
}

func extractKeywordsFallback(raw string) []string {
	raw = strings.Trim(raw, "[]\"")
	parts := strings.Split(raw, ",")
	keywords := make([]string, 0)
	for _, p := range parts {
		k := strings.TrimSpace(strings.Trim(p, "\"' "))
		if k != "" && len([]rune(k)) >= 2 && len([]rune(k)) <= 16 {
			keywords = append(keywords, k)
		}
	}
	return keywords
}

func updateKeywordDict(keywords []string) {
	for _, kw := range keywords {
		kw = strings.TrimSpace(kw)
		if kw == "" {
			continue
		}

		var count int64
		database.GetDB().Table("dispute_keyword_dict").
			Where("keyword = ?", kw).
			Count(&count)

		if count > 0 {
			database.GetDB().Table("dispute_keyword_dict").
				Where("keyword = ?", kw).
				UpdateColumn("frequency", gorm.Expr("frequency + 1"))
		} else {
			category := classifyKeyword(kw)
			database.GetDB().Table("dispute_keyword_dict").Create(map[string]interface{}{
				"keyword":     kw,
				"category":    category,
				"frequency":   1,
				"source_type": "ai",
				"status":      1,
			})
		}
	}
}

func classifyKeyword(kw string) string {
	natureKeywords := []string{"噪音", "漏水", "油烟", "气味", "震动", "辐射", "粉尘", "污水", "占用", "侵占", "遮挡", "违建", "违章"}
	behaviorKeywords := []string{"拖欠", "欠薪", "拒付", "拖欠", "骚扰", "威胁", "殴打", "辱骂", "欺诈", "诈骗", "违约", "拒绝"}
	objectKeywords := []string{"物业", "房东", "租客", "雇主", "员工", "邻里", "业主", "开发商", "中介", "施工"}

	runeKw := []rune(kw)
	for _, k := range natureKeywords {
		if strings.Contains(kw, k) {
			return "纠纷性质"
		}
	}
	for _, k := range behaviorKeywords {
		if strings.Contains(kw, k) {
			return "行为"
		}
	}
	for _, k := range objectKeywords {
		if strings.Contains(kw, k) {
			return "对象"
		}
	}

	if len(runeKw) > 0 {
		last := string(runeKw[len(runeKw)-1:])
		if last == "重" || last == "大" || last == "急" || last == "久" {
			return "程度"
		}
	}

	return "纠纷性质"
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
