package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	common "github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
)

type EvidenceUploadRequest struct {
	CaseID     int64  `form:"caseId"`
	FileType   int32  `form:"fileType" binding:"required"`
	Remark     string `form:"remark"`
	SortOrder  int32  `form:"sortOrder"`
	UploadFrom int32  `form:"uploadFrom"`
}

type EvidenceListRequest struct {
	common.BaseQuery
	CaseID            int64  `form:"caseId"`
	Category          int    `form:"category"`
	GroupByCategory   bool   `form:"groupByCategory"`
}

func UploadEvidence(ctx context.Context, c *app.RequestContext) {
	var req EvidenceUploadRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("请选择要上传的文件"))
		return
	}

	const maxFileSize = 50 * 1024 * 1024
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, response.BadRequest("文件大小不能超过50MB"))
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true,
		".mp4": true, ".avi": true, ".mov": true, ".mkv": true,
		".mp3": true, ".wav": true, ".m4a": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true, ".txt": true,
	}

	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, response.BadRequest("不支持的文件格式"))
		return
	}

	fileType := req.FileType
	if fileType == 0 {
		if isImage(ext) {
			fileType = constants.FileTypeImage
		} else if isVideo(ext) {
			fileType = constants.FileTypeVideo
		} else if isAudio(ext) {
			fileType = constants.FileTypeAudio
		} else {
			fileType = constants.FileTypeDocument
		}
	}

	objectName := fmt.Sprintf("%s/%d/%s%s",
		constants.MinIOPathEvidence,
		userInfo.OrganizationID,
		utils.GenerateUUID(),
		ext,
	)

	src, err := file.Open()
	if err != nil {
		logger.Error("Open file failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("文件读取失败"))
		return
	}
	defer src.Close()

	fileContent, err := io.ReadAll(src)
	if err != nil {
		logger.Error("Read file failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("文件读取失败"))
		return
	}

	fileURL := fmt.Sprintf("/api/v1/public/file/%s", objectName)

	evidenceID := utils.GenerateID()

	classifyResult := ClassifyEvidenceByFileName(file.Filename, fileType)

	evidence := map[string]interface{}{
		"id":                 evidenceID,
		"case_id":            req.CaseID,
		"file_name":          file.Filename,
		"file_path":          objectName,
		"file_url":           fileURL,
		"file_type":          fileType,
		"file_size":          file.Size,
		"file_ext":           ext,
		"mime_type":          file.Header.Get("Content-Type"),
		"remark":             req.Remark,
		"sort_order":         req.SortOrder,
		"upload_from":        req.UploadFrom,
		"uploader_id":        userInfo.UserID,
		"uploader_name":      userInfo.RealName,
		"organization_id":    userInfo.OrganizationID,
		"evidence_category":  classifyResult.Category,
		"ai_category":        classifyResult.Category,
		"ai_confidence":      classifyResult.Confidence,
		"ai_keywords":        classifyResult.Keywords,
		"ai_processed":       constants.AIEvidenceProcessing,
	}

	tx := database.GetDB().Begin()
	if err := tx.Table("dispute_evidence").Create(evidence).Error; err != nil {
		tx.Rollback()
		logger.Error("Save evidence record failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("证据保存失败"))
		return
	}

	if req.CaseID > 0 {
		var caseNo string
		tx.Table("dispute_case").
			Select("case_no").
			Where("id = ?", req.CaseID).
			Scan(&caseNo)

		tx.Table("dispute_evidence").
			Where("id = ?", evidenceID).
			Update("case_no", caseNo)

		history := map[string]interface{}{
			"case_id":          req.CaseID,
			"case_no":          caseNo,
			"operation_type":   "EVIDENCE_UPLOAD",
			"operation_detail": fmt.Sprintf("上传证据: %s", file.Filename),
			"operator_id":      userInfo.UserID,
			"operator_name":    userInfo.RealName,
		}
		tx.Table("dispute_case_history").Create(history)

		cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, req.CaseID)
		cache.Del(ctx, cacheKey)
	}

	tx.Commit()

	go func() {
		if fileType == constants.FileTypeImage {
			msg := map[string]interface{}{
				"evidenceId": evidenceID,
				"filePath":   objectName,
				"fileType":   fileType,
				"fileName":   file.Filename,
			}
			mq.SendMessage(constants.MQTopicAIProcess, msg)
		}

		classifyMsg := map[string]interface{}{
			"evidenceId": evidenceID,
			"filePath":   objectName,
			"fileType":   fileType,
			"fileName":   file.Filename,
			"fileExt":    ext,
			"category":   classifyResult.Category,
			"confidence": classifyResult.Confidence,
			"keywords":   classifyResult.Keywords,
		}
		mq.SendMessage(constants.MQTopicEvidenceClassify, classifyMsg)
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id":              evidenceID,
		"fileUrl":         fileURL,
		"fileName":        file.Filename,
		"fileType":        fileType,
		"evidenceCategory": classifyResult.Category,
		"categoryName":    GetEvidenceCategoryName(classifyResult.Category),
		"aiConfidence":    classifyResult.Confidence,
		"aiKeywords":      classifyResult.Keywords,
	}, "上传成功"))
}

