package notification

type NotificationRecord struct {
	Id           int64             `json:"id" th:"1,optional"`
	ReceiverId   int64             `json:"receiver_id" th:"2,optional"`
	ReceiverName string            `json:"receiver_name" th:"3,optional"`
	TemplateId   int64             `json:"template_id" th:"4,optional"`
	TemplateName string            `json:"template_name" th:"5,optional"`
	TemplateType int32             `json:"template_type" th:"6,optional"`
	Title        string            `json:"title" th:"7,optional"`
	Content      string            `json:"content" th:"8,optional"`
	Channel      int32             `json:"channel" th:"9,optional"`
	Status       int32             `json:"status" th:"10,optional"`
	ReadTime     string            `json:"read_time" th:"11,optional"`
	Params       map[string]string `json:"params" th:"12,optional"`
	CreatedAt    string            `json:"created_at" th:"13,optional"`
}

type NotificationTemplate struct {
	Id              int64  `json:"id" th:"1,optional"`
	TemplateName    string `json:"template_name" th:"2,optional"`
	TemplateCode    string `json:"template_code" th:"3,optional"`
	TemplateType    int32  `json:"template_type" th:"4,optional"`
	TitleTemplate   string `json:"title_template" th:"5,optional"`
	ContentTemplate string `json:"content_template" th:"6,optional"`
	Channel         int32  `json:"channel" th:"7,optional"`
	Status          int32  `json:"status" th:"8,optional"`
	CreatedAt       string `json:"created_at" th:"9,optional"`
}

type GetNotificationsRequest struct {
	UserId   int64  `json:"user_id" th:"1,optional"`
	Page     int32  `json:"page" th:"2,optional"`
	PageSize int32  `json:"page_size" th:"3,optional"`
	Type     int32  `json:"type" th:"4,optional"`
	Status   int32  `json:"status" th:"5,optional"`
	IsRead   bool   `json:"is_read" th:"6,optional"`
	Keyword  string `json:"keyword" th:"7,optional"`
}

type GetNotificationsResponse struct {
	Code        int32                 `json:"code" th:"1,optional"`
	Message     string                `json:"message" th:"2,optional"`
	Records     []*NotificationRecord `json:"records" th:"3,optional"`
	Total       int64                 `json:"total" th:"4,optional"`
	UnreadCount int64                 `json:"unread_count" th:"5,optional"`
}

type SendNotificationRequest struct {
	ReceiverIds []int64           `json:"receiver_ids" th:"1,optional"`
	TemplateId  int64             `json:"template_id" th:"2,optional"`
	Params      map[string]string `json:"params" th:"3,optional"`
	NotifyType  string            `json:"notify_type" th:"4,optional"`
	SenderId    int64             `json:"sender_id" th:"5,optional"`
}

type SendNotificationResponse struct {
	Code         int32  `json:"code" th:"1,optional"`
	Message      string `json:"message" th:"2,optional"`
	SuccessCount int32  `json:"success_count" th:"3,optional"`
}

type MarkAsReadRequest struct {
	Id     int64 `json:"id" th:"1,optional"`
	UserId int64 `json:"user_id" th:"2,optional"`
}

type MarkAsReadResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type MarkAllAsReadRequest struct {
	UserId int64 `json:"user_id" th:"1,optional"`
}

type MarkAllAsReadResponse struct {
	Code        int32  `json:"code" th:"1,optional"`
	Message     string `json:"message" th:"2,optional"`
	MarkedCount int64  `json:"marked_count" th:"3,optional"`
}

type GetUnreadCountRequest struct {
	UserId int64 `json:"user_id" th:"1,optional"`
}

type GetUnreadCountResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
	Count   int64  `json:"count" th:"3,optional"`
}

type GetTemplatesRequest struct {
	TemplateType int32 `json:"template_type" th:"1,optional"`
}

type GetTemplatesResponse struct {
	Code      int32                  `json:"code" th:"1,optional"`
	Message   string                 `json:"message" th:"2,optional"`
	Templates []*NotificationTemplate `json:"templates" th:"3,optional"`
}

type DeleteNotificationRequest struct {
	Id     int64 `json:"id" th:"1,optional"`
	UserId int64 `json:"user_id" th:"2,optional"`
}

type DeleteNotificationResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type BatchDeleteRequest struct {
	Ids    []int64 `json:"ids" th:"1,optional"`
	UserId int64   `json:"user_id" th:"2,optional"`
}

type BatchDeleteResponse struct {
	Code         int32  `json:"code" th:"1,optional"`
	Message      string `json:"message" th:"2,optional"`
	DeletedCount int64  `json:"deleted_count" th:"3,optional"`
}

type SendByMQRequest struct {
	TemplateCode string            `json:"template_code" th:"1,optional"`
	ReceiverIds  []int64           `json:"receiver_ids" th:"2,optional"`
	Params       map[string]string `json:"params" th:"3,optional"`
}

type SendByMQResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}
