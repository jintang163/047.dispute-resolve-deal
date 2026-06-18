import { request } from '../utils/request';

export interface JudicialConfirmation {
  id: number;
  confirmNo: string;
  caseId: number;
  caseNo: string;
  caseTitle: string;
  status: number;
  statusName?: string;
  applicantName: string;
  applicantPhone: string;
  respondentName: string;
  respondentPhone: string;
  courtId: number;
  courtName: string;
  agreementContent: string;
  performanceDeadline?: string;
  confirmAmount?: number;
  courtCaseNo?: string;
  documentNo?: string;
  documentUrl?: string;
  sealStatus?: number;
  sealTime?: string;
  performanceRemindTime?: string;
  expirationRemindTime?: string;
  fulfilledTime?: string;
  remark?: string;
  orgId?: number;
  orgName?: string;
  createTime?: string;
  updateTime?: string;
  daysLeft?: number;
}

export interface JudicialConfirmLog {
  id: number;
  confirmId: number;
  confirmNo: string;
  actionType: number;
  actionTypeName: string;
  operatorId?: number;
  operatorName?: string;
  operatorType: number;
  operatorTypeName: string;
  remark?: string;
  detail?: string;
  createTime: string;
}

export interface CourtConfig {
  id: number;
  courtCode: string;
  courtName: string;
  courtLevel?: number;
  jurisdictionArea?: string;
  address?: string;
  contact?: string;
  phone?: string;
  apiEndpoint?: string;
  apiAppId?: string;
  sealCertNo?: string;
  sealImageUrl?: string;
  sortOrder?: number;
  status: number;
  createTime?: string;
  updateTime?: string;
}

export interface JudicialListParams {
  page?: number;
  pageSize?: number;
  status?: number;
  keyword?: string;
  startTime?: string;
  endTime?: string;
}

export interface JudicialListResponse {
  list: JudicialConfirmation[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CreateJudicialParams {
  caseId: number;
  caseNo?: string;
  caseTitle?: string;
  mediationRecordId?: number;
  protocolId?: number;
  applicantName: string;
  applicantPhone: string;
  applicantIdCard?: string;
  applicantAddress?: string;
  respondentName: string;
  respondentPhone: string;
  respondentIdCard?: string;
  respondentAddress?: string;
  courtId: number;
  courtName?: string;
  agreementContent: string;
  performanceDeadline?: string;
  confirmAmount?: number;
  remark?: string;
}

export interface CourtConfigParams {
  courtCode: string;
  courtName: string;
  courtLevel?: number;
  jurisdictionArea?: string;
  address?: string;
  contact?: string;
  phone?: string;
  apiEndpoint?: string;
  apiAppId?: string;
  apiSecret?: string;
  apiPublicKey?: string;
  sealCertNo?: string;
  sealImageUrl?: string;
  sortOrder?: number;
  status: number;
}

export const judicialService = {
  getList: (params?: JudicialListParams) => {
    return request.get<JudicialListResponse>('/judicial/list', { params });
  },

  getDetail: (id: number) => {
    return request.get<JudicialConfirmation>(`/judicial/${id}`);
  },

  queryByNo: (confirmNo: string, idCard: string) => {
    return request.get<JudicialConfirmation>('/judicial/query', { 
      params: { confirmNo, idCard } 
    });
  },

  create: (params: CreateJudicialParams) => {
    return request.post<{ confirmNo: string }>('/judicial', params);
  },

  submitToCourt: (id: number) => {
    return request.post(`/judicial/${id}/submit`);
  },

  queryCourtStatus: (id: number) => {
    return request.post(`/judicial/${id}/query-status`);
  },

  generateDocument: (id: number) => {
    return request.post<{ documentUrl: string }>(`/judicial/${id}/generate-doc`);
  },

  sealDocument: (id: number) => {
    return request.post(`/judicial/${id}/seal`);
  },

  getLogs: (id: number) => {
    return request.get<JudicialConfirmLog[]>(`/judicial/${id}/logs`);
  },

  sendPerformanceReminder: (id: number) => {
    return request.post(`/judicial/${id}/remind/performance`);
  },

  sendExpirationReminder: (id: number) => {
    return request.post(`/judicial/${id}/remind/expiration`);
  },

  getCourtConfigList: (params?: { page?: number; pageSize?: number; keyword?: string }) => {
    return request.get<{ list: CourtConfig[]; total: number }>('/court/config/list', { params });
  },

  getCourtConfigDetail: (id: number) => {
    return request.get<CourtConfig>(`/court/config/${id}`);
  },

  createCourtConfig: (params: CourtConfigParams) => {
    return request.post<{ id: number }>('/court/config', params);
  },

  updateCourtConfig: (id: number, params: Partial<CourtConfigParams>) => {
    return request.put(`/court/config/${id}`, params);
  },

  deleteCourtConfig: (id: number) => {
    return request.delete(`/court/config/${id}`);
  },

  getCourtOptions: () => {
    return request.get<{ id: number; courtName: string }[]>('/court/options');
  },
};

export const JudicialStatusMap: Record<number, string> = {
  10: '已提交',
  20: '审核中',
  30: '已确认',
  40: '已驳回',
  50: '已失效',
};

export const JudicialStatusColorMap: Record<number, string> = {
  10: 'default',
  20: 'processing',
  30: 'success',
  40: 'error',
  50: 'warning',
};

export const ActionTypeMap: Record<number, string> = {
  10: '提交申请',
  20: '法院受理',
  30: '审核通过',
  40: '审核驳回',
  50: '已签章',
  60: '确认书送达',
  70: '履行提醒',
  80: '失效提醒',
  90: '已履行',
  99: '已失效',
};

export const OperatorTypeMap: Record<number, string> = {
  1: '系统',
  2: '管理员',
  3: '法院',
  4: '当事人',
};