func GetEvidenceList(ctx context.Context, c *app.RequestContext) {
	var req EvidenceListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	db := database.GetDB().Table("dispute_evidence").
		Where("deleted_at IS NULL")

	if req.CaseID > 0 {
		db = db.Where("case_id = ?", req.CaseID)
	} else {
		db = db.Where("uploader_id = ? OR organization_id = ?", userInfo.UserID, userInfo.OrganizationID)
	}

	if req.Category > 0 {
		db = db.Where("evidence_category = ?", req.Category)
	}

	if req.Keyword != "" {
		db = db.Where("file_name LIKE ? OR remark LIKE ? OR ai_keywords LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("evidence_category ASC, sort_order ASC, created_at DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	fileTypeMap := map[int]string{
		constants.FileTypeImage:    "图片",
		constants.FileTypeVideo:    "视频",
		constants.FileTypeAudio:    "音频",
		constants.FileTypeDocument: "文档",
		constants.FileTypeOther:    "其他",
	}

	for _, item := range list {
		if ft, ok := item["file_type"].(int); ok {
			item["file_type_name"] = fileTypeMap[ft]
		} else if ft, ok := item["file_type"].(int64); ok {
			item["file_type_name"] = fileTypeMap[int(ft)]
		}
		if size, ok := item["file_size"].(int64); ok {
			item["file_size_format"] = formatFileSize(size)
		}
		if cat, ok := item["evidence_category"].(int); ok {
			item["category_name"] = GetEvidenceCategoryName(cat)
		} else if cat, ok := item["evidence_category"].(int64); ok {
			item["category_name"] = GetEvidenceCategoryName(int(cat))
		}
		if aiCat, ok := item["ai_category"].(int); ok {
			item["ai_category_name"] = GetEvidenceCategoryName(aiCat)
		} else if aiCat, ok := item["ai_category"].(int64); ok {
			item["ai_category_name"] = GetEvidenceCategoryName(int(aiCat))
		}
		if manualCat, ok := item["manual_category"].(int); ok && manualCat > 0 {
			item["category_name"] = GetEvidenceCategoryName(manualCat)
			item["is_manual_updated"] = true
		} else if manualCat, ok := item["manual_category"].(int64); ok && manualCat > 0 {
			item["category_name"] = GetEvidenceCategoryName(int(manualCat))
			item["is_manual_updated"] = true
		}
	}

	var result interface{}
	if req.GroupByCategory {
		grouped := make(map[int]map[string]interface{})
		for category := range constants.EvidenceCategoryMap {
			grouped[category] = map[string]interface{}{
				"category":     category,
				"categoryName": GetEvidenceCategoryName(category),
				"list":         make([]map[string]interface{}, 0),
				"total":        0,
			}
		}

		for _, item := range list {
			var cat int
			if c, ok := item["evidence_category"].(int); ok {
				cat = c
			} else if c, ok := item["evidence_category"].(int64); ok {
				cat = int(c)
			}
			if _, exists := grouped[cat]; !exists {
				grouped[cat] = map[string]interface{}{
					"category":     cat,
					"categoryName": GetEvidenceCategoryName(cat),
					"list":         make([]map[string]interface{}, 0),
					"total":        0,
				}
			}
			grouped[cat]["list"] = append(grouped[cat]["list"].([]map[string]interface{}), item)
			grouped[cat]["total"] = grouped[cat]["total"].(int) + 1
		}

		groups := make([]map[string]interface{}, 0)
		for category := 0; category <= 9; category++ {
			if g, exists := grouped[category]; exists {
				groups = append(groups, g)
			}
		}

		result = map[string]interface{}{
			"groups": groups,
			"total":  total,
		}
	} else {
		result = list
	}

	c.JSON(http.StatusOK, response.Page(result, total, req.Page, req.PageSize))
}

func DeleteEvidence(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userInfo := middleware.GetUserInfo(c)

	var evidence struct {
		CaseID      int64  `gorm:"column:case_id"`
		CaseNo      string `gorm:"column:case_no"`
		FileName    string `gorm:"column:file_name"`
		UploaderID  int64  `gorm:"column:uploader_id"`
	}

	result := database.GetDB().Table("dispute_evidence").
		Where("id = ?", id).
		First(&evidence)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, response.NotFound("证据不存在"))
		return
	}

	if evidence.UploaderID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限删除此证据"))
		return
	}

	tx := database.GetDB().Begin()

	tx.Table("dispute_evidence").
		Where("id = ?", id).
		Update("deleted_at", time.Now())

	if evidence.CaseID > 0 {
		history := map[string]interface{}{
			"case_id":          evidence.CaseID,
			"case_no":          evidence.CaseNo,
			"operation_type":   "EVIDENCE_DELETE",
			"operation_detail": fmt.Sprintf("删除证据: %s", evidence.FileName),
			"operator_id":      userInfo.UserID,
			"operator_name":    userInfo.RealName,
		}
		tx.Table("dispute_case_history").Create(history)

		cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, evidence.CaseID)
		cache.Del(ctx, cacheKey)
	}

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "删除成功"))
}

