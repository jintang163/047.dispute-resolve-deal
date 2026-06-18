namespace go notification

struct NotificationRecord {
    1: i64 id
    2: i64 receiver_id
    3: string receiver_name
    4: i64 template_id
    5: string template_name
    6: i32 template_type
    7: string title
    8: string content
    9: i32 channel
    10: i32 status
    11: string read_time
    12: map<string, string> params
    13: string created_at
}

struct NotificationTemplate {
    1: i64 id
    2: string template_name
    3: string template_code
    4: i32 template_type
    5: string title_template
    6: string content_template
    7: i32 channel
    8: i32 status
    9: string created_at
}

struct GetNotificationsRequest {
    1: i64 user_id
    2: i32 page
    3: i32 page_size
    4: i32 type
    5: i32 status
    6: bool is_read
    7: string keyword
}

struct GetNotificationsResponse {
    1: i32 code
    2: string message
    3: list<NotificationRecord> records
    4: i64 total
    5: i64 unread_count
}

struct SendNotificationRequest {
    1: list<i64> receiver_ids
    2: i64 template_id
    3: map<string, string> params
    4: string notify_type
    5: i64 sender_id
}

struct SendNotificationResponse {
    1: i32 code
    2: string message
    3: i32 success_count
}

struct MarkAsReadRequest {
    1: i64 id
    2: i64 user_id
}

struct MarkAsReadResponse {
    1: i32 code
    2: string message
}

struct MarkAllAsReadRequest {
    1: i64 user_id
}

struct MarkAllAsReadResponse {
    1: i32 code
    2: string message
    3: i64 marked_count
}

struct GetUnreadCountRequest {
    1: i64 user_id
}

struct GetUnreadCountResponse {
    1: i32 code
    2: string message
    3: i64 count
}

struct GetTemplatesRequest {
    1: i32 template_type
}

struct GetTemplatesResponse {
    1: i32 code
    2: string message
    3: list<NotificationTemplate> templates
}

struct DeleteNotificationRequest {
    1: i64 id
    2: i64 user_id
}

struct DeleteNotificationResponse {
    1: i32 code
    2: string message
}

struct BatchDeleteRequest {
    1: list<i64> ids
    2: i64 user_id
}

struct BatchDeleteResponse {
    1: i32 code
    2: string message
    3: i64 deleted_count
}

struct SendByMQRequest {
    1: string template_code
    2: list<i64> receiver_ids
    3: map<string, string> params
}

struct SendByMQResponse {
    1: i32 code
    2: string message
}

service NotificationService {
    GetNotificationsResponse GetMyNotifications(1: GetNotificationsRequest request)
    GetNotificationsResponse GetNotificationDetail(1: i64 id, 2: i64 user_id)
    MarkAsReadResponse MarkAsRead(1: MarkAsReadRequest request)
    MarkAllAsReadResponse MarkAllAsRead(1: MarkAllAsReadRequest request)
    SendNotificationResponse SendNotification(1: SendNotificationRequest request)
    GetTemplatesResponse GetNotificationTemplates(1: GetTemplatesRequest request)
    GetUnreadCountResponse GetUnreadCount(1: GetUnreadCountRequest request)
    DeleteNotificationResponse DeleteNotification(1: DeleteNotificationRequest request)
    BatchDeleteResponse BatchDeleteNotifications(1: BatchDeleteRequest request)
    SendByMQResponse SendNotificationByMQ(1: SendByMQRequest request)
}
