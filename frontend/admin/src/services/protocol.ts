import { request } from '../utils/request';

export interface GenerateProtocolParams {
  caseId: number;
  caseNo?: string;
  caseTitle?: string;
  disputeType?: string;
  partyAName: string;
  partyAGender?: string;
  partyAIDCard?: string;
  partyAAddress?: string;
  partyAPhone?: string;
  partyBName: string;
  partyBGender?: string;
  partyBIDCard?: string;
  partyBAddress?: string;
  partyBPhone?: string;
  disputeSummary: string;
  liabilityParty?: string;
  liabilityRatioA?: number;
  liabilityRatioB?: number;
  liabilityReason?: string;
  compensationAmount: number;
  compensationType?: string;
  paymentMethod?: string;
  performanceDate: string;
  paymentAccount?: string;
  otherTerms?: string[];
  breachClause?: string;
  signPlace?: string;
  signDate?: string;
  regionPrefix?: string;
  protocolYear?: number;
  protocolSeq?: number;
}

export interface MediationProtocol {
  id: string;
  caseId: number;
  protocolNo: string;
  title: string;
  content: string;
  partyAName: string;
  partyBName: string;
  mediatorName: string;
  agreementItems: string;
  breachClause: string;
  legalBasis?: string[];
  isSigned: number;
  isAIGenerated: number;
  aiGeneratedAt?: string;
  isAdopted: number;
  adoptedBy?: number;
  adoptedAt?: string;
  fileUrl?: string;
  createdAt?: string;
  generatedAt?: string;
  esignFlowId?: string;
  bcCertNo?: string;
  bcTxId?: string;
  bcOnChainTime?: string;
}

export const protocolService = {
  generate: (caseId: number | string, params: GenerateProtocolParams) => {
    return request.post<{
      protocolId: string;
      protocolNo: string;
      title: string;
      content: string;
      partyAName: string;
      partyBName: string;
      mediatorName: string;
      agreementItems: string;
      breachClause: string;
      legalBasis: string[];
      generatedAt: string;
    }>(`/dispute/${caseId}/mediation/protocol/generate`, params);
  },

  list: (caseId: number | string) => {
    return request.get<MediationProtocol[]>(`/dispute/${caseId}/mediation/protocols`);
  },

  adopt: (caseId: number | string, protocolId: number | string) => {
    return request.post<void>(`/dispute/${caseId}/mediation/protocol/${protocolId}/adopt`);
  },
};
