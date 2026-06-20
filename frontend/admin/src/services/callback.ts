import { request } from '../utils/request';

export interface CallbackRecord {
  id: string;
  caseId: string;
  caseNo: string;
  caseTitle: string;
  applicantId: string;
  applicantName: string;
  applicantPhone: string;
  taskId?: string;
  callId?: string;
  status: number;
  statusName?: string;
  callStatus: number;
  callStatusName?: string;
  callTime?: string;
  callDuration?: number;
  retryCount: number;
  maxRetryCount: number;
  nextRetryTime?: string;
  scheduledTime?: string;
  transcriptText?: string;
  sentimentResult?: string;
  sentimentScore?: number;
  emotion?: string;
  emotionName?: string;
  performanceScore?: number;
  satisfactionScore?: number;
  recordingUrl?: string;
  recordingSize?: number;
  expireAt?: string;
  resultData?: any;
  remark?: string;
  createdAt: string;
  updatedAt: string;
}

export interface SentimentDetail {
  emotion: string;
  emotionLabel: string;
  sentimentScore: number;
  confidence: number;
  positiveKeywords: string[];
  negativeKeywords: string[];
  keyPoints: Array<{
    content: string;
    sentiment: string;
    score: number;
  }>;
  satisfaction: number;
  performance: number;
  summary: string;
}

export interface CallbackListParams {
  page?: number;
  pageSize?: number;
  caseId?: string;
  status?: number;
  callStatus?: number;
  keyword?: string;
  startTime?: string;
  endTime?: string;
}

export interface CallbackListResponse {
  list: CallbackRecord[];
  total: number;
  page: number;
  pageSize: number;
}

export const callbackService = {
  getList: (params?: CallbackListParams) => {
    return request.get<CallbackListResponse>('/callback', { params });
  },

  getDetail: (id: string) => {
    return request.get<CallbackRecord>(`/callback/${id}`);
  },

  create: (caseId: string) => {
    return request.post<CallbackRecord>('/callback', { caseId });
  },

  initiate: (id: string) => {
    return request.post(`/callback/${id}/initiate`);
  },

  retry: (id: string) => {
    return request.post(`/callback/${id}/retry`);
  },

  cancel: (id: string) => {
    return request.post(`/callback/${id}/cancel`);
  },

  refresh: (id: string) => {
    return request.post(`/callback/${id}/refresh`);
  },

  archiveRecording: (id: string) => {
    return request.post(`/callback/${id}/archive-recording`);
  },

  getByCase: (caseId: string) => {
    return request.get<CallbackRecord[]>(`/callback/case/${caseId}`);
  },
};

export const CallbackStatusEnum = {
  PENDING: 10,
  CALLING: 20,
  SUCCESS: 30,
  FAILED: 40,
  CANCELLED: 99,
};

export const CallbackStatusMap: Record<number, { text: string; color: string }> = {
  10: { text: '待回访', color: 'default' },
  20: { text: '回访中', color: 'processing' },
  30: { text: '回访成功', color: 'success' },
  40: { text: '回访失败', color: 'error' },
  99: { text: '已取消', color: 'default' },
};

export const CallStatusMap: Record<number, { text: string; color: string }> = {
  0: { text: '未呼叫', color: 'default' },
  10: { text: '振铃中', color: 'processing' },
  20: { text: '已接听', color: 'success' },
  30: { text: '无人接听', color: 'warning' },
  40: { text: '用户忙', color: 'warning' },
  50: { text: '呼叫失败', color: 'error' },
  60: { text: '已挂断', color: 'default' },
};

export const EmotionMap: Record<string, { text: string; color: string }> = {
  positive: { text: '正面', color: 'success' },
  neutral: { text: '中性', color: 'default' },
  negative: { text: '负面', color: 'error' },
};
