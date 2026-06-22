package handler

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/aliyun"
	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type SubmitTranscribeRequest struct {
	FileUrl           string `json:"fileUrl"`
	FileBase64        string `json:"fileBase64"`
	FileName          string `json:"fileName"`
	Format            string `json:"format"`
	CaseIdStr         string `json:"caseId"`
	RecordIdStr       string `json:"recordId"`
	EnableDiarization bool   `json:"enableDiarization"`
	SpeakerCount      int    `json:"speakerCount"`
}

type TranscribeTaskQuery struct {
	model.BaseQuery
	Status int   `form:"status"`
	CaseID int64 `form:"caseId"`
}

type TingwuCallbackRequest struct {
	TaskId         string              `json:"TaskId"`
	Status         string              `json:"Status"`
	TranscriptText string              `json:"TranscriptText"`
	Duration       int                 `json:"Duration"`
	WordCount      int                 `json:"WordCount"`
	Sentences      []aliyun.TingwuSentence `json:"Sentences"`
	RequestId      string              `json:"RequestId"`
	Code           string              `json:"Code"`
	Message        string              `json:"Message"`
}

type SyncTranscribeRequest struct {
	FileName   string `json:"fileName"`
	FileBase64 string `json:"fileBase64"`
	FileUrl    string `json:"fileUrl"`
	Format     string `json:"format"`
}

func SubmitTranscribeTask(ctx context.Context, c *app.RequestContext) {
	var req SubmitTranscribeRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	if req.FileUrl == "" && req.FileBase64 == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("音频文件URL或base64数据不能为空"))
		return
	}

	if req.Format == "" {
		req.Format = "mp3"
	}
	if req.SpeakerCount <= 0 {
		req.SpeakerCount = 2
	}

	caseId, _ := strconv.ParseInt(req.CaseIdStr, 10, 64)
	recordId, _ := strconv.ParseInt(req.RecordIdStr, 10, 64)

	var createdBy int64
	if userInfo := middleware.GetUserInfo(c); userInfo != nil {
		createdBy = userInfo.UserID
	}

	fileUrl := req.FileUrl
	fileSize := int64(0)

	if req.FileBase64 != "" && fileUrl == "" {
		base64Str := req.FileBase64
		if strings.Contains(base64Str, ",") {
			base64Str = strings.SplitN(base64Str, ",", 2)[1]
		}

		audioData, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频base64解码失败"))
			return
		}

		fileSize = int64(len(audioData))

		uploadedUrl, err := uploadAudioToMinIO(ctx, audioData, req.Format, req.FileName)
		if err != nil {
			logger.Error("Upload audio to MinIO failed", logger.Error(err))
			c.JSON(http.StatusInternalServerError, response.ServerError("上传音频文件失败: "+err.Error()))
			return
		}
		fileUrl = uploadedUrl
	}

	taskId, requestId, err := aliyun.GetTingwuClient().SubmitTask(
		fileUrl,
		req.FileName,
		req.Format,
		req.EnableDiarization,
		req.SpeakerCount,
		"",
		"",
	)
	if err != nil {
		logger.Error("Submit transcribe task failed",
			zap.String("fileUrl", fileUrl),
			logger.Error(err),
		)
		c.JSON(http.StatusInternalServerError, response.ServerError("提交转写任务失败: "+err.Error()))
		return
	}

	task := &model.VoiceTranscribeTask{
		TaskID:            taskId,
		CaseID:            caseId,
		RecordID:          recordId,
		CreatedBy:         createdBy,
		FileName:          req.FileName,
		FileURL:           fileUrl,
		FileSize:          fileSize,
		Format:            req.Format,
		Status:            model.TranscribeStatusProcessing,
		TaskType:          2,
		SpeakerCount:      req.SpeakerCount,
		EnableDiarization: req.EnableDiarization,
		RequestID:         requestId,
	}

	db := database.GetDB()
	if err := db.Create(task).Error; err != nil {
		logger.Error("Create transcribe task record failed",
			zap.String("taskId", taskId),
			logger.Error(err),
		)
		c.JSON(http.StatusInternalServerError, response.ServerError("创建任务记录失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"taskId": taskId,
		"status": "processing",
		"id":     task.ID,
	}))
}