func BatchDeleteEvidence(ctx context.Context, c *app.RequestContext) {
	var req struct {
		IDs []int64 `json:"ids" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var evidences []map[string]interface{}
	database.GetDB().Table("dispute_evidence").
		Select("id, case_id, case_no, file_name, uploader_id").
		Where("id IN ?", req.IDs).
		Find(&evidences)

	tx := database.GetDB().Begin()

	now := time.Now()
	for _, ev := range evidences {
		uploaderID := ev["uploader_id"].(int64)
		if uploaderID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
			continue
		}

		tx.Table("dispute_evidence").
			Where("id = ?", ev["id"]).
			Update("deleted_at", now)

		if caseID, ok := ev["case_id"].(int64); ok && caseID > 0 {
			history := map[string]interface{}{
				"case_id":          caseID,
				"case_no":          ev["case_no"],
				"operation_type":   "EVIDENCE_DELETE",
				"operation_detail": fmt.Sprintf("删除证据: %s", ev["file_name"]),
				"operator_id":      userInfo.UserID,
				"operator_name":    userInfo.RealName,
			}
			tx.Table("dispute_case_history").Create(history)

			cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, caseID)
			cache.Del(ctx, cacheKey)
		}
	}

	tx.Commit()

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "批量删除成功"))
}

func UpdateEvidenceRemark(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Remark    string `json:"remark"`
		SortOrder int32  `json:"sortOrder"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var evidence struct {
		CaseID     int64 `gorm:"column:case_id"`
		UploaderID int64 `gorm:"column:uploader_id"`
	}

	database.GetDB().Table("dispute_evidence").
		Where("id = ?", id).
		First(&evidence)

	if evidence.UploaderID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限修改此证据"))
		return
	}

	updates := map[string]interface{}{}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.SortOrder > 0 {
		updates["sort_order"] = req.SortOrder
	}

	database.GetDB().Table("dispute_evidence").
		Where("id = ?", id).
		Updates(updates)

	if evidence.CaseID > 0 {
		cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, evidence.CaseID)
		cache.Del(ctx, cacheKey)
	}

	c.JSON(http.StatusOK, response.SuccessWithMessage(nil, "更新成功"))
}

