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
}

export interface PerformanceParams {
  startDate?: string;
  endDate?: string;
  orgId?: string;
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
    return request.get<PerformanceStats[]>('/performance/stats', { params });
  },

  getRank: (params?: PerformanceParams) => {
    return request.get<PerformanceStats[]>('/performance/rank', { params });
  },

  getSummary: (params?: PerformanceParams) => {
    return request.get<any>('/performance/summary', { params });
  },
};
