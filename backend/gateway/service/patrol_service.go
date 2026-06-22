package service

import "context"

type PatrolService interface {
	CreateTask(ctx context.Context, req map[string]interface{}, assignerID int64, assignerName string) (int64, error)
	GetTaskList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error)
	GetTaskDetail(ctx context.Context, taskID int64) (map[string]interface{}, error)
	UpdateTask(ctx context.Context, taskID int64, req map[string]interface{}) error
	DeleteTask(ctx context.Context, taskID int64) error
	CancelTask(ctx context.Context, taskID int64, reason string) error
	StartTask(ctx context.Context, taskID int64, memberID int64) error
	CompleteTask(ctx context.Context, taskID int64, memberID int64) error
	PlanRoute(ctx context.Context, startLng, startLat float64, points []map[string]interface{}, strategy int) (map[string]interface{}, error)
	GetMemberTasks(ctx context.Context, memberID int64, status int, page, pageSize int) ([]map[string]interface{}, int64, error)
	GetTaskPoints(ctx context.Context, taskID int64) ([]map[string]interface{}, error)
	Checkin(ctx context.Context, req map[string]interface{}, memberID int64, memberName string, ipAddress string) (map[string]interface{}, error)
	GetCheckinRecords(ctx context.Context, memberID int64, page, pageSize int) ([]map[string]interface{}, int64, error)
	GetCheckinStatistics(ctx context.Context, memberID int64) (map[string]interface{}, error)
	CreateVisitRecord(ctx context.Context, req map[string]interface{}, memberID int64, memberName string, orgID int64) (int64, error)
	GetVisitRecords(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error)
	GetVisitRecordDetail(ctx context.Context, id int64) (map[string]interface{}, error)
	UpdateVisitRecord(ctx context.Context, id int64, req map[string]interface{}) error
	AuditVisitRecord(ctx context.Context, id int64, status int32, remark string, auditorID int64) error
	DeleteVisitRecord(ctx context.Context, id int64) error
	GetVisitStatistics(ctx context.Context, memberID int64) (map[string]interface{}, error)
	ReportDanger(ctx context.Context, req map[string]interface{}, reporterID int64, reporterName string, orgID int64) (int64, error)
	GetDangerList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error)
	GetDangerDetail(ctx context.Context, id int64) (map[string]interface{}, error)
	HandleDanger(ctx context.Context, id int64, status int32, handlerID int64, handlerName string, result string) error
	GetDangerStatistics(ctx context.Context) (map[string]interface{}, error)
	GetMemberList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error)
	GetMemberDetail(ctx context.Context, id int64) (map[string]interface{}, error)
	GetMemberByUserID(ctx context.Context, userID int64) (map[string]interface{}, error)
	CreateMember(ctx context.Context, req map[string]interface{}) (int64, error)
	UpdateMember(ctx context.Context, id int64, req map[string]interface{}) error
	DeleteMember(ctx context.Context, id int64) error
}
