import { request } from '../utils/request';

export interface DisputeCase {
  id: string;
  caseNo: string;
  title: string;
  typeId?: string;
  type?: string;
  typeName?: string;
  status: string;
  statusName?: string;
  description?: string;
  occurAddress?: string;
  address?: string;
  urgency?: string;
  caseLevel?: string;
  reporterName?: string;
  reporterPhone?: string;
  respondentName?: string;
  respondentPhone?: string;
  partyA?: string;
  partyB?: string;
  partyAPhone?: string;
  partyBPhone?: string;
  orgId?: string;
  organizationId?: string;
  orgName?: string;
  mediatorId?: string;
  mediatorName?: string;
  createTime?: string;
  updateTime?: string;
  createdAt?: string;
  creator?: string;
  creatorName?: string;
  keywords?: string[];
}

export interface DisputeListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
  tagKeyword?: string;
  type?: string;
  status?: string;
  orgId?: string;
  startDate?: string;
  endDate?: string;
}

export interface DisputeListResponse {
  list: DisputeCase[];
  total: number;
  pageNum: number;
  pageSize: number;
}

export interface CreateDisputeParams {
  title: string;
  typeId: number;
  description?: string;
  occurAddress?: string;
  occurTime?: string;
  expectation?: string;
  caseLevel?: number;
  caseSource?: number;
  reporterName: string;
  reporterPhone?: string;
  reporterAddress?: string;
  reporterIdCard?: string;
  respondentName: string;
  respondentPhone?: string;
  respondentAddress?: string;
  organizationId?: number;
  longitude?: number;
  latitude?: number;
  evidenceIds?: number[];
  keywords?: string[];
}

export interface KeywordExtractResult {
  keywords: string[];
  count: number;
  suggestedTypeId?: number;
  suggestedTypeName?: string;
  reason?: string;
}

export interface DisputeTypeNode {
  id: number;
  typeCode: string;
  typeName: string;
  parentId: number;
  level: number;
  icon?: string;
  children?: DisputeTypeNode[];
}

export interface KeywordStatItem {
  keyword: string;
  count: number;
  ratio: number;
  category: string;
  sampleSize: number;
}

export interface ApprovalRecord {
  id: string;
  caseId: string;
  step: number;
  stepName: string;
  approverId?: string;
  approverName?: string;
  status: string;
  statusName?: string;
  opinion?: string;
  createTime?: string;
  updateTime?: string;
}

export interface PartyInfo {
  id: string;
  caseId: string;
  type: 'A' | 'B';
  name: string;
  phone?: string;
  idCard?: string;
  address?: string;
}

export interface DisputeDetail {
  caseInfo: DisputeCase;
  parties: PartyInfo[];
  approvalRecords: ApprovalRecord[];
  history?: any[];
}

export interface HeatmapPoint {
  latitude: number;
  longitude: number;
  id?: number;
  case_no?: string;
  title?: string;
  type_name?: string;
  org_name?: string;
  status?: number;
  status_name?: string;
  created_at?: string;
  count?: number;
  organization_id?: number;
}

export interface HeatmapTimelineDay {
  date: string;
  count: number;
  items: HeatmapPoint[];
}

export interface TopCommunity {
  cluster_id?: string;
  org_id?: number;
  org_name?: string;
  cluster_name: string;
  longitude: string | number;
  latitude: string | number;
  case_count: number;
  rank: number;
  bbox?: {
    west: number;
    south: number;
    east: number;
    north: number;
  };
  radius_meters?: number;
}

export interface DrilldownCase {
  id: number;
  case_no: string;
  title: string;
  applicant_name?: string;
  respondent_name?: string;
  event_address?: string;
  latitude: number;
  longitude: number;
  status: number;
  status_name?: string;
  created_at: string;
  type_name?: string;
  org_name?: string;
}

export interface DrilldownResponse {
  total: number;
  page: number;
  pageSize: number;
  list: DrilldownCase[];
}

export interface MediatorOption {
  id: string | number;
  realName: string;
  phone?: string;
  avatar?: string;
  specialty?: string;
  orgName?: string;
  pendingCaseCount?: number;
  mediatingCaseCount?: number;
  pendingAssignCount?: number;
  isHighLoad?: boolean;
}

export interface MediatorLoadInfo {
  mediatorId: number;
  mediatorName: string;
  pendingCaseCount: number;
  mediatingCaseCount: number;
  pendingAssignCount: number;
  isHighLoad: boolean;
  loadThreshold: number;
  suggestion?: string;
}