func uploadAudioToMinIO(ctx context.Context, audioData []byte, format string, fileName string) (string, error) {
	minioClient := database.GetMinioClient()
	if minioClient == nil {
		return "", nil
	}

	cfg := config.GetConfig().MinIO
	if cfg.Bucket == "" {
		return "", nil
	}

	now := time.Now()
	datePath := now.Format("200601")
	uuid := utils.GenerateUUID()
	objectName := "voice/" + datePath + "/" + uuid + "." + format

	reader := bytes.NewReader(audioData)
	_, err := minioClient.PutObject(
		ctx,
		cfg.Bucket,
		objectName,
		reader,
		int64(len(audioData)),
		minio.PutObjectOptions{
			ContentType: "audio/" + format,
		},
	)
	if err != nil {
		return "", err
	}

	scheme := "http"
	if cfg.UseSSL {
		scheme = "https"
	}
	fileUrl := scheme + "://" + cfg.Endpoint + "/" + cfg.Bucket + "/" + objectName

	return fileUrl, nil
}

func GetTranscribeTask(ctx context.Context, c *app.RequestContext) {
	taskId := c.Param("taskId")
	if taskId == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("任务ID不能为空"))
		return
	}

	db := database.GetDB()
	var task model.VoiceTranscribeTask
	if err := db.Where("task_id = ?", taskId).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("转写任务不存在"))
		return
	}

	if task.Status == model.TranscribeStatusCompleted || task.Status == model.TranscribeStatusFailed || task.Status == model.TranscribeStatusCanceled {
		returnTranscribeTask(c, &task)
		return
	}

	status, transcriptText, duration, wordCount, sentences, err := aliyun.GetTingwuClient().GetTaskResult(taskId)
	if err != nil {
		logger.Warn("Get task result from tingwu failed, return local data",
			zap.String("taskId", taskId),
			logger.Error(err),
		)
		returnTranscribeTask(c, &task)
		return
	}

	now := time.Now()
	updates := map[string]interface{}{}

	switch status {
	case "Queuing":
		updates["status"] = model.TranscribeStatusQueuing
	case "Processing":
		updates["status"] = model.TranscribeStatusProcessing
	case "Completed":
		updates["status"] = model.TranscribeStatusCompleted
		updates["transcript_text"] = transcriptText
		updates["duration"] = duration
		updates["word_count"] = wordCount
		updates["completed_at"] = &now
		if sentencesJSON, err := json.Marshal(sentences); err == nil {
			updates["sentences"] = string(sentencesJSON)
		}
	case "Failed":
		updates["status"] = model.TranscribeStatusFailed
		updates["error_msg"] = "转写失败"
	}

	if len(updates) > 0 {
		if err := db.Model(&task).Updates(updates).Error; err != nil {
			logger.Error("Update transcribe task failed",
				zap.String("taskId", taskId),
				logger.Error(err),
			)
		}

		if status == "Completed" && task.RecordID > 0 {
			updateMediationRecordTranscript(task.RecordID, transcriptText, duration)
		}

		for k, v := range updates {
			switch k {
			case "status":
				task.Status = v.(int)
			case "transcript_text":
				task.TranscriptText = v.(string)
			case "duration":
				task.Duration = v.(int)
			case "word_count":
				task.WordCount = v.(int)
			case "completed_at":
				task.CompletedAt = v.(*time.Time)
			case "sentences":
				task.Sentences = v.(string)
			case "error_msg":
				task.ErrorMessage = v.(string)
			}
		}
	}

	returnTranscribeTask(c, &task)
}

