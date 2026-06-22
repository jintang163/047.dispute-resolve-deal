import { request } from '../utils/request';

export interface PatrolTask {
  id: number;
  taskNo: string;
  title: string;
  description?: string;
  taskType: number;
  typeName?: string;
  priority: number;
  priorityName?: string;
  assigneeId: number;
  assigneeName?: string;
  assigneeMemberNo?: string;
  status: number;
  statusName?: string;
  pointsReward: number;
  actualPoints?: number;
  startTime?: string;
  endTime?: string;
  startedAt?: string;
  completedAt?: string;
  cancelReason?: string;
  orgId?: number;
  organizationId?: number;
  gridCodes?: string;
  assignerId?: number;
  assignerName?: string;
  createdAt?: string;
  updatedAt?: string;
  points?: PatrolTaskPoint[];
  visitCount?: number;
  dangerCount?: number;
}

export interface PatrolTaskPoint {
  id: number;
  taskId: number;
  pointNo: string;
  pointName: string;
  address: string;
  longitude: number;
  latitude: number;
  pointType: number;
  checkinType: number;
  checkinRadius: number;
  requiredPhotos: number;
  sortOrder: number;
  isChecked?: boolean;
  createdAt?: string;
}

export interface CreatePatrolTaskParams {
  title: string;
  description: string;
  taskType: number;
  priority: number;
  assigneeId: number;
  startTime?: string;
  endTime?: string;
  orgId?: number;
  gridCodes?: string;
  points: {
    pointName: string;
    address: string;
    longitude: number;
    latitude: number;
    pointType: number;
    checkinType: number;
    checkinRadius?: number;
    requiredPhotos?: number;
  }[];
}

export interface UpdatePatrolTaskParams {
  title?: string;
  description?: string;
  taskType?: number;
  priority?: number;
  assigneeId?: number;
  startTime?: string;
  endTime?: string;
  status?: number;
}

