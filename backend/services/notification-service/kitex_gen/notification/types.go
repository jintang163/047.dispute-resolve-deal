package notification

type NotificationRecord struct {
	Id           int64             `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	ReceiverId   int64             `thrift:"receiver_id,2" frugal:"2,default,i64" json:"receiverId"`
	ReceiverName string            `thrift:"receiver_name,3" frugal:"3,default,string" json:"receiverName"`
	TemplateId   int64             `thrift:"template_id,4" frugal:"4,default,i64" json:"templateId"`
	TemplateName string            `thrift:"template_name,5" frugal:"5,default,string" json:"templateName"`
	TemplateType int32             `thrift:"template_type,6" frugal:"6,default,i32" json:"templateType"`
	Title        string            `thrift:"title,7" frugal:"7,default,string" json:"title"`
	Content      string            `thrift:"content,8" frugal:"8,default,string" json:"content"`
	Channel      int32             `thrift:"channel,9" frugal:"9,default,i32" json:"channel"`
	Status       int32             `thrift:"status,10" frugal:"10,default,i32" json:"status"`
	ReadTime     string            `thrift:"read_time,11" frugal:"11,default,string" json:"readTime"`
	Params       map[string]string `thrift:"params,12" frugal:"12,default,map<string:string>" json:"params"`
	CreatedAt    string            `thrift:"created_at,13" frugal:"13,default,string" json:"createdAt"`
}

type NotificationTemplate struct {
	Id              int64  `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	TemplateName    string `thrift:"template_name,2" frugal:"2,default,string" json:"templateName"`
	TemplateCode    string `thrift:"template_code,3" frugal:"3,default,string" json:"templateCode"`
	TemplateType    int32  `thrift:"template_type,4" frugal:"4,default,i32" json:"templateType"`
	TitleTemplate   string `thrift:"title_template,5" frugal:"5,default,string" json:"titleTemplate"`
	ContentTemplate string `thrift:"content_template,6" frugal:"6,default,string" json:"contentTemplate"`
	Channel         int32  `thrift:"channel,7" frugal:"7,default,i32" json:"channel"`
	Status          int32  `thrift:"status,8" frugal:"8,default,i32" json:"status"`
	CreatedAt       string `thrift:"created_at,9" frugal:"9,default,string" json:"createdAt"`
}

type GetNotificationsRequest struct {
	UserId   int64  `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
	Page     int32  `thrift:"page,2" frugal:"2,default,i32" json:"page"`
	PageSize int32  `thrift:"page_size,3" frugal:"3,default,i32" json:"pageSize"`
	Type     int32  `thrift:"type,4" frugal:"4,default,i32" json:"type"`
	Status   int32  `thrift:"status,5" frugal:"5,default,i32" json:"status"`
	IsRead   bool   `thrift:"is_read,6" frugal:"6,default,bool" json:"isRead"`
	Keyword  string `thrift:"keyword,7" frugal:"7,default,string" json:"keyword"`
}

type GetNotificationsResponse struct {
	Code        int32                 `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message     string                `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Records     []*NotificationRecord `thrift:"records,3" frugal:"3,default,list<*NotificationRecord>" json:"records,omitempty"`
	Total       int64                 `thrift:"total,4" frugal:"4,default,i64" json:"total"`
	UnreadCount int64                 `thrift:"unread_count,5" frugal:"5,default,i64" json:"unreadCount"`
}

type SendNotificationRequest struct {
	ReceiverIds []int64           `thrift:"receiver_ids,1" frugal:"1,default,list<i64>" json:"receiverIds"`
	TemplateId  int64             `thrift:"template_id,2" frugal:"2,default,i64" json:"templateId"`
	Params      map[string]string `thrift:"params,3" frugal:"3,default,map<string:string>" json:"params"`
	NotifyType  string            `thrift:"notify_type,4" frugal:"4,default,string" json:"notifyType"`
	SenderId    int64             `thrift:"sender_id,5" frugal:"5,default,i64" json:"senderId"`
}

type SendNotificationResponse struct {
	Code         int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message      string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	SuccessCount int32  `thrift:"success_count,3" frugal:"3,default,i32" json:"successCount"`
}

type MarkAsReadRequest struct {
	Id     int64 `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	UserId int64 `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
}

type MarkAsReadResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type MarkAllAsReadRequest struct {
	UserId int64 `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
}

type MarkAllAsReadResponse struct {
	Code        int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message     string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	MarkedCount int64  `thrift:"marked_count,3" frugal:"3,default,i64" json:"markedCount"`
}

type GetUnreadCountRequest struct {
	UserId int64 `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
}

type GetUnreadCountResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Count   int64  `thrift:"count,3" frugal:"3,default,i64" json:"count"`
}

type GetTemplatesRequest struct {
	TemplateType int32 `thrift:"template_type,1" frugal:"1,default,i32" json:"templateType"`
}

type GetTemplatesResponse struct {
	Code      int32                   `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message   string                  `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Templates []*NotificationTemplate `thrift:"templates,3" frugal:"3,default,list<*NotificationTemplate>" json:"templates,omitempty"`
}

type DeleteNotificationRequest struct {
	Id     int64 `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	UserId int64 `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
}

type DeleteNotificationResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type BatchDeleteRequest struct {
	Ids    []int64 `thrift:"ids,1" frugal:"1,default,list<i64>" json:"ids"`
	UserId int64   `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
}

type BatchDeleteResponse struct {
	Code         int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message      string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	DeletedCount int64  `thrift:"deleted_count,3" frugal:"3,default,i64" json:"deletedCount"`
}

type SendByMQRequest struct {
	TemplateCode string            `thrift:"template_code,1" frugal:"1,default,string" json:"templateCode"`
	ReceiverIds  []int64           `thrift:"receiver_ids,2" frugal:"2,default,list<i64>" json:"receiverIds"`
	Params       map[string]string `thrift:"params,3" frugal:"3,default,map<string:string>" json:"params"`
}

type SendByMQResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}