func returnTranscribeTask(c *app.RequestContext, task *model.VoiceTranscribeTask) {
	result := make(map[string]interface{})
	taskJSON, _ := json.Marshal(task)
	json.Unmarshal(taskJSON, &result)

	result["status_name"] = model.TranscribeStatusMap[task.Status]
	result["statusCode"] = model.TranscribeStatusCodeMap[task.Status]

	if task.Sentences != "" {
		var sentences []model.Sentence
		if err := json.Unmarshal([]byte(task.Sentences), &sentences); err == nil {
			result["sentences"] = sentences
		}
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func GetMyTranscribeTasks(ctx context.Context, c *app.RequestContext) {
	var req TranscribeTaskQuery
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("参数错误: "+err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized("请先登录"))
		return
	}

	db := database.GetDB().Model(&model.VoiceTranscribeTask{}).
		Where("created_by = ?", userInfo.UserID)

	if req.Status > 0 {
		db = db.Where("status = ?", req.Status)
	}
	if req.CaseID > 0 {
		db = db.Where("case_id = ?", req.CaseID)
	}
	if req.Keyword != "" {
		db = db.Where("file_name LIKE ? OR transcript_text LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []model.VoiceTranscribeTask
	db.Order(req.GetSort()).
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	resultList := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		result := make(map[string]interface{})
		itemJSON, _ := json.Marshal(item)
		json.Unmarshal(itemJSON, &result)
		result["status_name"] = model.TranscribeStatusMap[item.Status]
		result["statusCode"] = model.TranscribeStatusCodeMap[item.Status]
		resultList = append(resultList, result)
	}

	c.JSON(http.StatusOK, response.Page(resultList, total, req.Page, req.PageSize))
}

func TranscribeCallback(ctx context.Context, c *app.RequestContext) {
	body, err := io.ReadAll(c.Request.Body())
	if err != nil {
		logger.Error("Read tingwu callback body failed", logger.Error(err))
		c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "message": "success"})
		return
	}

	logger.Info("Received Tingwu callback",
		zap.String("body", string(body)),
	)

	var callbackData TingwuCallbackRequest
	if err := json.Unmarshal(body, &callbackData); err != nil {
		logger.Warn("Parse tingwu callback failed",
			zap.String("body", string(body)),
			logger.Error(err),
		)
		c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "message": "success"})
		return
	}

	if callbackData.TaskId == "" {
		c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "message": "success"})
		return
	}

	db := database.GetDB()
	var task model.VoiceTranscribeTask
	if err := db.Where("task_id = ?", callbackData.TaskId).First(&task).Error; err != nil {
		logger.Warn("Transcribe task not found for callback",
			zap.String("taskId", callbackData.TaskId),
		)
		c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "message": "success"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{}

	switch callbackData.Status {
	case "Queuing":
		updates["status"] = model.TranscribeStatusQueuing
	case "Processing":
		updates["status"] = model.TranscribeStatusProcessing
	case "Completed":
		updates["status"] = model.TranscribeStatusCompleted
		updates["transcript_text"] = callbackData.TranscriptText
		updates["duration"] = callbackData.Duration
		updates["word_count"] = callbackData.WordCount
		updates["completed_at"] = &now
		if sentencesJSON, err := json.Marshal(callbackData.Sentences); err == nil {
			updates["sentences"] = string(sentencesJSON)
		}
	case "Failed":
		updates["status"] = model.TranscribeStatusFailed
		if callbackData.Message != "" {
			updates["error_msg"] = callbackData.Message
		} else {
			updates["error_msg"] = "转写失败"
		}
	}

	if callbackData.RequestId != "" {
		updates["request_id"] = callbackData.RequestId
	}

	if len(updates) > 0 {
		if err := db.Model(&task).Updates(updates).Error; err != nil {
			logger.Error("Update transcribe task from callback failed",
				zap.String("taskId", callbackData.TaskId),
				logger.Error(err),
			)
		}

		if callbackData.Status == "Completed" && task.RecordID > 0 {
			updateMediationRecordTranscript(task.RecordID, callbackData.TranscriptText, callbackData.Duration)
		}
	}

	c.JSON(http.StatusOK, map[string]interface{}{"code": 0, "message": "success"})
}

