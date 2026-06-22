import { request } from '../utils/request';

export interface PatrolTask {
  id: string;
  taskNo: string;
  title: string;
  type: string;
  typeName?: string;
  status: string;
  statusName?: string;
  priority: string;
  priorityName?: string;
  description?: string;
  area?: string;
  areaId?: string;
  gridMemberId?: string;
  gridMemberName?: string;
  gridMemberPhone?: string;
  deadline?: string;
  longitude?: number;
  latitude?: number;
  requirement?: string;
  remark?: string;
  startTime?: string;
  endTime?: string;
  actualStartTime?: string;
  actualEndTime?: string;
  result?: string;
  attachmentIds?: string[];
  createTime?: string;
  updateTime?: string;
  creator?: string;
  creatorName?: string;
}

export interface PatrolTaskParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  type?: string;
  status?: string;
  priority?: string;
  gridMemberId?: string;
  startDate?: string;
  endDate?: string;
}

export interface PatrolTaskResponse {
  list: PatrolTask[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface CreatePatrolTaskParams {
  title: string;
  type: string;
  priority: string;
  description?: string;
  area?: string;
  areaId?: string;
  gridMemberId: string;
  deadline?: string;
  longitude?: number;
  latitude?: number;
  requirement?: string;
  remark?: string;
  startTime?: string;
  endTime?: string;
  attachmentIds?: string[];
}

export interface GridMember {
  id: string;
  memberNo: string;
  name: string;
  gender: string;
  genderName?: string;
  phone: string;
  idCard?: string;
  avatar?: string;
  area: string;
  areaId?: string;
  address?: string;
  email?: string;
  status: string;
  statusName?: string;
  points?: number;
  totalPoints?: number;
  usedPoints?: number;
  taskCount?: number;
  completedTaskCount?: number;
  joinDate?: string;
  createTime?: string;
  updateTime?: string;
}

export interface GridMemberParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  status?: string;
  areaId?: string;
}

export interface GridMemberResponse {
  list: GridMember[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface CreateGridMemberParams {
  name: string;
  gender: string;
  phone: string;
  idCard?: string;
  avatar?: string;
  area: string;
  areaId?: string;
  address?: string;
  email?: string;
  status?: string;
}

export interface PointRule {
  id: string;
  name: string;
  code: string;
  type: string;
  typeName?: string;
  points: number;
  description?: string;
  status: string;
  statusName?: string;
  sort?: number;
  createTime?: string;
  updateTime?: string;
}

export interface PointRuleParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  type?: string;
  status?: string;
}

export interface PointRuleResponse {
  list: PointRule[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface PointFlow {
  id: string;
  flowNo: string;
  memberId: string;
  memberName?: string;
  type: string;
  typeName?: string;
  ruleId?: string;
  ruleName?: string;
  points: number;
  balance: number;
  taskId?: string;
  taskNo?: string;
  orderId?: string;
  orderNo?: string;
  remark?: string;
  createTime?: string;
  operator?: string;
  operatorName?: string;
}

export interface PointFlowParams {
  pageNum?: number;
  pageSize?: number;
  memberId?: string;
  memberName?: string;
  type?: string;
  startDate?: string;
  endDate?: string;
}

export interface PointFlowResponse {
  list: PointFlow[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface GiftCategory {
  id: string;
  name: string;
  code: string;
  parentId?: string;
  level?: number;
  sort?: number;
  status: string;
  statusName?: string;
  description?: string;
  createTime?: string;
  updateTime?: string;
  children?: GiftCategory[];
}

export interface Gift {
  id: string;
  giftNo: string;
  name: string;
  categoryId: string;
  categoryName?: string;
  points: number;
  originalPrice?: number;
  stock: number;
  soldCount?: number;
  description?: string;
  images?: string[];
  status: string;
  statusName?: string;
  sort?: number;
  isHot?: boolean;
  isNew?: boolean;
  createTime?: string;
  updateTime?: string;
}

export interface GiftParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  categoryId?: string;
  status?: string;
}

export interface GiftResponse {
  list: Gift[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface CreateGiftParams {
  name: string;
  categoryId: string;
  points: number;
  originalPrice?: number;
  stock: number;
  description?: string;
  images?: string[];
  status?: string;
  sort?: number;
  isHot?: boolean;
  isNew?: boolean;
}

export interface ExchangeOrder {
  id: string;
  orderNo: string;
  memberId: string;
  memberName?: string;
  memberPhone?: string;
  giftId: string;
  giftName?: string;
  giftImage?: string;
  giftPoints?: number;
  quantity: number;
  totalPoints: number;
  receiverName?: string;
  receiverPhone?: string;
  receiverAddress?: string;
  status: string;
  statusName?: string;
  expressCompany?: string;
  expressNo?: string;
  auditRemark?: string;
  auditTime?: string;
  auditOperator?: string;
  auditOperatorName?: string;
  deliveryTime?: string;
  receiveTime?: string;
  remark?: string;
  createTime?: string;
  updateTime?: string;
}

export interface ExchangeOrderParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  orderNo?: string;
  memberName?: string;
  status?: string;
  startDate?: string;
  endDate?: string;
}

export interface ExchangeOrderResponse {
  list: ExchangeOrder[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface AuditOrderParams {
  orderId: string;
  status: string;
  remark?: string;
}

export interface DeliveryOrderParams {
  orderId: string;
  expressCompany: string;
  expressNo: string;
  remark?: string;
}

export interface PatrolTaskDetail {
  taskInfo: PatrolTask;
  gridMember?: GridMember;
  checkRecords?: PatrolCheckRecord[];
}

export interface PatrolCheckRecord {
  id: string;
  taskId: string;
  checkTime?: string;
  longitude?: number;
  latitude?: number;
  location?: string;
  description?: string;
  images?: string[];
  issue?: string;
  issueLevel?: string;
  issueLevelName?: string;
  handleResult?: string;
  createTime?: string;
  operator?: string;
  operatorName?: string;
}

export const patrolService = {
  getTaskList: (params?: PatrolTaskParams) => {
    return request.get<PatrolTaskResponse>('/patrol/task/list', { params });
  },

  getTaskDetail: (id: string) => {
    return request.get<PatrolTaskDetail>(`/patrol/task/${id}`);
  },

  createTask: (params: CreatePatrolTaskParams) => {
    return request.post<{ id: string }>('/patrol/task/create', params);
  },

  updateTask: (id: string, params: Partial<CreatePatrolTaskParams>) => {
    return request.put(`/patrol/task/${id}`, params);
  },

  deleteTask: (id: string) => {
    return request.delete(`/patrol/task/${id}`);
  },

  assignTask: (id: string, gridMemberId: string, remark?: string) => {
    return request.post(`/patrol/task/${id}/assign`, { gridMemberId, remark });
  },

  cancelTask: (id: string, reason?: string) => {
    return request.post(`/patrol/task/${id}/cancel`, { reason });
  },

  getTaskTypes: () => {
    return request.get<{ code: string; name: string }[]>('/patrol/task/types');
  },

  getMemberList: (params?: GridMemberParams) => {
    return request.get<GridMemberResponse>('/patrol/member/list', { params });
  },

  getMemberDetail: (id: string) => {
    return request.get<GridMember>(`/patrol/member/${id}`);
  },

  createMember: (params: CreateGridMemberParams) => {
    return request.post<{ id: string }>('/patrol/member/create', params);
  },

  updateMember: (id: string, params: Partial<CreateGridMemberParams>) => {
    return request.put(`/patrol/member/${id}`, params);
  },

  deleteMember: (id: string) => {
    return request.delete(`/patrol/member/${id}`);
  },

  getMemberOptions: (areaId?: string) => {
    const params: any = {};
    if (areaId) params.areaId = areaId;
    return request.get<{ id: string; name: string; area: string }[]>('/patrol/member/options', { params });
  },

  getPointRuleList: (params?: PointRuleParams) => {
    return request.get<PointRuleResponse>('/patrol/point/rule/list', { params });
  },

  getPointRuleDetail: (id: string) => {
    return request.get<PointRule>(`/patrol/point/rule/${id}`);
  },

  createPointRule: (params: Partial<PointRule>) => {
    return request.post<{ id: string }>('/patrol/point/rule/create', params);
  },

  updatePointRule: (id: string, params: Partial<PointRule>) => {
    return request.put(`/patrol/point/rule/${id}`, params);
  },

  deletePointRule: (id: string) => {
    return request.delete(`/patrol/point/rule/${id}`);
  },

  getPointFlowList: (params?: PointFlowParams) => {
    return request.get<PointFlowResponse>('/patrol/point/flow/list', { params });
  },

  getGiftCategoryTree: () => {
    return request.get<GiftCategory[]>('/patrol/gift/category/tree');
  },

  createGiftCategory: (params: Partial<GiftCategory>) => {
    return request.post<{ id: string }>('/patrol/gift/category/create', params);
  },

  updateGiftCategory: (id: string, params: Partial<GiftCategory>) => {
    return request.put(`/patrol/gift/category/${id}`, params);
  },

  deleteGiftCategory: (id: string) => {
    return request.delete(`/patrol/gift/category/${id}`);
  },

  getGiftList: (params?: GiftParams) => {
    return request.get<GiftResponse>('/patrol/gift/list', { params });
  },

  getGiftDetail: (id: string) => {
    return request.get<Gift>(`/patrol/gift/${id}`);
  },

  createGift: (params: CreateGiftParams) => {
    return request.post<{ id: string }>('/patrol/gift/create', params);
  },

  updateGift: (id: string, params: Partial<CreateGiftParams>) => {
    return request.put(`/patrol/gift/${id}`, params);
  },

  deleteGift: (id: string) => {
    return request.delete(`/patrol/gift/${id}`);
  },

  getExchangeOrderList: (params?: ExchangeOrderParams) => {
    return request.get<ExchangeOrderResponse>('/patrol/exchange/list', { params });
  },

  getExchangeOrderDetail: (id: string) => {
    return request.get<ExchangeOrder>(`/patrol/exchange/${id}`);
  },

  auditExchangeOrder: (params: AuditOrderParams) => {
    return request.post('/patrol/exchange/audit', params);
  },

  deliveryExchangeOrder: (params: DeliveryOrderParams) => {
    return request.post('/patrol/exchange/delivery', params);
  },

  completeExchangeOrder: (id: string, remark?: string) => {
    return request.post(`/patrol/exchange/${id}/complete`, { remark });
  },

  cancelExchangeOrder: (id: string, reason?: string) => {
    return request.post(`/patrol/exchange/${id}/cancel`, { reason });
  },

  getAreaOptions: () => {
    return request.get<{ id: string; name: string; parentId?: string }[]>('/patrol/area/list');
  },

  getTaskCheckRecords: (taskId: string) => {
    return request.get<PatrolCheckRecord[]>(`/patrol/task/${taskId}/check-records`);
  },
};
