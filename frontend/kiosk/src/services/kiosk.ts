import { get, post, upload } from '@/utils/request'
import type { IdCardInfo, EvidenceItem, DisputeTypeNode } from '@/stores/kiosk'

export interface SubmitCaseParams {
  idCardInfo: IdCardInfo
  disputeTypePath: DisputeTypeNode[]
  opponentName: string
  opponentPhone: string
  opponentAddress: string
  description: string
  expectedResolution: string
  evidenceList: EvidenceItem[]
}

export interface SubmitCaseResult {
  caseNumber: string
  createdAt: string
  estimatedDays: number
  mediator: string
  mediatorPhone: string
}

export interface AIConversationMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

export const kioskApi = {
  getDisputeTypes(): Promise<DisputeTypeNode[]> {
    return get<DisputeTypeNode[]>('/dispute/types')
  },

  verifyIdCard(idNumber: string, name: string): Promise<{ valid: boolean; message?: string }> {
    return post('/idcard/verify', { idNumber, name })
  },

  submitCase(params: SubmitCaseParams): Promise<SubmitCaseResult> {
    return post('/case/submit', params)
  },

  uploadEvidence(file: File, onProgress?: (percent: number) => void): Promise<EvidenceItem> {
    return upload<EvidenceItem>('/evidence/upload', file, onProgress)
  },

  getCaseByNumber(caseNumber: string): Promise<any> {
    return get(`/case/detail/${caseNumber}`)
  },

  sendAIMessage(
    message: string,
    conversationId?: string,
    context?: { disputeType?: string }
  ): Promise<{ conversationId: string; reply: string }> {
    return post('/ai/chat', {
      message,
      conversationId,
      context
    })
  },

  getAIQuickQuestions(disputeType?: string): Promise<string[]> {
    return get('/ai/quick-questions', {
      params: { disputeType }
    })
  },

  getReceiptData(caseNumber: string): Promise<any> {
    return get(`/case/receipt/${caseNumber}`)
  }
}

export default kioskApi
