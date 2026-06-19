import { request } from '../utils/request';
import axios from 'axios';

export interface EsignFlow {
  id: string;
  flowId: string;
  caseId: number;
  caseNo: string;
  docTitle: string;
  docUrl: string;
  signedDocumentUrl: string;
  status: number;
  statusName: string;
  signerCount: number;
  signedCount: number;
  expireTime: string;
  fadadaFlowId: string;
  crossPageSeal: number;
  bcCertNo: string;
  bcTxId: string;
  bcOnChainTime: string;
  bcStatus: number;
  createdAt: string;
  signers: EsignSigner[];
  blockchainCert?: BlockchainCertInfo;
}

export interface EsignSigner {
  id: string;
  userId: number;
  userName: string;
  userPhone: string;
  signOrder: number;
  signStatus: number;
  signStatusName: string;
  signedAt: string;
  signatureUrl: string;
  fadadaSignUrl: string;
  notifyStatus: number;
  notifyStatusName: string;
  notifySentAt: string;
}

export interface BlockchainCertInfo {
  certNo: string;
  txId: string;
  blockHeight: number;
  onChainTime: string;
  certUrl: string;
  qrcodeUrl: string;
  verifyUrl: string;
  status: number;
  evidenceHash: string;
}

export interface CreateEsignFlowParams {
  caseId: number;
  docType: number;
  docTitle: string;
  docContent?: string;
  docUrl?: string;
  templateId?: number;
  signerIds: number[];
  expireHours?: number;
  needNotify?: boolean;
  notifyType?: string;
  crossPageSeal?: boolean;
}

export interface BlockchainCertificate {
  id: string;
  certNo: string;
  evidenceId: string;
  evidenceType: string;
  evidenceTypeName: string;
  evidenceName: string;
  evidenceHash: string;
  caseId: number;
  caseNo: string;
  caseTitle: string;
  flowId: string;
  txId: string;
  blockHeight: number;
  onChainTime: string;
  certUrl: string;
  qrcodeUrl: string;
  verifyUrl: string;
  status: number;
  statusName: string;
  metadata: string;
  createdBy: number;
  createdAt: string;
}

export const esignService = {
  createFlow: (caseId: number | string, params: CreateEsignFlowParams) =>
    request.post<EsignFlow>(`/dispute/${caseId}/esign`, params),

  getList: (caseId: number | string, params?: { status?: number }) =>
    request.get<EsignFlow[]>(`/dispute/${caseId}/esign`, { params }),

  getDetail: (caseId: number | string, flowId: string) =>
    request.get<EsignFlow>(`/dispute/${caseId}/esign/${flowId}`),

  signDocument: (caseId: number | string, flowId: string, data: { recordId: number; verifyCode?: string }) =>
    request.post(`/dispute/${caseId}/esign/${flowId}/sign`, data),

  revokeFlow: (caseId: number | string, flowId: string, reason: string) =>
    request.post(`/dispute/${caseId}/esign/${flowId}/revoke`, { reason }),

  sendVerifyCode: (caseId: number | string, flowId: string) =>
    request.post(`/dispute/${caseId}/esign/${flowId}/send-code`),

  getProgress: (caseId: number | string, flowId: string) =>
    request.get(`/dispute/${caseId}/esign/${flowId}/progress`),
};

export const blockchainService = {
  storeEvidence: (caseId: number | string, data: {
    evidenceId: string;
    evidenceType: string;
    evidenceName: string;
    evidenceHash: string;
    flowId?: string;
    metadata?: string;
  }) => request.post(`/dispute/${caseId}/blockchain/store`, { case_id: caseId, ...data }),

  getCertList: (caseId: number | string, params?: { evidenceType?: string; page?: number; pageSize?: number }) =>
    request.get<BlockchainCertificate[]>(`/dispute/${caseId}/blockchain/certs`, { params }),

  getCertDetail: (caseId: number | string, certNo: string) =>
    request.get<BlockchainCertificate>(`/dispute/${caseId}/blockchain/cert/${certNo}`),

  verifyEvidence: (caseId: number | string, certNo: string) =>
    request.get(`/dispute/${caseId}/blockchain/verify`, { params: { certNo } }),

  downloadCert: (caseId: number | string, certNo: string) =>
    request.get(`/dispute/${caseId}/blockchain/cert/${certNo}/download`),

  publicVerify: (certNo: string) =>
    axios.get(`/api/v1/public/blockchain/verify/${certNo}`).then(res => res.data),
};
