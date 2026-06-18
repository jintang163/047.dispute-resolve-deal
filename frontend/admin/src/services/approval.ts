import { request } from '../utils/request';

export interface ApprovalItem {
  id: string;
  caseId: string;
  caseNo: string;
  caseTitle: string;
  caseType?: string;
  workflowNode: string;
  workflowNodeName: string;
  submitterId?: string;
  submitterName?: string;
  submitTime?: string;
  deadline?: string;
  priority?: string;
  status: string;
  statusName?: string;
}

export interface ApprovalListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  workflowNode?: string;
  priority?: string;
  startDate?: string;
  endDate?: string;
}

export interface ApprovalListResponse {
  list: ApprovalItem[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface ApproveParams {
  id: string;
  status: 'approved' | 'rejected';
  opinion?: string;
}

export const approvalService = {
  getTodoList: (params?: ApprovalListParams) => {
    return request.get<ApprovalListResponse>('/approval/todo', { params });
  },

  getDoneList: (params?: ApprovalListParams) => {
    return request.get<ApprovalListResponse>('/approval/done', { params });
  },

  getTodoCount: () => {
    return request.get<{ count: number }>('/approval/todo-count');
  },

  approve: (params: ApproveParams) => {
    return request.post('/approval/approve', params);
  },

  reject: (params: ApproveParams) => {
    return request.post('/approval/reject', params);
  },

  submitApproval: (caseId: string, workflowNode: string) => {
    return request.post('/approval/submit', { caseId, workflowNode });
  },

  getApprovalHistory: (caseId: string) => {
    return request.get<any[]>(`/approval/history/${caseId}`);
  },
};
