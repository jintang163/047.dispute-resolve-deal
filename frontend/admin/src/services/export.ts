import { request } from '../utils/request';

export interface DataExportLog {
  id: string | number;
  exportNo: string;
  exportType: number;
  exportTypeName?: string;
  exportName: string;
  filterConditions?: string;
  filterConditionsObj?: any;
  recordCount: number;
  fileName?: string;
  filePath?: string;
  fileSize: number;
  encryptionAlgorithm?: string;
  passwordSmsSent: number;
  passwordSmsSentName?: string;
  exportStatus: number;
  exportStatusName?: string;
  errorMessage?: string;
  operatorId: string | number;
  operatorName: string;
  operatorPhone?: string;
  orgId?: string | number;
  orgName?: string;
  ipAddress?: string;
  completedAt?: string;
  expiredAt?: string;
  createdAt?: string;
}

export interface CaseExportParams {
  typeId?: number | string;
  mediatorId?: number | string;
  status?: number;
  caseLevel?: number;
  startTime?: string;
  endTime?: string;
  keyword?: string;
  tagKeyword?: string;
  ids?: Array<number | string>;
}

export interface ExportListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  exportType?: number;
  exportStatus?: number;
  startTime?: string;
  endTime?: string;
}

export interface CaseExportResult {
  exportId: string | number;
  exportNo: string;
  recordCount: number;
  message?: string;
}

export const exportService = {
  createCaseExport: (params: CaseExportParams) => {
    return request.post<CaseExportResult>('/v1/dispute/export', params);
  },

  getExportList: (params?: ExportListParams) => {
    return request.get<{ list: DataExportLog[]; total: number; page: number; pageSize: number }>(
      '/v1/export/log',
      { params },
    );
  },

  getExportDetail: (id: string | number) => {
    return request.get<DataExportLog>(`/v1/export/log/${id}`);
  },

  downloadExport: (id: string | number, exportNo: string) => {
    const token = localStorage.getItem('token') || '';
    const url = `/api/v1/export/log/${id}/download`;
    return fetch(url, {
      method: 'GET',
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }).then(async (res) => {
      if (!res.ok) {
        const text = await res.text().catch(() => '下载失败');
        throw new Error(text || '下载失败');
      }
      const blob = await res.blob();
      const a = document.createElement('a');
      const href = window.URL.createObjectURL(blob);
      a.href = href;
      a.download = `${exportNo}.enc`;
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      window.URL.revokeObjectURL(href);
    });
  },
};
