import { request } from '../utils/request';

export interface Notification {
  id: number;
  title: string;
  content: string;
  templateCode: string;
  templateType: number;
  templateName: string;
  channel: number;
  status: number;
  readTime?: string;
  createdAt: string;
  params?: string;
  bizType?: number;
  bizId?: number;
}

export interface NotificationListParams {
  page?: number;
  pageSize?: number;
  type?: number;
  status?: number;
  isRead?: boolean;
  keyword?: string;
}

export interface NotificationListResponse {
  list: Notification[];
  total: number;
  page: number;
  pageSize: number;
  extra: {
    unreadCount: number;
  };
}

export interface UnreadCountResponse {
  total: number;
  typeCounts: { template_type: number; count: number }[];
}

export const notificationService = {
  getMyNotifications: (params?: NotificationListParams) => {
    return request.get<NotificationListResponse>('/v1/notification', { params });
  },

  getUnreadCount: () => {
    return request.get<UnreadCountResponse>('/v1/notification/unread-count');
  },

  getNotificationDetail: (id: number) => {
    return request.get<Notification>(`/v1/notification/${id}`);
  },

  markAsRead: (id: number) => {
    return request.put(`/v1/notification/${id}/read`);
  },

  markAllAsRead: () => {
    return request.put('/v1/notification/read-all');
  },

  deleteNotification: (id: number) => {
    return request.delete(`/v1/notification/${id}`);
  },

  batchDelete: (ids: number[]) => {
    return request.post('/v1/notification/batch-delete', { ids });
  },

  getTemplates: (type?: number) => {
    return request.get('/v1/notification/templates', { params: { type } });
  },

  sendNotification: (data: { receiverIds: number[]; templateId: number; params?: Record<string, any>; notifyType?: string }) => {
    return request.post('/v1/notification/send', data);
  },
};
