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
	CaseID int64 `form:"caseId"`
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
	evidence := map[string]interface{}{
		"id":          evidenceID,
		"case_id":     req.CaseID,
		"file_name":   file.Filename,
		"file_path":   objectName,
		"file_url":    fileURL,
		"file_type":   fileType,
		"file_size":   file.Size,
		"file_ext":    ext,
		"mime_type":   file.Header.Get("Content-Type"),
		"remark":      req.Remark,
		"sort_order":  req.SortOrder,
		"upload_from": req.UploadFrom,
		"uploader_id": userInfo.UserID,
		"uploader_name": userInfo.RealName,
		"organization_id": userInfo.OrganizationID,
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
			}
			mq.SendMessage(constants.MQTopicAIProcess, msg)
		}
	}()

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"id":       evidenceID,
		"fileUrl":  fileURL,
		"fileName": file.Filename,
		"fileType": fileType,
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

	if req.Keyword != "" {
		db = db.Where("file_name LIKE ? OR remark LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("sort_order ASC, created_at DESC").
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
		}
		if size, ok := item["file_size"].(int64); ok {
			item["file_size_format"] = formatFileSize(size)
		}
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
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