func updateMediationRecordTranscript(recordID int64, transcriptText string, duration int) {
	db := database.GetDB()
	updates := map[string]interface{}{
		"transcript_text": transcriptText,
	}
	if duration > 0 {
		updates["duration"] = duration
	}
	if err := db.Table("mediation_record").Where("id = ?", recordID).Updates(updates).Error; err != nil {
		logger.Error("Update mediation record transcript failed",
			zap.Int64("recordId", recordID),
			logger.Error(err),
		)
	}
}

func SyncTranscribe(ctx context.Context, c *app.RequestContext) {
	contentType := string(c.ContentType())
	var audioData []byte
	var format string
	var fileName string

	if strings.Contains(contentType, "multipart/form-data") {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("请选择要识别的音频文件"))
			return
		}

		const maxFileSize = 10 * 1024 * 1024
		if file.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频文件大小不能超过10MB"))
			return
		}

		ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
		allowedFormats := map[string]bool{
			"mp3": true, "wav": true, "m4a": true, "amr": true, "aac": true, "webm": true,
		}
		if !allowedFormats[ext] {
			c.JSON(http.StatusBadRequest, response.BadRequest(
				"不支持的音频格式，支持格式：mp3, wav, m4a, amr, aac, webm",
			))
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ServerError("文件读取失败"))
			return
		}
		defer src.Close()

		data, err := io.ReadAll(src)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.ServerError("文件读取失败"))
			return
		}

		audioData = data
		format = ext
		fileName = file.Filename
	} else {
		var req SyncTranscribeRequest
		if err := c.BindAndValidate(&req); err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("参数错误: "+err.Error()))
			return
		}

		if req.FileBase64 == "" && req.FileUrl == "" {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频数据或文件URL不能为空"))
			return
		}

		if req.FileUrl != "" {
			submitAsyncTaskByURL(c, req.FileUrl, req.FileName, req.Format)
			return
		}

		if req.Format == "" && req.FileName != "" {
			req.Format = strings.ToLower(strings.TrimPrefix(filepath.Ext(req.FileName), "."))
		}

		allowedFormats := map[string]bool{
			"mp3": true, "wav": true, "m4a": true, "amr": true, "aac": true, "webm": true,
		}
		if !allowedFormats[req.Format] {
			c.JSON(http.StatusBadRequest, response.BadRequest(
				"不支持的音频格式，支持格式：mp3, wav, m4a, amr, aac, webm",
			))
			return
		}

		base64Str := req.FileBase64
		if strings.Contains(base64Str, ",") {
			base64Str = strings.SplitN(base64Str, ",", 2)[1]
		}

		data, err := base64.StdEncoding.DecodeString(base64Str)
		if err != nil {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频base64解码失败"))
			return
		}

		const maxFileSize = 10 * 1024 * 1024
		if len(data) > maxFileSize {
			c.JSON(http.StatusBadRequest, response.BadRequest("音频文件大小不能超过10MB"))
			return
		}

		audioData = data
		format = req.Format
		fileName = req.FileName
		if fileName == "" {
			fileName = "audio." + format
		}
	}

	const syncMaxSize = 1 * 1024 * 1024
	if len(audioData) > syncMaxSize {
		submitAsyncTaskByData(ctx, c, audioData, fileName, format)
		return
	}

	client := aliyun.GetVoiceClient()
	result, err := client.RecognizeSpeech(audioData, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ServerError("语音识别失败："+err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"text":     result.Text,
		"duration": result.Duration,
		"taskId":   result.TaskID,
		"format":   format,
		"fileSize": len(audioData),
		"fileName": fileName,
		"sync":     true,
	}))
}

