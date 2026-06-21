import { request } from '../utils/request';

export interface MediationTemplateCategory {
  code: string;
  name: string;
  count: number;
}

export interface MediationTemplate {
  id: string;
  templateName: string;
  templateCode: string;
  category: string;
  categoryName?: string;
  disputeTypeIds?: string;
  recordType: number;
  mediationPlace?: string;
  processContentTemplate: string;
  disputeFocusTemplate: string;
  mediationOpinionTemplate: string;
  agreementContentTemplate: string;
  nextStepTemplate?: string;
  defaultDuration: number;
  participantsTemplate?: string;
  tips?: string;
  isSystem: boolean;
  useCount: number;
  sortOrder: number;
  status: number;
  creatorName?: string;
  orgName?: string;
  createdAt?: string;
}

export interface MediationTemplateListParams {
  page?: number;
  pageSize?: number;
  category?: string;
  status?: number;
  keyword?: string;
  isSystem?: number;
}

export interface MediationTemplateListResponse {
  list: MediationTemplate[];
  total: number;
  page: number;
  pageSize: number;
}

export interface CreateMediationTemplateParams {
  templateName: string;
  templateCode: string;
  category: string;
  disputeTypeIds?: string;
  recordType?: number;
  mediationPlace?: string;
  processContentTemplate: string;
  disputeFocusTemplate?: string;
  mediationOpinionTemplate?: string;
  agreementContentTemplate?: string;
  nextStepTemplate?: string;
  defaultDuration?: number;
  participantsTemplate?: string;
  tips?: string;
  sortOrder?: number;
}

export interface ApplyMediationTemplateRequest {
  caseId: string;
}

export interface ApplyMediationTemplateResponse {
  recordId: string;
  templateId: string;
  templateName: string;
  recordType: number;
  recordTypeName: string;
  mediationPlace: string;
  mediationDuration: number;
  mediationTime: string;
  mediatorId: string;
  mediatorName: string;
  participantNames: string;
  processContent: string;
  disputeFocus: string;
  mediationOpinion: string;
  agreementContent: string;
  nextStep: string;
  tips: string;
  isDraft: boolean;
  tip: string;
}

export const mediationTemplateService = {
  getCategories: () => {
    return request.get<MediationTemplateCategory[]>('/dispute/mediation-template/categories');
  },

  getList: (params?: MediationTemplateListParams) => {
    return request.get<MediationTemplateListResponse>('/dispute/mediation-template', { params });
  },

  getRecommend: (caseId?: string) => {
    const params: any = {};
    if (caseId) params.caseId = caseId;
    return request.get<MediationTemplate[]>('/dispute/mediation-template/recommend', { params });
  },

  getDetail: (id: string) => {
    return request.get<MediationTemplate>(`/dispute/mediation-template/${id}`);
  },

  create: (params: CreateMediationTemplateParams) => {
    return request.post<{ id: string }>('/dispute/mediation-template', params);
  },

  update: (id: string, params: Partial<CreateMediationTemplateParams> & { status?: number }) => {
    return request.put(`/dispute/mediation-template/${id}`, params);
  },

  delete: (id: string) => {
    return request.delete(`/dispute/mediation-template/${id}`);
  },

  apply: (templateId: string, params: ApplyMediationTemplateRequest) => {
    return request.post<ApplyMediationTemplateResponse>(
      `/dispute/mediation-template/${templateId}/apply`,
      params,
    );
  },
};
