package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dispute-resolve/common/auth"
	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/response"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/middleware"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CaseExportRequest struct {
	TypeID     int64   `json:"typeId"`
	MediatorID int64   `json:"mediatorId"`
	Status     int32   `json:"status"`
	CaseLevel  int32   `json:"caseLevel"`
	StartTime  string  `json:"startTime"`
	EndTime    string  `json:"endTime"`
	Keyword    string  `json:"keyword"`
	TagKeyword string  `json:"tagKeyword"`
	IDs        []int64 `json:"ids"`
}

type ExportListRequest struct {
	model.BaseQuery
	ExportType   int    `form:"exportType"`
	ExportStatus int    `form:"exportStatus"`
	model.DateRangeQuery
}

func CreateCaseExport(ctx context.Context, c *app.RequestContext) {
	var req CaseExportRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
		return
	}

	ip := c.ClientIP()
	userAgent := string(c.GetHeader("User-Agent"))

	if req.StartTime != "" && req.EndTime != "" {
		start, err1 := time.Parse("2006-01-02", req.StartTime)
		end, err2 := time.Parse("2006-01-02", req.EndTime)
		if err1 == nil && err2 == nil {
			if end.Sub(start).Hours() > 24*366 {
				c.JSON(http.StatusBadRequest, response.BadRequest("时间范围不能超过1年"))
				return
			}
		}
	}

	var operatorPhone string
	database.GetDB().Table("sys_user").
		Select("phone").
		Where("id = ?", userInfo.UserID).
		Scan(&operatorPhone)

	db := buildDisputeExportQuery(&req, userInfo)

	var cases []map[string]interface{}
	if err := db.Find(&cases).Error; err != nil {
		logger.Error("Query dispute cases for export failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("查询案件数据失败"))
		return
	}

	recordCount := len(cases)
	if recordCount == 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("没有符合条件的数据可导出"))
		return
	}

	csvBytes := generateCaseCSV(cases)

	csvFileName := fmt.Sprintf("cases_%s.csv", time.Now().Format("20060102150405"))
	zipBytes, err := zipCSV(csvBytes, csvFileName)
	if err != nil {
		logger.Error("Zip CSV failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("压缩文件失败"))
		return
	}

	aesKeyHex, err := utils.GenerateAES256Key()
	if err != nil {
		logger.Error("Generate AES key failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("生成加密密钥失败"))
		return
	}

	exportPassword := utils.GenerateExportPassword(16)

	encryptedBytes, err := utils.AESEncryptGCM(zipBytes, aesKeyHex)
	if err != nil {
		logger.Error("AES encrypt failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("加密文件失败"))
		return
	}

	now := time.Now()
	objectPath := fmt.Sprintf("%s/%s/EXP%d_%s.enc",
		constants.MinIOPathExport,
		now.Format("200601"),
		now.Unix(),
		utils.GenerateRandomString(6),
	)

	fileSize, err := uploadToMinIO(encryptedBytes, objectPath)
	if err != nil {
		logger.Error("Upload to MinIO failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("上传文件失败"))
		return
	}

	exportNo := "EXP" + now.Format("20060102") + utils.GenerateRandomNumber(6)
	exportID := utils.GenerateID()
	expiredAt := now.AddDate(0, 0, constants.ExportExpireDays)

	filterConditionsJSON, _ := json.Marshal(req)

	exportLog := &model.DataExportLog{
		BaseModel: model.BaseModel{
			ID:        exportID,
			CreatedAt: now,
			UpdatedAt: now,
		},
		ExportNo:            exportNo,
		ExportType:          constants.ExportTypeCase,
		ExportName:          fmt.Sprintf("案件数据导出_%s", now.Format("2006-01-02 15:04:05")),
		FilterConditions:    string(filterConditionsJSON),
		RecordCount:         recordCount,
		FileName:            exportNo + ".enc",
		FilePath:            objectPath,
		FileSize:            fileSize,
		EncryptionAlgorithm: "AES-256-GCM",
		PasswordSmsSent:     constants.ExportPasswordSmsStatusPending,
		ExportStatus:        constants.ExportStatusSuccess,
		OperatorID:          userInfo.UserID,
		OperatorName:        userInfo.RealName,
		OperatorPhone:       operatorPhone,
		OrgID:               userInfo.OrganizationID,
		IPAddress:           ip,
		UserAgent:           userAgent,
		CompletedAt:         &now,
		ExpiredAt:           expiredAt,
	}

	if err := database.GetDB().Create(exportLog).Error; err != nil {
		logger.Error("Create data export log failed", logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("保存导出记录失败"))
		return
	}

	operationLog := map[string]interface{}{
		"operation_type":   "DATA_EXPORT",
		"operation_detail": fmt.Sprintf("导出案件数据 %d 条，导出单号: %s", recordCount, exportNo),
		"operator_id":      userInfo.UserID,
		"operator_name":    userInfo.RealName,
		"operator_role":    userInfo.Role,
		"ip_address":       ip,
		"user_agent":       userAgent,
		"biz_id":           exportID,
		"biz_type":         "data_export",
		"created_at":       now,
	}
	go database.GetDB().Table("operation_log").Create(operationLog)

	go sendExportPasswordMQ(exportID, exportNo, exportLog.ExportName, exportPassword, operatorPhone, userInfo.RealName, recordCount)

	logger.Info("Case export created",
		zap.Int64("exportId", exportID),
		zap.String("exportNo", exportNo),
		zap.Int("recordCount", recordCount),
		zap.Int64("operatorId", userInfo.UserID),
	)

	c.JSON(http.StatusOK, response.SuccessWithMessage(map[string]interface{}{
		"exportId":    exportID,
		"exportNo":    exportNo,
		"recordCount": recordCount,
	}, "导出任务已提交，密码将通过短信发送到您的手机"))
}

func GetExportList(ctx context.Context, c *app.RequestContext) {
	var req ExportListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.BadRequest(err.Error()))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
		return
	}

	db := database.GetDB().Table("data_export_log del").
		Select("del.*, su.real_name as operator_name, so.org_name").
		Joins("LEFT JOIN sys_user su ON del.operator_id = su.id").
		Joins("LEFT JOIN sys_organization so ON del.org_id = so.id").
		Where("del.deleted_at IS NULL")

	switch userInfo.Role {
	case constants.RoleMediator:
		db = db.Where("del.operator_id = ?", userInfo.UserID)
	case constants.RoleLeader:
		db = db.Where("del.org_id IN (SELECT id FROM sys_organization WHERE parent_id = ? OR id = ?)",
			userInfo.OrganizationID, userInfo.OrganizationID)
	case constants.RoleDirector:
		db = db.Where("del.org_id IN (SELECT id FROM sys_organization WHERE parent_id = ? OR id = ?)",
			userInfo.OrganizationID, userInfo.OrganizationID)
	case constants.RoleAdmin:
	default:
		db = db.Where("del.operator_id = ?", userInfo.UserID)
	}

	if req.ExportType > 0 {
		db = db.Where("del.export_type = ?", req.ExportType)
	}
	if req.ExportStatus > 0 {
		db = db.Where("del.export_status = ?", req.ExportStatus)
	}
	if req.StartTime != "" {
		db = db.Where("del.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("del.created_at <= ?", req.EndTime)
	}
	if req.Keyword != "" {
		db = db.Where("del.export_no LIKE ? OR del.export_name LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	var total int64
	db.Count(&total)

	var list []map[string]interface{}
	db.Order("del.created_at DESC").
		Offset(req.GetOffset()).
		Limit(req.GetLimit()).
		Find(&list)

	for _, item := range list {
		if exportType, ok := item["export_type"].(int); ok {
			item["export_type_name"] = constants.ExportTypeMap[exportType]
		} else if exportType, ok := item["export_type"].(int64); ok {
			item["export_type_name"] = constants.ExportTypeMap[int(exportType)]
		}
		if exportStatus, ok := item["export_status"].(int); ok {
			item["export_status_name"] = constants.ExportStatusMap[exportStatus]
		} else if exportStatus, ok := item["export_status"].(int64); ok {
			item["export_status_name"] = constants.ExportStatusMap[int(exportStatus)]
		}
		if smsStatus, ok := item["password_sms_sent"].(int); ok {
			item["password_sms_sent_name"] = constants.ExportPasswordSmsStatusMap[smsStatus]
		} else if smsStatus, ok := item["password_sms_sent"].(int64); ok {
			item["password_sms_sent_name"] = constants.ExportPasswordSmsStatusMap[int(smsStatus)]
		}
		delete(item, "operator_phone")
	}

	c.JSON(http.StatusOK, response.Page(list, total, req.Page, req.PageSize))
}

func GetExportDetail(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的导出记录ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
		return
	}

	var detail map[string]interface{}
	db := database.GetDB().Table("data_export_log del").
		Select("del.*, su.real_name as operator_name, so.org_name").
		Joins("LEFT JOIN sys_user su ON del.operator_id = su.id").
		Joins("LEFT JOIN sys_organization so ON del.org_id = so.id").
		Where("del.id = ? AND del.deleted_at IS NULL", id)

	switch userInfo.Role {
	case constants.RoleMediator:
		db = db.Where("del.operator_id = ?", userInfo.UserID)
	case constants.RoleLeader, constants.RoleDirector:
		db = db.Where("del.org_id IN (SELECT id FROM sys_organization WHERE parent_id = ? OR id = ?)",
			userInfo.OrganizationID, userInfo.OrganizationID)
	case constants.RoleAdmin:
	default:
		db = db.Where("del.operator_id = ?", userInfo.UserID)
	}

	db.Find(&detail)

	if detail == nil {
		c.JSON(http.StatusNotFound, response.NotFound("导出记录不存在"))
		return
	}

	if exportType, ok := detail["export_type"].(int); ok {
		detail["export_type_name"] = constants.ExportTypeMap[exportType]
	} else if exportType, ok := detail["export_type"].(int64); ok {
		detail["export_type_name"] = constants.ExportTypeMap[int(exportType)]
	}
	if exportStatus, ok := detail["export_status"].(int); ok {
		detail["export_status_name"] = constants.ExportStatusMap[exportStatus]
	} else if exportStatus, ok := detail["export_status"].(int64); ok {
		detail["export_status_name"] = constants.ExportStatusMap[int(exportStatus)]
	}
	if smsStatus, ok := detail["password_sms_sent"].(int); ok {
		detail["password_sms_sent_name"] = constants.ExportPasswordSmsStatusMap[smsStatus]
	} else if smsStatus, ok := detail["password_sms_sent"].(int64); ok {
		detail["password_sms_sent_name"] = constants.ExportPasswordSmsStatusMap[int(smsStatus)]
	}
	delete(detail, "operator_phone")

	c.JSON(http.StatusOK, response.Success(detail))
}

func DownloadExport(ctx context.Context, c *app.RequestContext) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		c.JSON(http.StatusBadRequest, response.BadRequest("无效的导出记录ID"))
		return
	}

	userInfo := middleware.GetUserInfo(c)
	if userInfo == nil {
		c.JSON(http.StatusUnauthorized, response.Unauthorized(""))
		return
	}

	var exportLog model.DataExportLog
	db := database.GetDB().Where("id = ? AND deleted_at IS NULL", id)

	switch userInfo.Role {
	case constants.RoleMediator:
		db = db.Where("operator_id = ?", userInfo.UserID)
	case constants.RoleLeader, constants.RoleDirector:
		db = db.Where("org_id IN (SELECT id FROM sys_organization WHERE parent_id = ? OR id = ?)",
			userInfo.OrganizationID, userInfo.OrganizationID)
	case constants.RoleAdmin:
	default:
		db = db.Where("operator_id = ?", userInfo.UserID)
	}

	if err := db.First(&exportLog).Error; err != nil {
		c.JSON(http.StatusNotFound, response.NotFound("导出记录不存在或无权限下载"))
		return
	}

	if exportLog.ExportStatus != constants.ExportStatusSuccess {
		c.JSON(http.StatusBadRequest, response.BadRequest("导出文件尚未准备好"))
		return
	}

	if time.Now().After(exportLog.ExpiredAt) {
		c.JSON(http.StatusBadRequest, response.BadRequest("导出文件已过期"))
		return
	}

	cfg := config.GetConfig().MinIO
	minioClient := database.GetMinioClient()
	object, err := minioClient.GetObject(ctx, cfg.Bucket, exportLog.FilePath, minio.GetObjectOptions{})
	if err != nil {
		logger.Error("Get object from MinIO failed",
			zap.String("filePath", exportLog.FilePath),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("获取文件失败"))
		return
	}
	defer object.Close()

	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(object); err != nil {
		logger.Error("Read object from MinIO failed",
			zap.String("filePath", exportLog.FilePath),
			logger.Error(err))
		c.JSON(http.StatusInternalServerError, response.ServerError("读取文件失败"))
		return
	}

	fileName := exportLog.ExportNo + ".enc"
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Header("Content-Length", strconv.FormatInt(exportLog.FileSize, 10))

	c.Data(http.StatusOK, "application/octet-stream", buf.Bytes())
}

func buildDisputeExportQuery(req *CaseExportRequest, userInfo *auth.UserInfo) *gorm.DB {
	db := database.GetDB().Table("dispute_case dc").
		Select("dc.*, dt.type_name, dt.level_path as type_path, su.real_name as mediator_name, so.org_name").
		Joins("LEFT JOIN dispute_type dt ON dc.type_id = dt.id").
		Joins("LEFT JOIN sys_user su ON dc.mediator_id = su.id").
		Joins("LEFT JOIN sys_organization so ON dc.organization_id = so.id").
		Where("dc.deleted_at IS NULL")

	if userInfo != nil {
		if userInfo.Role == constants.RoleMediator {
			db = db.Where("dc.mediator_id = ?", userInfo.UserID)
		} else if userInfo.Role == constants.RoleLeader || userInfo.Role == constants.RoleDirector {
			db = db.Where("dc.organization_id IN (SELECT id FROM sys_organization WHERE parent_id = ? OR id = ?)",
				userInfo.OrganizationID, userInfo.OrganizationID)
		}
	}

	if len(req.IDs) > 0 {
		db = db.Where("dc.id IN ?", req.IDs)
	}
	if req.Status > 0 {
		db = db.Where("dc.status = ?", req.Status)
	}
	if req.CaseLevel > 0 {
		db = db.Where("dc.case_level = ?", req.CaseLevel)
	}
	if req.TypeID > 0 {
		db = db.Where("dc.type_id = ?", req.TypeID)
	}
	if req.MediatorID > 0 {
		db = db.Where("dc.mediator_id = ?", req.MediatorID)
	}
	if req.Keyword != "" {
		db = db.Where("dc.title LIKE ? OR dc.description LIKE ? OR dc.case_no LIKE ?",
			"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.TagKeyword != "" {
		db = db.Where("JSON_CONTAINS(dc.keywords, ?)", fmt.Sprintf(`"%s"`, req.TagKeyword))
	}
	if req.StartTime != "" {
		db = db.Where("dc.created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("dc.created_at <= ?", req.EndTime)
	}

	return db
}

func generateCaseCSV(cases []map[string]interface{}) []byte {
	headers := []string{
		"案件编号", "标题", "纠纷类型", "紧急程度", "状态",
		"报案人", "报案电话", "被申请人", "调解员", "所属机构",
		"创建时间", "结案时间", "调解结果", "满意度",
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Write(headers)

	for _, c := range cases {
		row := make([]string, 14)

		row[0] = getMapStringValue(c, "case_no")
		row[1] = getMapStringValue(c, "title")
		row[2] = getMapStringValue(c, "type_name")

		caseLevel := getMapIntValue(c, "case_level")
		row[3] = constants.CaseLevelMap[caseLevel]

		status := getMapIntValue(c, "status")
		row[4] = constants.CaseStatusMap[status]

		row[5] = getMapStringValue(c, "reporter_name")
		row[6] = getMapStringValue(c, "reporter_phone")
		row[7] = getMapStringValue(c, "respondent_name")
		row[8] = getMapStringValue(c, "mediator_name")
		row[9] = getMapStringValue(c, "org_name")
		row[10] = getMapTimeValue(c, "created_at")
		row[11] = getMapTimeValue(c, "closed_at")

		mediationResult := getMapIntValue(c, "mediation_result")
		mediationResultMap := map[int]string{
			constants.MediationResultPending: "待调解",
			constants.MediationResultSuccess: "调解成功",
			constants.MediationResultFail:    "调解失败",
			constants.MediationResultPartial: "部分达成",
		}
		if name, ok := mediationResultMap[mediationResult]; ok {
			row[12] = name
		} else {
			row[12] = ""
		}

		satisfactionScore := getMapIntValue(c, "satisfaction_score")
		if satisfactionScore > 0 {
			row[13] = fmt.Sprintf("%d分", satisfactionScore)
		} else {
			row[13] = ""
		}

		writer.Write(row)
	}

	writer.Flush()
	return buf.Bytes()
}

func getMapStringValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case string:
			return val
		case []byte:
			return string(val)
		default:
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

func getMapIntValue(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case int:
			return val
		case int32:
			return int(val)
		case int64:
			return int(val)
		case float64:
			return int(val)
		}
	}
	return 0
}

func getMapTimeValue(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		switch val := v.(type) {
		case time.Time:
			if !val.IsZero() {
				return val.Format("2006-01-02 15:04:05")
			}
		case *time.Time:
			if val != nil && !val.IsZero() {
				return val.Format("2006-01-02 15:04:05")
			}
		case string:
			if val != "" && val != "null" {
				return strings.Split(val, ".")[0]
			}
		}
	}
	return ""
}

func zipCSV(csvBytes []byte, fileName string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	writer, err := zipWriter.Create(fileName)
	if err != nil {
		return nil, fmt.Errorf("create zip entry failed: %w", err)
	}

	if _, err := writer.Write(csvBytes); err != nil {
		return nil, fmt.Errorf("write to zip failed: %w", err)
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("close zip writer failed: %w", err)
	}

	return buf.Bytes(), nil
}

func uploadToMinIO(encryptedBytes []byte, objectPath string) (int64, error) {
	cfg := config.GetConfig().MinIO
	minioClient := database.GetMinioClient()

	reader := bytes.NewReader(encryptedBytes)
	uploadInfo, err := minioClient.PutObject(
		context.Background(),
		cfg.Bucket,
		objectPath,
		reader,
		int64(len(encryptedBytes)),
		minio.PutObjectOptions{
			ContentType: "application/octet-stream",
		},
	)
	if err != nil {
		return 0, err
	}

	return uploadInfo.Size, nil
}

func sendExportPasswordMQ(exportID int64, exportNo, exportName, password, operatorPhone, operatorName string, recordCount int) {
	msg := map[string]interface{}{
		"exportId":      exportID,
		"exportNo":      exportNo,
		"exportName":    exportName,
		"password":      password,
		"operatorPhone": operatorPhone,
		"operatorName":  operatorName,
		"recordCount":   recordCount,
	}

	callback := func(ctx context.Context, result *primitive.SendResult, err error) {
		if err != nil {
			logger.Error("Send export password MQ failed",
				zap.Int64("exportId", exportID),
				zap.String("exportNo", exportNo),
				logger.Error(err),
			)
			now := time.Now()
			database.GetDB().Table("data_export_log").
				Where("id = ?", exportID).
				Updates(map[string]interface{}{
					"password_sms_sent": constants.ExportPasswordSmsStatusFailed,
					"updated_at":        now,
				})
		} else {
			logger.Info("Send export password MQ success",
				zap.Int64("exportId", exportID),
				zap.String("exportNo", exportNo),
				zap.String("msgId", result.MsgID),
			)
			now := time.Now()
			database.GetDB().Table("data_export_log").
				Where("id = ?", exportID).
				Updates(map[string]interface{}{
					"password_sms_sent": constants.ExportPasswordSmsStatusSent,
					"password_sms_time": now,
					"updated_at":        now,
				})
		}
	}

	if err := mq.SendAsyncMessage(constants.MQTopicExportPassword, msg, callback, constants.MQTagSms); err != nil {
		logger.Error("Send export password MQ async call failed",
			zap.Int64("exportId", exportID),
			zap.String("exportNo", exportNo),
			logger.Error(err),
		)
	}
}