func submitAsyncTaskByData(ctx context.Context, c *app.RequestContext, audioData []byte, fileName, format string) {
	var createdBy int64
	if userInfo := middleware.GetUserInfo(c); userInfo != nil {
		createdBy = userInfo.UserID
	}

	fileUrl, err := uploadAudioToMinIO(ctx, audioData, format, fileName)
	if err != nil {
		logger.Error("Upload audio to MinIO for async task failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("上传音频文件失败: "+err.Error()))
		return
	}

	taskId, requestId, err := aliyun.GetTingwuClient().SubmitTask(
		fileUrl,
		fileName,
		format,
		false,
		2,
		"",
		"",
	)
	if err != nil {
		logger.Error("Submit async transcribe task failed",
			zap.String("fileUrl", fileUrl),
			logger.Error(err),
		)
		c.JSON(http.StatusInternalServerError, response.ServerError("提交转写任务失败: "+err.Error()))
		return
	}

	task := &model.VoiceTranscribeTask{
		TaskID:       taskId,
		CreatedBy:    createdBy,
		FileName:     fileName,
		FileURL:      fileUrl,
		FileSize:     int64(len(audioData)),
		Format:       format,
		Status:       model.TranscribeStatusProcessing,
		TaskType:     2,
		SpeakerCount: 2,
		RequestID:    requestId,
	}

	db := database.GetDB()
	if err := db.Create(task).Error; err != nil {
		logger.Error("Create transcribe task record failed",
			zap.String("taskId", taskId),
			logger.Error(err),
		)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"taskId":  taskId,
		"status":  "processing",
		"sync":    false,
		"message": "音频较长，已提交异步转写任务，请使用taskId查询结果",
	}))
}

func submitAsyncTaskByURL(c *app.RequestContext, fileUrl, fileName, format string) {
	var createdBy int64
	if userInfo := middleware.GetUserInfo(c); userInfo != nil {
		createdBy = userInfo.UserID
	}

	if format == "" && fileName != "" {
		format = strings.ToLower(strings.TrimPrefix(filepath.Ext(fileName), "."))
	}
	if format == "" {
		format = "mp3"
	}

	taskId, requestId, err := aliyun.GetTingwuClient().SubmitTask(
		fileUrl,
		fileName,
		format,
		false,
		2,
		"",
		"",
	)
	if err != nil {
		logger.Error("Submit async transcribe task failed",
			zap.String("fileUrl", fileUrl),
			logger.Error(err),
		)
		c.JSON(http.StatusInternalServerError, response.ServerError("提交转写任务失败: "+err.Error()))
		return
	}

	task := &model.VoiceTranscribeTask{
		TaskID:       taskId,
		CreatedBy:    createdBy,
		FileName:     fileName,
		FileURL:      fileUrl,
		Format:       format,
		Status:       model.TranscribeStatusProcessing,
		TaskType:     2,
		SpeakerCount: 2,
		RequestID:    requestId,
	}

	db := database.GetDB()
	if err := db.Create(task).Error; err != nil {
		logger.Error("Create transcribe task record failed",
			zap.String("taskId", taskId),
			logger.Error(err),
		)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"taskId":  taskId,
		"status":  "processing",
		"sync":    false,
		"message": "音频较长，已提交异步转写任务，请使用taskId查询结果",
	}))
}

func CancelTranscribeTask(ctx context.Context, c *app.RequestContext) {
	taskIdStr := c.Param("taskId")
	if taskIdStr == "" {
		c.JSON(http.StatusBadRequest, response.BadRequest("任务ID不能为空"))
		return
	}

	var taskID int64
	var err error
	if taskID, err = strconv.ParseInt(taskIdStr, 10, 64); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest("任务ID格式错误"))
		return
	}

	db := database.GetDB()
	var task model.VoiceTranscribeTask
	if err := db.Where("id = ?", taskID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("转写任务不存在"))
		return
	}

	if task.Status == model.TranscribeStatusCompleted ||
		task.Status == model.TranscribeStatusFailed ||
		task.Status == model.TranscribeStatusCanceled {
		c.JSON(http.StatusBadRequest, response.BadRequest("任务已结束，无法取消"))
		return
	}

	success, err := aliyun.GetTingwuClient().CancelTask(task.TaskID)
	if err != nil {
		logger.Warn("Cancel task from tingwu failed",
			zap.String("taskId", task.TaskID),
			logger.Error(err),
		)
	}

	if success {
		task.Status = model.TranscribeStatusCanceled
		db.Save(&task)
	}

	c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"success": success,
		"taskId":  task.TaskID,
	}))
}
