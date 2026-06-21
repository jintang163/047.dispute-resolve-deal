import { request } from '../utils/request';

export interface LoginParams {
  username: string;
  password: string;
  role: string;
}

export interface LoginResponse {
  token: string;
  userInfo: {
    id: string;
    username: string;
    realName: string;
    role: string;
    roleName: string;
    avatar?: string;
    orgId?: string;
    orgName?: string;
    phone?: string;
  };
}

export interface User {
  id: string;
  username: string;
  realName: string;
  role: string;
  roleName?: string;
  phone?: string;
  email?: string;
  orgId?: string;
  orgName?: string;
  status: 'active' | 'disabled';
  createTime?: string;
}

export interface UserListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  role?: string;
  status?: string;
  orgId?: string;
}

export interface UserListResponse {
  list: User[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface CreateUserParams {
  username: string;
  password: string;
  realName: string;
  role: string;
  phone?: string;
  email?: string;
  orgId?: string;
}

export interface Organization {
  id: string;
  name: string;
  code: string;
  parentId?: string;
  type?: string;
  address?: string;
  leader?: string;
  leaderPhone?: string;
  sort?: number;
  status: 'active' | 'disabled';
  children?: Organization[];
}

export interface MediationRecord {
  id: string;
  caseId: string;
  caseNo: string;
  caseTitle: string;
  mediatorId?: string;
  mediatorName?: string;
  mediationTime?: string;
  place?: string;
  duration?: number;
  result?: string;
  resultName?: string;
  protocol?: string;
  createTime?: string;
  isDraft?: boolean;
  templateId?: string;
  templateName?: string;
  recordType?: number;
  recordTypeName?: string;
  processContent?: string;
  disputeFocus?: string;
  mediationOpinion?: string;
  agreementContent?: string;
  nextStep?: string;
  participantNames?: string;
}

export interface MediationListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  mediatorId?: string;
  result?: string;
  startDate?: string;
  endDate?: string;
}

export interface MediationListResponse {
  list: MediationRecord[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface PerformanceStats {
  userId: string;
  userName: string;
  orgId?: string;
  orgName?: string;
  totalCases?: number;
  completedCases?: number;
  mediationCount?: number;
  mediationSuccessRate?: number;
  avgDuration?: number;
  satisfaction?: number;
  score?: number;
  rank?: number;
  urgeCount?: number;
  closeRate?: number;
  avgDays?: number;
  avgSatisfaction?: number;
  totalScore?: number;
  level?: string;
  caseCount?: number;
  closedCount?: number;
  successRate?: number;
}

export interface PerformanceParams {
  startDate?: string;
  endDate?: string;
  orgId?: string;
  year?: number;
  month?: number;
  userId?: number;
  organizationId?: number;
}

export interface PerformanceDashboardData {
  summary: {
    year: number;
    month: number;
    mediatorCount: number;
    totalCases: number;
    totalClosed: number;
    totalSuccess: number;
    totalUrge: number;
    avgCloseRate: number;
    avgSuccessRate: number;
    avgDays: number;
    avgSatisfaction: number;
    avgScore: number;
  };
  mediators: PerformanceStats[];
}

export interface PerformanceMonthComparison {
  current: Record<string, any>;
  previous: Record<string, any>;
  comparison: {
    caseCountChange: number;
    closeRateChange: number;
    successRateChange: number;
    avgDaysChange: number;
    avgSatisfactionChange: number;
    urgeCountChange: number;
    totalScoreChange: number;
  };
  trend: Record<string, any>[];
}

export interface IndicatorConfig {
  id: number;
  indicator_code: string;
  indicator_name: string;
  indicator_type: number;
  weight: number;
  max_score: number;
  description: string;
  status: number;
}

export interface PerformanceInterview {
  id: number;
  interview_no: string;
  user_id: number;
  user_name: string;
  interviewer_id: number;
  interviewer_name: string;
  interview_time: string;
  interview_place: string;
  interview_type: number;
  interview_type_name: string;
  strengths: string;
  weaknesses: string;
  improvement_plan: string;
  target_next_period: string;
  mediator_comment: string;
  status: number;
  status_name: string;
  total_score: number;
  level: string;
  period_type: number;
  period_value: string;
}

export const userService = {
  login: (params: LoginParams) => {
    return request.post<LoginResponse>('/auth/login', params);
  },

  logout: () => {
    return request.post('/auth/logout');
  },

  getCurrentUser: () => {
    return request.get<User>('/auth/userinfo');
  },

  getList: (params?: UserListParams) => {
    return request.get<UserListResponse>('/system/users', { params });
  },

  create: (params: CreateUserParams) => {
    return request.post<{ id: string }>('/system/users', params);
  },

  update: (id: string, params: Partial<CreateUserParams> & { status?: string }) => {
    return request.put(`/system/users/${id}`, params);
  },

  delete: (id: string) => {
    return request.delete(`/system/users/${id}`);
  },

  resetPassword: (id: string, newPassword: string) => {
    return request.post(`/system/users/${id}/reset-password`, { newPassword });
  },

  getRoles: () => {
    return request.get<{ id: string; name: string; code: string }[]>('/system/roles');
  },
};

export const orgService = {
  getTree: () => {
    return request.get<Organization[]>('/system/orgs/tree');
  },

  getList: () => {
    return request.get<Organization[]>('/system/orgs');
  },

  create: (params: Partial<Organization>) => {
    return request.post<{ id: string }>('/system/orgs', params);
  },

  update: (id: string, params: Partial<Organization>) => {
    return request.put(`/system/orgs/${id}`, params);
  },

  delete: (id: string) => {
    return request.delete(`/system/orgs/${id}`);
  },
};

export const mediationService = {
  getList: (params?: MediationListParams) => {
    return request.get<MediationListResponse>('/mediation/list', { params });
  },

  getDetail: (id: string) => {
    return request.get<any>(`/mediation/${id}`);
  },

  create: (params: any) => {
    return request.post<{ id: string }>('/mediation/create', params);
  },

  getMediators: (orgId?: string) => {
    return request.get<User[]>(`/mediation/mediators${orgId ? `?orgId=${orgId}` : ''}`);
  },
};

export const performanceService = {
  getStats: (params?: PerformanceParams) => {
    return request.get<PerformanceStats[]>('/v1/performance', { params });
  },

  getRank: (params?: PerformanceParams) => {
    return request.get<PerformanceStats[]>('/v1/performance/ranking', { params });
  },

  getSummary: (params?: PerformanceParams) => {
    return request.get<any>('/v1/performance/dashboard', { params });
  },

  getDashboard: (params?: PerformanceParams) => {
    return request.get<PerformanceDashboardData>('/v1/performance/dashboard', { params });
  },

  getMonthComparison: (params?: PerformanceParams) => {
    return request.get<PerformanceMonthComparison>('/v1/performance/month-comparison', { params });
  },

  getTrend: (params?: PerformanceParams) => {
    return request.get<any>('/v1/performance/trend', { params });
  },

  getIndicatorConfig: () => {
    return request.get<{ indicators: IndicatorConfig[]; totalWeight: number }>('/v1/performance/indicator-config');
  },

  updateIndicatorConfig: (indicators: { id: number; weight: number }[], autoRecalculate?: boolean) => {
    return request.put('/v1/performance/indicator-config', { indicators, autoRecalculate });
  },

  batchCalculateScore: (data: { year: number; month: number; organizationId?: number }) => {
    return request.post('/v1/performance/batch-calculate', data);
  },

  getInterviewList: (params?: any) => {
    return request.get('/v1/performance/interview', { params });
  },

  createInterview: (data: any) => {
    return request.post('/v1/performance/interview', data);
  },

  getInterviewDetail: (id: number) => {
    return request.get(`/v1/performance/interview/${id}`);
  },

  confirmInterview: (id: number, mediatorComment: string) => {
    return request.post(`/v1/performance/interview/${id}/confirm`, { mediatorComment });
  },

  calculateScore: (data: any) => {
    return request.post('/v1/performance/calculate', data);
  },

  exportExcel: (params?: PerformanceParams) => {
    return request.get('/v1/performance/export', {
      params,
      responseType: 'blob',
    } as any);
  },
};