func UpdateEvidenceCategory(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Category int `json:"category" binding:"required"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	if req.Category < 0 || req.Category > 9 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的证据类别"))
		return
	}

	userInfo := middleware.GetUserInfo(c)

	var evidence struct {
		CaseID     int64 `gorm:"column:case_id"`
		UploaderID int64 `gorm:"column:uploader_id"`
	}
	database.GetDB().Table("dispute_evidence").
		Where("id = ?", id).
		First(&evidence)

	if evidence.UploaderID != userInfo.UserID && userInfo.Role > constants.RoleLeader {
		c.JSON(http.StatusForbidden, response.Forbidden("无权限修改此证据类别"))
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"evidence_category":   req.Category,
		"manual_category":     req.Category,
		"manual_updated_at":   now,
		"manual_updated_by":   userInfo.UserID,
	}

	if err := database.GetDB().Table("dispute_evidence").
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		logger.Error("Update evidence category failed",
			logger.Int64("evidenceId", id),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("更新失败"))
		return
	}

	if evidence.CaseID > 0 {
		cacheKey := fmt.Sprintf("%s%d", constants.RedisKeyPrefixCase, evidence.CaseID)
		cache.Del(ctx, cacheKey)

		history := map[string]interface{}{
			"case_id":          evidence.CaseID,
			"operation_type":   "EVIDENCE_CATEGORY_UPDATE",
			"operation_detail": fmt.Sprintf("修正证据类别: %s → %s",
				c.Param("id"), GetEvidenceCategoryName(req.Category)),
			"operator_id":      userInfo.UserID,
			"operator_name":    userInfo.RealName,
		}
		var caseNo string
		database.GetDB().Table("dispute_case").
			Select("case_no").Where("id = ?", evidence.CaseID).Scan(&caseNo)
		history["case_no"] = caseNo
		database.GetDB().Table("dispute_case_history").Create(history)
	}

	logger.Info("Evidence category manually updated",
		logger.Int64("evidenceId", id),
		logger.Int("newCategory", req.Category),
		logger.String("categoryName", GetEvidenceCategoryName(req.Category)),
		logger.Int64("operatorId", userInfo.UserID))

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"category":     req.Category,
		"categoryName": GetEvidenceCategoryName(req.Category),
	}, "类别修正成功"))
}

func isImage(ext string) bool {
	imageExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true}
	return imageExts[ext]
}

func isVideo(ext string) bool {
	videoExts := map[string]bool{".mp4": true, ".avi": true, ".mov": true, ".mkv": true}
	return videoExts[ext]
}

func isAudio(ext string) bool {
	audioExts := map[string]bool{".mp3": true, ".wav": true, ".m4a": true}
	return audioExts[ext]
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

type EvidenceClassifyResult struct {
	Category     int
	Confidence   float64
	Keywords     string
	MatchedRules []string
}

func ClassifyEvidenceByFileName(fileName string, fileType int32) EvidenceClassifyResult {
	result := EvidenceClassifyResult{
		Category:     constants.EvidenceCategoryUncategorized,
		Confidence:   0.0,
		MatchedRules: make([]string, 0),
	}

	lowerName := strings.ToLower(fileName)

	categoryRules := []struct {
		Category int
		Keywords []string
		Priority int
	}{
		{
			Category: constants.EvidenceCategoryIDCard,
			Keywords: []string{"身份证", "id", "idcard", "id_card", "身份", "证件照", "身份证正面", "身份证反面", "身份证明"},
			Priority: 100,
		},
		{
			Category: constants.EvidenceCategoryContract,
			Keywords: []string{"合同", "协议", "contract", "agreement", "合作协议", "租赁合同", "买卖合同", "借款合同", "劳动合同", "服务协议", "约定书", "承诺书"},
			Priority: 90,
		},
		{
			Category: constants.EvidenceCategoryReceipt,
			Keywords: []string{"收据", "收条", "receipt", "收款", "押金条", "定金条", "保证金", "付款凭证", "转账凭证", "汇款凭证", "支付凭证", "收款凭证"},
			Priority: 85,
		},
		{
			Category: constants.EvidenceCategoryInvoice,
			Keywords: []string{"发票", "invoice", "增值税", "专票", "普票", "开票", "税票", "fapiao"},
			Priority: 80,
		},
		{
			Category: constants.EvidenceCategoryChatRecord,
			Keywords: []string{"聊天", "微信", "wechat", "wx", "qq", "短信", "sms", "message", "聊天记录", "对话", "通话记录", "微信聊天", "微信截图", "聊天截图", "短信截图"},
			Priority: 75,
		},
		{
			Category: constants.EvidenceCategoryPhoto,
			Keywords: []string{"照片", "photo", "img", "image", "图片", "现场", "场景", "拍照", "截图", "screenshot", "证据照片", "现场照片", "事故照片", "伤情照片"},
			Priority: 60,
		},
		{
			Category: constants.EvidenceCategoryMedia,
			Keywords: []string{"录音", "录像", "video", "audio", "语音", "通话录音", "现场录音", "录像视频", "mp3", "mp4", "mov", "avi", "mkv", "wav", "m4a"},
			Priority: 70,
		},
		{
			Category: constants.EvidenceCategoryCertificate,
			Keywords: []string{"证明", "certificate", "证书", "房产证", "土地证", "驾驶证", "行驶证", "结婚证", "离婚证", "户口本", "营业执照", "经营许可证", "资质证书", "产权证明"},
			Priority: 65,
		},
	}

	highestPriority := 0
	bestCategory := constants.EvidenceCategoryUncategorized
	var matchedKeywords []string

	for _, rule := range categoryRules {
		for _, keyword := range rule.Keywords {
			if strings.Contains(lowerName, keyword) {
				if rule.Priority > highestPriority {
					highestPriority = rule.Priority
					bestCategory = rule.Category
					matchedKeywords = append(matchedKeywords, keyword)
				}
				result.MatchedRules = append(result.MatchedRules, keyword)
			}
		}
	}

	if bestCategory == constants.EvidenceCategoryUncategorized {
		switch fileType {
		case constants.FileTypeImage:
			bestCategory = constants.EvidenceCategoryPhoto
			result.MatchedRules = append(result.MatchedRules, "image_default")
		case constants.FileTypeVideo, constants.FileTypeAudio:
			bestCategory = constants.EvidenceCategoryMedia
			result.MatchedRules = append(result.MatchedRules, "media_default")
		case constants.FileTypeDocument:
			bestCategory = constants.EvidenceCategoryOther
			result.MatchedRules = append(result.MatchedRules, "document_default")
		}
	}

	confidence := float64(highestPriority) / 100.0
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence == 0 {
		confidence = 0.3
	}

	result.Category = bestCategory
	result.Confidence = confidence
	result.Keywords = strings.Join(matchedKeywords, ",")

	return result
}

func ClassifyEvidenceByContent(fileName string, fileType int32, fileContent []byte) EvidenceClassifyResult {
	result := ClassifyEvidenceByFileName(fileName, fileType)

	if fileType == constants.FileTypeImage && len(fileContent) > 0 {
		fileSize := len(fileContent)
		if fileSize > 500*1024 {
			if result.Category == constants.EvidenceCategoryPhoto {
				result.Confidence = result.Confidence*0.7 + 0.3
				result.Keywords = result.Keywords + ",large_image"
			}
		}

		if len(result.MatchedRules) > 0 && result.Confidence < 0.8 {
			result.Confidence = result.Confidence*0.8 + 0.16
		}
	}

	return result
}

func GetEvidenceCategoryName(category int) string {
	if name, ok := constants.EvidenceCategoryMap[category]; ok {
		return name
	}
	return "其他"
}

func ProcessEvidenceAIClassify(ctx context.Context, evidenceID int64, filePath string, fileName string, fileType int32) error {
	now := time.Now()

	db := database.GetDB()

	var existingData struct {
		EvidenceCategory int     `gorm:"column:evidence_category"`
		AICategory       int     `gorm:"column:ai_category"`
		AIConfidence     float64 `gorm:"column:ai_confidence"`
		AIKeywords       string  `gorm:"column:ai_keywords"`
	}
	err := db.Table("dispute_evidence").
		Select("evidence_category, ai_category, ai_confidence, ai_keywords").
		Where("id = ?", evidenceID).
		Scan(&existingData).Error
	if err != nil {
		logger.Error("Query existing evidence for AI classify failed",
			logger.Int64("evidenceId", evidenceID),
			logger.Error(err))
		return err
	}

	classifyResult := ClassifyEvidenceByFileName(fileName, fileType)

	if existingData.EvidenceCategory != constants.EvidenceCategoryUncategorized &&
		existingData.EvidenceCategory != classifyResult.Category {
		classifyResult.Confidence = classifyResult.Confidence * 0.9
	}

	updates := map[string]interface{}{
		"ai_processed":  constants.AIEvidenceProcessDone,
		"ai_processed_at": now,
		"ai_category":    classifyResult.Category,
		"ai_confidence":  classifyResult.Confidence,
		"ai_keywords":    classifyResult.Keywords,
	}

	if existingData.EvidenceCategory == constants.EvidenceCategoryUncategorized ||
		existingData.EvidenceCategory == 0 {
		updates["evidence_category"] = classifyResult.Category
	}

	if err := db.Table("dispute_evidence").
		Where("id = ?", evidenceID).
		Updates(updates).Error; err != nil {
		logger.Error("Update evidence AI classify result failed",
			logger.Int64("evidenceId", evidenceID),
			logger.Error(err))

		db.Table("dispute_evidence").
			Where("id = ?", evidenceID).
			Update("ai_processed", constants.AIEvidenceProcessFailed)
		return err
	}

	logger.Info("Evidence AI classify completed",
		logger.Int64("evidenceId", evidenceID),
		logger.Int("category", classifyResult.Category),
		logger.String("categoryName", GetEvidenceCategoryName(classifyResult.Category)),
		logger.Float64("confidence", classifyResult.Confidence),
		logger.String("keywords", classifyResult.Keywords))

	return nil
}

func HandleEvidenceClassifyMQ(ctx context.Context, msg map[string]interface{}) {
	evidenceID, _ := msg["evidenceId"].(int64)
	if evidenceID == 0 {
		if idStr, ok := msg["evidenceId"].(string); ok {
			evidenceID, _ = strconv.ParseInt(idStr, 10, 64)
		}
	}
	if evidenceID == 0 {
		logger.Error("Invalid evidenceId in classify MQ message")
		return
	}

	filePath, _ := msg["filePath"].(string)
	fileName, _ := msg["fileName"].(string)

	var fileType int32
	if ft, ok := msg["fileType"].(int); ok {
		fileType = int32(ft)
	} else if ft, ok := msg["fileType"].(int32); ok {
		fileType = ft
	} else if ft, ok := msg["fileType"].(float64); ok {
		fileType = int32(ft)
	}

	ProcessEvidenceAIClassify(ctx, evidenceID, filePath, fileName, fileType)
}
