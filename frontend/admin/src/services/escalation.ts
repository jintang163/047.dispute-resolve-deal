import request from '@/utils/request';

export interface EscalationRecord {
  id: number;
  caseId: number;
  caseNo: string;
  escalateType: number;
  fromLevel: number;
  toLevel: number;
  fromUserId: number;
  fromUserName: string;
  toUserId: number;
  toUserName: string;
  toOrgId: number;
  toOrgName: string;
  reason: string;
  urgeCount: number;
  firstUrgeTime?: string;
  timeoutHours: number;
  operatorId: number;
  operatorName: string;
  status: number;
  remark: string;
  createdAt: string;
  updatedAt: string;
}

export interface UrgeRecord {
  id: number;
  caseId: number;
  urgeType: number;
  urgeLevel: number;
  escalateTriggered: number;
  urgeContent: string;
  operatorId: number;
  operatorName: string;
  createdAt: string;
}

export interface EscalationListParams {
  page?: number;
  pageSize?: number;
  toLevel?: number;
  status?: number;
}

export interface EscalationListResponse {
  list: EscalationRecord[];
  total: number;
  page: number;
  size: number;
}

export const escalationApi = {
  getList: (params: EscalationListParams): Promise<EscalationListResponse> => {
    return request.get('/api/v1/escalation', { params });
  },

  getDetail: (id: number): Promise<EscalationRecord> => {
    return request.get(`/api/v1/escalation/${id}`);
  },

  handle: (id: number, remark?: string): Promise<void> => {
    return request.post(`/api/v1/escalation/${id}/handle`, { remark });
  },

  close: (id: number, remark: string): Promise<void> => {
    return request.post(`/api/v1/escalation/${id}/close`, { remark });
  },

  getCaseEscalations: (caseId: number): Promise<EscalationRecord[]> => {
    return request.get(`/api/v1/dispute/${caseId}/escalations`);
  },

  getCaseUrges: (caseId: number): Promise<UrgeRecord[]> => {
    return request.get(`/api/v1/urge/case/${caseId}`);
  },
};

export const ESCALATION_TYPE_MAP: Record<number, string> = {
  1: '待分派超时',
  2: '调解中超时无进展',
};

export const ESCALATION_LEVEL_MAP: Record<number, string> = {
  0: '未分派',
  1: '调解员',
  2: '组长',
  3: '主任',
};

export const ESCALATION_STATUS_MAP: Record<number, string> = {
  10: '待处理',
  20: '处理中',
  30: '已处理',
  40: '已关闭',
};

export const URGE_TYPE_MAP: Record<number, string> = {
  1: '用户催办',
  2: '领导催办',
  3: '系统催办',
  4: '升级催办',
};
