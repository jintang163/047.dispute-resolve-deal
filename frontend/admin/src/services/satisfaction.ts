import { request } from '../utils/request';

export interface SatisfactionSentimentStats {
  totalAnalyzed: number;
  positiveCount: number;
  neutralCount: number;
  negativeCount: number;
  positiveRate: number;
  neutralRate: number;
  negativeRate: number;
  avgSentimentScore: number;
  issueTypeStats: Array<{ issue_type: string; count: number }>;
}

export interface ImprovementOrder {
  id: string;
  orderNo: string;
  caseId: string;
  caseNo: string;
  caseTitle: string;
  applicantName: string;
  mediatorId: string;
  mediatorName: string;
  orgName: string;
  satisfactionScore: number;
  satisfactionComment: string;
  sentimentEmotion: string;
  sentimentScore: number;
  sentimentSummary: string;
  issueType: string;
  issueDescription: string;
  improvementSuggestion: string;
  status: number;
  priority: number;
  deadline: string;
  assignedAt: string;
  rectifyContent: string;
  rectifyResult: string;
  rectifiedAt: string;
  reviewOpinion: string;
  reviewedByName: string;
  reviewedAt: string;
  deductionScore: number;
  deductionReason: string;
  isDeductionApplied: number;
  remark: string;
  createdAt: string;
}

export const satisfactionService = {
  analyzeSatisfaction: (caseId: string) => {
    return request.post(`/satisfaction/analyze/${caseId}`);
  },
  getStats: (params?: { orgId?: number; startDate?: string; endDate?: string }) => {
    return request.get<SatisfactionSentimentStats>('/satisfaction/stats', { params });
  },
};

export const improvementService = {
  getList: (params?: { mediatorId?: number; status?: number; page?: number; pageSize?: number }) => {
    return request.get<{ list: ImprovementOrder[]; total: number; page: number; pageSize: number }>('/improvement', { params });
  },
  getDetail: (id: string) => {
    return request.get<ImprovementOrder>(`/improvement/${id}`);
  },
  submitRectification: (id: string, content: string, result: string) => {
    return request.post(`/improvement/${id}/rectify`, { content, result });
  },
  review: (id: string, opinion: string, approved: boolean) => {
    return request.post(`/improvement/${id}/review`, { opinion, approved });
  },
  close: (id: string, remark?: string) => {
    return request.post(`/improvement/${id}/close`, null, { params: { remark } });
  },
};

export const ImprovementStatusMap: Record<number, { text: string; color: string }> = {
  10: { text: '待整改', color: 'error' },
  20: { text: '整改中', color: 'warning' },
  30: { text: '已整改', color: 'processing' },
  40: { text: '已审核', color: 'success' },
  99: { text: '已关闭', color: 'default' },
};

export const IssueTypeMap: Record<string, string> = {
  attitude: '态度问题',
  efficiency: '效率问题',
  professional: '专业性问题',
  result: '结果不满意',
  process: '流程问题',
  other: '其他',
};

export const PriorityMap: Record<number, { text: string; color: string }> = {
  1: { text: '高', color: 'error' },
  2: { text: '中', color: 'warning' },
  3: { text: '低', color: 'default' },
};
