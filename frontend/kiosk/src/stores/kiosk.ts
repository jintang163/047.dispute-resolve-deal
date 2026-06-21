import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface IdCardInfo {
  name: string
  gender: string
  nation: string
  birthDate: string
  age: number | string
  address: string
  idNumber: string
  phone: string
  household: string
  issuer: string
  validPeriod: string
  photo?: string
}

export interface DisputeTypeNode {
  id: string
  name: string
  children?: DisputeTypeNode[]
}

export interface CaseDraft {
  disputeTypePath: DisputeTypeNode[]
  opponentName: string
  opponentPhone: string
  opponentAddress: string
  description: string
  expectedResolution: string
  evidenceList: EvidenceItem[]
}

export interface EvidenceItem {
  id: string
  name: string
  type: 'image' | 'document' | 'video' | 'audio'
  url: string
  size: number
  uploadTime: string
}

export const useKioskStore = defineStore('kiosk', () => {
  const currentStep = ref<number>(0)
  const totalSteps = 6

  const stepLabels = [
    '首页',
    '身份证登记',
    '选择纠纷类型',
    '填写纠纷信息',
    '上传证据材料',
    '信息确认',
    '登记完成'
  ]

  const idCardInfo = ref<IdCardInfo>({
    name: '',
    gender: '',
    nation: '',
    birthDate: '',
    age: '',
    address: '',
    idNumber: '',
    phone: '',
    household: '',
    issuer: '',
    validPeriod: ''
  })

  const caseDraft = ref<CaseDraft>({
    disputeTypePath: [],
    opponentName: '',
    opponentPhone: '',
    opponentAddress: '',
    description: '',
    expectedResolution: '',
    evidenceList: []
  })

  const caseNumber = ref<string>('')
  const createdAt = ref<string>('')

  const currentStepLabel = computed(() => stepLabels[currentStep.value] || '')
  const progressPercent = computed(() => 
    Math.round((currentStep.value / totalSteps) * 100)
  )

  function setIdCardInfo(info: Partial<IdCardInfo>) {
    idCardInfo.value = { ...idCardInfo.value, ...info }
  }

  function clearIdCardInfo() {
    idCardInfo.value = {
      name: '',
      gender: '',
      nation: '',
      birthDate: '',
      age: '',
      address: '',
      idNumber: '',
      phone: '',
      household: '',
      issuer: '',
      validPeriod: ''
    }
  }

  function setDisputeTypePath(path: DisputeTypeNode[]) {
    caseDraft.value.disputeTypePath = [...path]
  }

  function setCaseInfo(info: Partial<CaseDraft>) {
    caseDraft.value = { ...caseDraft.value, ...info }
  }

  function addEvidence(item: EvidenceItem) {
    caseDraft.value.evidenceList.push(item)
  }

  function removeEvidence(id: string) {
    const index = caseDraft.value.evidenceList.findIndex(e => e.id === id)
    if (index > -1) {
      caseDraft.value.evidenceList.splice(index, 1)
    }
  }

  function setCaseNumber(number: string) {
    caseNumber.value = number
    createdAt.value = new Date().toLocaleString('zh-CN')
  }

  function resetAll() {
    currentStep.value = 0
    clearIdCardInfo()
    caseDraft.value = {
      disputeTypePath: [],
      opponentName: '',
      opponentPhone: '',
      opponentAddress: '',
      description: '',
      expectedResolution: '',
      evidenceList: []
    }
    caseNumber.value = ''
    createdAt.value = ''
  }

  function canProceedToStep(step: number): boolean {
    if (step <= 0) return true
    if (step === 1) return true
    if (step === 2) return !!idCardInfo.value.name && !!idCardInfo.value.idNumber
    if (step === 3) return caseDraft.value.disputeTypePath.length > 0
    if (step === 4) return !!caseDraft.value.description
    if (step === 5) return true
    if (step === 6) return !!caseNumber.value
    return false
  }

  return {
    currentStep,
    totalSteps,
    stepLabels,
    idCardInfo,
    caseDraft,
    caseNumber,
    createdAt,
    currentStepLabel,
    progressPercent,
    setIdCardInfo,
    clearIdCardInfo,
    setDisputeTypePath,
    setCaseInfo,
    addEvidence,
    removeEvidence,
    setCaseNumber,
    resetAll,
    canProceedToStep
  }
})