export interface PageResult<T> {
  list: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface GridMember {
  id: number;
  userId?: number;
  memberNo: string;
  realName: string;
  phone?: string;
  organizationId?: number;
  orgId?: number;
  gridCodes?: string;
  status: number;
  points?: number;
  totalPoints?: number;
  level?: number;
  levelName?: string;
  joinDate?: string;
  createdAt?: string;
  updatedAt?: string;
  username?: string;
}

export interface CreateGridMemberParams {
  userId: number;
  realName: string;
  phone?: string;
  orgId?: number;
  gridCodes?: string;
  status?: number;
}

export interface UpdateGridMemberParams {
  realName?: string;
  phone?: string;
  orgId?: number;
  gridCodes?: string;
  status?: number;
}

export interface PointRule {
  id: number;
  ruleCode: string;
  ruleName: string;
  ruleType: string;
  points: number;
  maxPointsPerDay: number;
  maxPointsPerMonth: number;
  isActive: number;
  description?: string;
  expireDays: number;
  sortOrder: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface PointRecord {
  id: number;
  recordNo: string;
  memberId: number;
  memberName?: string;
  organizationId?: number;
  type: number;
  typeName?: string;
  businessType?: string;
  businessNo?: string;
  points: number;
  balanceBefore?: number;
  balanceAfter?: number;
  description?: string;
  expireDate?: string;
  isExpired?: number;
  createdAt?: string;
}

export interface GiftCategory {
  id: number;
  name: string;
  parentId?: number;
  iconUrl?: string;
  description?: string;
  sortOrder: number;
  status: number;
  giftCount?: number;
  createdAt?: string;
  updatedAt?: string;
  children?: GiftCategory[];
}

export interface Gift {
  id: number;
  giftNo: string;
  categoryId: number;
  categoryName?: string;
  name: string;
  description?: string;
  imageUrl?: string;
  bannerImageUrl?: string;
  pointsRequired: number;
  originalPrice?: number;
  stock: number;
  soldCount?: number;
  exchangeLimit: number;
  status: number;
  isHot: number;
  isNew: number;
  isRecommend: number;
  sortOrder: number;
  weight?: number;
  freight?: number;
  virtualType: number;
  createdAt?: string;
  updatedAt?: string;
}

export interface CreateGiftParams {
  categoryId: number;
  name: string;
  description?: string;
  imageUrl?: string;
  bannerImageUrl?: string;
  pointsRequired: number;
  originalPrice?: number;
  stock: number;
  exchangeLimit?: number;
  status?: number;
  isHot?: number;
  isNew?: number;
  isRecommend?: number;
  sortOrder?: number;
  weight?: number;
  freight?: number;
  virtualType?: number;
}

export interface UpdateGiftParams {
  categoryId?: number;
  name?: string;
  description?: string;
  imageUrl?: string;
  bannerImageUrl?: string;
  pointsRequired?: number;
  originalPrice?: number;
  stock?: number;
  exchangeLimit?: number;
  status?: number;
  isHot?: number;
  isNew?: number;
  isRecommend?: number;
  sortOrder?: number;
  weight?: number;
  freight?: number;
  virtualType?: number;
}

export interface GiftExchange {
  id: number;
  exchangeNo: string;
  memberId: number;
  memberName?: string;
  memberPhone?: string;
  organizationId?: number;
  giftId: number;
  giftName?: string;
  giftImage?: string;
  giftDescription?: string;
  giftPoints: number;
  quantity: number;
  totalPoints: number;
  receiverName?: string;
  receiverPhone?: string;
  receiverAddress?: string;
  status: number;
  statusName?: string;
  expressCompany?: string;
  expressNo?: string;
  remark?: string;
  auditRemark?: string;
  auditTime?: string;
  shipTime?: string;
  receiveTime?: string;
  cancelTime?: string;
  cancelReason?: string;
  createdAt?: string;
}

export interface AuditExchangeParams {
  id: number;
  status: number;
  remark?: string;
}

export interface ShipExchangeParams {
  id: number;
  expressCompany: string;
  expressNo: string;
}

export interface PatrolCheckRecord {
  id: number;
  checkinNo: string;
  taskId: number;
  taskPointId: number;
  memberId: number;
  memberName?: string;
  longitude: number;
  latitude: number;
  locationAccuracy?: number;
  address?: string;
  photoUrl?: string;
  livePhotoUrl?: string;
  isLiveVerified: number;
  liveVerifyScore?: number;
  checkinDistance: number;
  checkinRadius: number;
  isValid: number;
  invalidReason?: string;
  ipAddress?: string;
  deviceInfo?: string;
  checkinTime?: string;
  remark?: string;
  createdAt?: string;
  pointName?: string;
  taskName?: string;
}

export interface HiddenDanger {
  id: number;
  dangerNo: string;
  reporterId: number;
  reporterName?: string;
  taskId?: number;
  taskPointId?: number;
  dangerType: number;
  dangerTypeName?: string;
  level: number;
  levelName?: string;
  title: string;
  description: string;
  longitude?: number;
  latitude?: number;
  address?: string;
  photoUrls?: string;
  videoUrl?: string;
  involvedPerson?: string;
  organizationId?: number;
  status: number;
  statusName?: string;
  source: number;
  sourceName?: string;
  handlerId?: number;
  handlerName?: string;
  handleResult?: string;
  handledAt?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface PatrolVisitRecord {
  id: number;
  recordNo: string;
  memberId: number;
  memberName?: string;
  taskId?: number;
  taskPointId?: number;
  visitType: number;
  visitTypeName?: string;
  visitObject: string;
  visitContent: string;
  visitResult?: string;
  longitude?: number;
  latitude?: number;
  address?: string;
  photoUrls?: string;
  residentId?: number;
  disputeCaseId?: number;
  organizationId?: number;
  status: number;
  statusName?: string;
  remark?: string;
  auditRemark?: string;
  auditTime?: string;
  auditorId?: number;
  createdAt?: string;
  updatedAt?: string;
  taskName?: string;
}

export interface PointsSummary {
  memberId: number;
  memberName?: string;
  balance: number;
  todayEarned: number;
  monthEarned: number;
  totalEarned: number;
  totalSpent: number;
  level: number;
  levelName: string;
  nextLevelPoints: number;
  levelProgress: number;
  checkinDays: number;
  checkinContinuousDays?: number;
  rankInOrg?: number;
}

export interface PlanRouteResult {
  totalDistance: number;
  totalDuration: number;
  totalTaxiCost: number;
  strategy: number;
  strategyName: string;
  points: {
    originalIndex: number;
    sortedIndex: number;
    pointName: string;
    address: string;
    longitude: number;
    latitude: number;
    distanceFromPrev: number;
    durationFromPrev: number;
  }[];
  paths: any[];
  localOptimization?: boolean;
  amapAvailable?: boolean;
  amapError?: string;
}

export const patrolService = {
  getTaskList: (params?: {
    status?: number;
    assigneeId?: number;
    taskType?: number;
    priority?: number;
    keyword?: string;
    orgId?: number;
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<PatrolTask>>('/patrol/task', { params });
  },

  getTaskDetail: (id: number) => {
    return request.get<PatrolTask>(`/patrol/task/${id}`);
  },

  createTask: (params: CreatePatrolTaskParams) => {
    return request.post<number>('/patrol/task', params);
  },

  updateTask: (id: number, params: UpdatePatrolTaskParams) => {
    return request.put<void>(`/patrol/task/${id}`, params);
  },

  deleteTask: (id: number) => {
    return request.delete<void>(`/patrol/task/${id}`);
  },

  cancelTask: (id: number, reason?: string) => {
    return request.post<void>(`/patrol/task/${id}/cancel`, { reason });
  },

  startTask: (id: number) => {
    return request.post<void>(`/patrol/task/${id}/start`, {});
  },

  completeTask: (id: number) => {
    return request.post<void>(`/patrol/task/${id}/complete`, {});
  },

  getTaskPoints: (taskId: number) => {
    return request.get<PatrolTaskPoint[]>(`/patrol/task/${taskId}/points`);
  },

  planRoute: (startLng: number, startLat: number, points: {
    pointName: string;
    address: string;
    longitude: number;
    latitude: number;
  }[], strategy: number = 10) => {
    return request.post<PlanRouteResult>('/patrol/route/plan', {
      startLng,
      startLat,
      points,
      strategy
    });
  },

  createCheckin: (params: {
    taskId?: number;
    taskPointId?: number;
    memberId: number;
    longitude: number;
    latitude: number;
    locationAccuracy?: number;
    address?: string;
    photoUrl?: string;
    livePhotoUrl?: string;
    checkinDistance?: number;
    checkinRadius?: number;
    remark?: string;
  }) => {
    return request.post<number>('/patrol/checkin', params);
  },

  getCheckinRecords: (params?: {
    memberId?: number;
    status?: number;
    startDate?: string;
    endDate?: string;
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<PatrolCheckRecord>>('/patrol/checkin/records', { params });
  },

  getCheckinStatistics: (memberId?: number) => {
    return request.get<{
      todayCount: number;
      weekCount: number;
      monthCount: number;
      totalCount: number;
    }>('/patrol/checkin/statistics', { params: memberId ? { memberId } : {} });
  },

  createVisitRecord: (params: {
    memberId: number;
    taskId?: number;
    taskPointId?: number;
    visitType: number;
    visitObject: string;
    visitContent: string;
    visitResult?: string;
    longitude?: number;
    latitude?: number;
    address?: string;
    photoUrls?: string;
    residentId?: number;
    disputeCaseId?: number;
    organizationId?: number;
    remark?: string;
  }) => {
    return request.post<number>('/patrol/visit', params);
  },

  getVisitRecords: (params?: {
    memberId?: number;
    status?: number;
    visitType?: number;
    startDate?: string;
    endDate?: string;
    orgId?: number;
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<PatrolVisitRecord>>('/patrol/visit', { params });
  },

  getVisitRecordDetail: (id: number) => {
    return request.get<PatrolVisitRecord>(`/patrol/visit/${id}`);
  },

  updateVisitRecord: (id: number, params: Partial<PatrolVisitRecord>) => {
    return request.put<void>(`/patrol/visit/${id}`, params);
  },

  auditVisitRecord: (id: number, status: number, remark?: string) => {
    return request.post<void>(`/patrol/visit/${id}/audit`, { status, remark });
  },

  deleteVisitRecord: (id: number) => {
    return request.delete<void>(`/patrol/visit/${id}`);
  },

  getVisitStatistics: (memberId?: number) => {
    return request.get<{
      todayCount: number;
      monthCount: number;
      totalCount: number;
      typeCount: Record<string, number>;
    }>('/patrol/visit/statistics', { params: memberId ? { memberId } : {} });
  },

  reportDanger: (params: {
    reporterId: number;
    taskId?: number;
    taskPointId?: number;
    dangerType: number;
    level: number;
    title: string;
    description: string;
    longitude?: number;
    latitude?: number;
    address?: string;
    photoUrls?: string;
    videoUrl?: string;
    involvedPerson?: string;
    organizationId?: number;
    source?: number;
  }) => {
    return request.post<number>('/patrol/danger', params);
  },

  getDangerList: (params?: {
    reporterId?: number;
    status?: number;
    dangerType?: number;
    level?: number;
    keyword?: string;
    orgId?: number;
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<HiddenDanger>>('/patrol/danger', { params });
  },

  getDangerDetail: (id: number) => {
    return request.get<HiddenDanger>(`/patrol/danger/${id}`);
  },

  handleDanger: (id: number, status: number, result?: string) => {
    return request.post<void>(`/patrol/danger/${id}/handle`, { status, result });
  },

  getDangerStatistics: () => {
    return request.get<{
      total: number;
      statusStats: Record<string, number>;
      typeStats: Record<string, number>;
    }>('/patrol/danger/statistics');
  },

  getMemberList: (params?: {
    orgId?: number;
    status?: number;
    keyword?: string;
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<GridMember>>('/patrol/member', { params });
  },

  getCurrentMember: () => {
    return request.get<GridMember>('/patrol/member/me');
  },

  getMyTasks: (params?: {
    status?: number;
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<PatrolTask>>('/patrol/my/tasks', { params });
  },

  getMemberDetail: (id: number) => {
    return request.get<GridMember>(`/patrol/member/${id}`);
  },

  createMember: (params: CreateGridMemberParams) => {
    return request.post<number>('/patrol/member', params);
  },

  updateMember: (id: number, params: UpdateGridMemberParams) => {
    return request.put<void>(`/patrol/member/${id}`, params);
  },

  deleteMember: (id: number) => {
    return request.delete<void>(`/patrol/member/${id}`);
  },

  getMemberOptions: (orgId?: number) => {
    const params: any = { status: 1, pageSize: 1000 };
    if (orgId) params.orgId = orgId;
    return request.get<PageResult<GridMember>>('/patrol/member', { params })
      .then(res => res.list.map(m => ({ id: m.id, name: m.realName, memberNo: m.memberNo })));
  },

  getPointsSummary: () => {
    return request.get<PointsSummary>('/points/summary');
  },

  addPoints: (memberId: number, points: number, businessType?: string, businessNo?: string, description?: string) => {
    return request.post<void>('/points/add', { memberId, points, businessType, businessNo, description });
  },

  deductPoints: (memberId: number, points: number, businessType?: string, businessNo?: string, description?: string) => {
    return request.post<void>('/points/deduct', { memberId, points, businessType, businessNo, description });
  },

  getPointsRecords: (memberId?: number, params?: {
    page?: number;
    pageSize?: number;
    type?: number;
    startDate?: string;
    endDate?: string;
  }) => {
    const query: any = { ...params };
    if (memberId) query.memberId = memberId;
    return request.get<PageResult<PointRecord>>('/points/records', { params: query });
  },

  getPointsRules: (ruleType?: string) => {
    const params: any = {};
    if (ruleType) params.ruleType = ruleType;
    return request.get<PointRule[]>('/points/rules', { params });
  },

  createPointsRule: (params: Partial<PointRule>) => {
    return request.post<number>('/points/rules', params);
  },

  updatePointsRule: (id: number, params: Partial<PointRule>) => {
    return request.put<void>(`/points/rules/${id}`, params);
  },

  deletePointsRule: (id: number) => {
    return request.delete<void>(`/points/rules/${id}`);
  },

  exchangeGift: (giftId: number, quantity: number = 1, receiverName?: string, receiverPhone?: string, receiverAddress?: string, remark?: string) => {
    return request.post<number>('/points/exchange', { giftId, quantity, receiverName, receiverPhone, receiverAddress, remark });
  },

  processExpiredPoints: () => {
    return request.post<void>('/points/process-expired', {});
  },

  getGiftList: (params?: {
    categoryId?: number;
    status?: number;
    keyword?: string;
    isHot?: number;
    isNew?: number;
    minPoints?: number;
    maxPoints?: number;
    sortBy?: string;
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<Gift>>('/gift', { params });
  },

  getGiftDetail: (id: number) => {
    return request.get<Gift>(`/gift/${id}`);
  },

  createGift: (params: CreateGiftParams) => {
    return request.post<number>('/gift', params);
  },

  updateGift: (id: number, params: UpdateGiftParams) => {
    return request.put<void>(`/gift/${id}`, params);
  },

  deleteGift: (id: number) => {
    return request.delete<void>(`/gift/${id}`);
  },

  getGiftCategories: () => {
    return request.get<GiftCategory[]>('/gift/categories');
  },

  createGiftCategory: (params: Partial<GiftCategory>) => {
    return request.post<number>('/gift/categories', params);
  },

  updateGiftCategory: (id: number, params: Partial<GiftCategory>) => {
    return request.put<void>(`/gift/categories/${id}`, params);
  },

  deleteGiftCategory: (id: number) => {
    return request.delete<void>(`/gift/categories/${id}`);
  },

  getExchangeList: (params?: {
    memberId?: number;
    status?: number;
    giftId?: number;
    startDate?: string;
    endDate?: string;
    keyword?: string;
    orgId?: number;
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<GiftExchange>>('/gift/exchange', { params });
  },

  getExchangeDetail: (id: number) => {
    return request.get<GiftExchange>(`/gift/exchange/${id}`);
  },

  auditExchange: (params: AuditExchangeParams) => {
    return request.post<void>(`/gift/exchange/${params.id}/audit`, { status: params.status, remark: params.remark });
  },

  shipExchange: (params: ShipExchangeParams) => {
    return request.post<void>(`/gift/exchange/${params.id}/ship`, { expressCompany: params.expressCompany, expressNo: params.expressNo });
  },

  receiveExchange: (id: number) => {
    return request.post<void>(`/gift/exchange/${id}/receive`, {});
  },

  cancelExchange: (id: number, reason?: string) => {
    return request.post<void>(`/gift/exchange/${id}/cancel`, { reason });
  },

  getMyExchanges: (params?: {
    page?: number;
    pageSize?: number;
  }) => {
    return request.get<PageResult<GiftExchange>>('/gift/my/exchanges', { params });
  },

  getGiftStatistics: () => {
    return request.get<{
      totalGifts: number;
      totalExchanges: number;
      totalPoints: number;
      pendingCount: number;
      shippedCount: number;
      categoryStats: Record<string, number>;
    }>('/gift/statistics');
  },

  getTaskTypes: () => {
    return Promise.resolve([
      { code: 1, name: '日常排查' },
      { code: 2, name: '专项排查' },
      { code: 3, name: '重点排查' },
      { code: 4, name: '紧急排查' },
    ]);
  },

  getPriorities: () => {
    return Promise.resolve([
      { code: 1, name: '紧急', color: 'red' },
      { code: 2, name: '高', color: 'orange' },
      { code: 3, name: '中', color: 'blue' },
      { code: 4, name: '低', color: 'green' },
    ]);
  },

  getTaskStatusList: () => {
    return Promise.resolve([
      { code: 10, name: '待执行', color: 'default' },
      { code: 20, name: '进行中', color: 'processing' },
      { code: 30, name: '已完成', color: 'success' },
      { code: 50, name: '已取消', color: 'error' },
    ]);
  },

  getDangerTypes: () => {
    return Promise.resolve([
      { code: 1, name: '消防安全' },
      { code: 2, name: '治安隐患' },
      { code: 3, name: '矛盾纠纷' },
      { code: 4, name: '安全生产' },
      { code: 5, name: '环境卫生' },
      { code: 6, name: '其他隐患' },
    ]);
  },

  getDangerLevels: () => {
    return Promise.resolve([
      { code: 1, name: '一般', color: 'blue' },
      { code: 2, name: '较大', color: 'orange' },
      { code: 3, name: '重大', color: 'red' },
      { code: 4, name: '特别重大', color: 'magenta' },
    ]);
  },

  getVisitTypes: () => {
    return Promise.resolve([
      { code: 1, name: '日常走访' },
      { code: 2, name: '重点人员走访' },
      { code: 3, name: '纠纷回访' },
      { code: 4, name: '特殊人群走访' },
      { code: 5, name: '隐患排查' },
    ]);
  },

  getExchangeStatusList: () => {
    return Promise.resolve([
      { code: 10, name: '待审核', color: 'warning' },
      { code: 20, name: '审核通过', color: 'processing' },
      { code: 30, name: '审核拒绝', color: 'error' },
      { code: 40, name: '已发货', color: 'processing' },
      { code: 50, name: '已取消', color: 'default' },
      { code: 60, name: '已完成', color: 'success' },
    ]);
  },

  getPointRuleTypes: () => {
    return Promise.resolve([
      { code: 'earn', name: '获得积分' },
      { code: 'spend', name: '消费积分' },
      { code: 'expire', name: '过期积分' },
    ]);
  },

  getTaskCheckRecords: (taskId: number) => {
    return patrolService.getTaskDetail(taskId)
      .then(() => patrolService.getCheckinRecords())
      .then(res => res.list.filter(r => r.taskId === taskId));
  },
};