export interface DrilldownParams extends HeatmapQueryParams {
  westLng?: number;
  southLat?: number;
  eastLng?: number;
  northLat?: number;
  centerLng?: number;
  centerLat?: number;
  radiusMeters?: number;
  gridKey?: string;
  clusterId?: string;
  page?: number;
  pageSize?: number;
}

export interface AmapConfig {
  web_key: string;
  security_code: string;
  default_city: string;
  default_lng: string;
  default_lat: string;
  default_zoom: number;
  cluster_radius: number;
  grid_level: number;
  use_spatial: boolean;
}

export interface HeatmapQueryParams {
  startTime?: string;
  endTime?: string;
  typeId?: number;
  organizationId?: number;
  useSpatial?: boolean;
}

export interface PopulationInfo {
  name: string;
  gender: number;
  genderName: string;
  age: number;
  nation: string;
  birthDate: string;
  idcard: string;
  address: string;
  phone: string;
  household: string;
  ethnicCode: string;
  issuer: string;
  validPeriod: string;
}

export const disputeService = {
  getList: (params?: DisputeListParams) => {
    return request.get<DisputeListResponse>('/dispute/list', { params });
  },

  getDetail: (id: string) => {
    return request.get<DisputeDetail>(`/dispute/${id}`);
  },

  create: (params: CreateDisputeParams) => {
    return request.post<{ id: string }>('/dispute/create', params);
  },

  update: (id: string, params: Partial<CreateDisputeParams>) => {
    return request.put(`/dispute/${id}`, params);
  },

  delete: (id: string) => {
    return request.delete(`/dispute/${id}`);
  },

  getTypes: () => {
    return request.get<DisputeTypeNode[]>('/dispute/types');
  },

  getStats: () => {
    return request.get<any>('/stats/dispute');
  },

  getTrend: (days: number = 30) => {
    return request.get<any>(`/stats/dispute/trend?days=${days}`);
  },

  getKeywordStats: (params?: { days?: number; limit?: number; typeId?: number; organizationId?: number }) => {
    return request.get<{ list: KeywordStatItem[]; total_cases: number; unique_kws: number; days: number }>(
      '/stats/keywords/aggregation',
      { params },
    );
  },

  getKeywordDict: (params?: { category?: string; limit?: number }) => {
    return request.get<{ keyword: string; category: string; frequency: number; status: number }[]>(
      '/ai/keywords/dict',
      { params },
    );
  },

  getHotKeywords: (params?: { days?: number; limit?: number }) => {
    return request.get<{ keyword: string; category: string; frequency: number; status: number }[]>(
      '/ai/keywords/hot',
      { params },
    );
  },

  assignMediator: (id: string, mediatorId: string) => {
    return request.post(`/dispute/${id}/assign`, { mediatorId });
  },

  urge: (id: string, reason?: string) => {
    return request.post(`/dispute/${id}/urge`, { reason });
  },

  getHeatmapData: (params?: HeatmapQueryParams) => {
    return request.get<HeatmapPoint[]>('/stats/heatmap', { params });
  },

  getHeatmapTimeline: (params?: HeatmapQueryParams) => {
    return request.get<HeatmapTimelineDay[]>('/stats/heatmap/timeline', { params });
  },

  getTopCommunities: (params?: HeatmapQueryParams & { limit?: number }) => {
    return request.get<TopCommunity[]>('/stats/heatmap/top-communities', { params });
  },

  getHeatmapDrilldown: (params?: DrilldownParams) => {
    return request.get<DrilldownResponse>('/stats/heatmap/drilldown', { params });
  },

  getAmapConfig: () => {
    return request.get<AmapConfig>('/stats/heatmap/amap-config');
  },

  extractKeywords: (text: string, title?: string, typeId?: number) => {
    return request.post<KeywordExtractResult>('/ai/keywords/extract', {
      text,
      title: title || '',
      typeId: typeId || 0,
      maxKeywords: 8,
    });
  },

  getMediatorsForAssign: (specialty?: string) => {
    const params: any = {};
    if (specialty) params.specialty = specialty;
    return request.get<MediatorOption[]>('/dispute/mediators', { params });
  },

  getMediatorLoad: (mediatorId: string | number) => {
    return request.get<MediatorLoadInfo>(`/dispute/mediators/${mediatorId}/load`);
  },

  queryPopulationByIDCard: (idCard: string) => {
    return request.post<PopulationInfo>('/v1/idcard/query', { idCard });
  },
};
