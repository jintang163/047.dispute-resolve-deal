import { request } from '../utils/request';

export interface DisputeCase {
  id: string;
  caseNo: string;
  title: string;
  type: string;
  typeName?: string;
  status: string;
  statusName?: string;
  description?: string;
  address?: string;
  urgency?: string;
  partyA?: string;
  partyB?: string;
  partyAPhone?: string;
  partyBPhone?: string;
  orgId?: string;
  orgName?: string;
  mediatorId?: string;
  mediatorName?: string;
  createTime?: string;
  updateTime?: string;
  creator?: string;
  creatorName?: string;
}

export interface DisputeListParams {
  pageNum?: number;
  pageSize?: number;
  keyword?: string;
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
  type: string;
  description?: string;
  address?: string;
  urgency?: string;
  partyA: string;
  partyB: string;
  partyAPhone?: string;
  partyBPhone?: string;
  orgId?: string;
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
    return request.get<{ id: string; name: string; code: string }[]>('/dispute/types');
  },

  getStats: () => {
    return request.get<any>('/stats/dispute');
  },

  getTrend: (days: number = 30) => {
    return request.get<any>(`/stats/dispute/trend?days=${days}`);
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
};
