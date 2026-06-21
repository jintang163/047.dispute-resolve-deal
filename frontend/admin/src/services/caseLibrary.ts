import { request } from '../utils/request';

export interface CaseLibraryItem {
  id: number;
  caseNo: string;
  title: string;
  description: string;
  disputeType: string;
  typeId: number;
  mediationTactics: string;
  keyPoints: string;
  resultSummary: string;
  difficultyLevel: number;
  isSuccess: number;
  mediatorName: string;
  mediatorId: number;
  orgName: string;
  orgId: number;
  sourceCaseId: number;
  keywords: string;
  tags: string;
  vectorId: string;
  vectorStatus: number;
  referenceCount: number;
  avgScore: number;
  scoreCount: number;
  totalScore: number;
  lastUsedAt: string;
  status: number;
  archivedAt: string;
  createdBy: number;
  createdAt: string;
  updatedAt: string;
}

export interface CaseSearchResult {
  id: number;
  score: number;
  caseId: number;
  title: string;
  description: string;
  disputeType: string;
  mediationTactics: string;
  keyPoints: string;
  keywords: string;
  difficultyLevel: number;
  isSuccess: number;
  distance: number;
}

export interface CaseLibraryArchive {
  id: number;
  originalId: number;
  caseNo: string;
  title: string;
  archiveReason: number;
  avgScore: number;
  referenceCount: number;
  lastUsedAt: string;
  archivedBy: number;
  caseData: string;
  createdAt: string;
}

export interface CaseLibraryQuote {
  id: number;
  sourceCaseId: number;
  libraryCaseId: number;
  libraryCaseNo: string;
  quoteType: number;
  quoteContent: string;
  userId: number;
  userName: string;
  mediationRecordId: number;
  createdAt: string;
}

export interface CaseListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  disputeType?: string;
  difficultyLevel?: number;
  status?: number;
}

export const caseLibraryService = {
  getList: (params: CaseListParams) =>
    request.get('/case-library', { params }),

  getDetail: (id: number) =>
    request.get(`/case-library/${id}`),

  create: (data: Partial<CaseLibraryItem>) =>
    request.post('/case-library', data),

  update: (id: number, data: Partial<CaseLibraryItem>) =>
    request.put(`/case-library/${id}`, data),

  delete: (id: number) =>
    request.delete(`/case-library/${id}`),

  searchSimilar: (query?: string, topK: number = 5, caseId?: number) =>
    request.post('/case-library/search', { query, topK, caseId }),

  score: (id: number, score: number, sourceCaseId?: number, comment?: string) =>
    request.post(`/case-library/${id}/score`, { score, sourceCaseId, comment }),

  quote: (id: number, sourceCaseId: number, quoteType: number = 1, quoteContent?: string, mediationRecordId?: number) =>
    request.post(`/case-library/${id}/quote`, { sourceCaseId, quoteType, quoteContent, mediationRecordId }),

  getQuotes: (sourceCaseId: number) =>
    request.get('/case-library/quotes', { params: { sourceCaseId } }),

  archive: (id: number, reason: number = 2) =>
    request.post(`/case-library/${id}/archive`, { reason }),

  restore: (id: number) =>
    request.post(`/case-library/${id}/restore`),

  vectorize: (id: number) =>
    request.post(`/case-library/${id}/vectorize`),

  vectorizeAll: () =>
    request.post('/case-library/vectorize-all'),

  getArchiveList: (params: { page?: number; pageSize?: number }) =>
    request.get('/case-library/archives', { params }),
};
