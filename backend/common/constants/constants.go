package constants

const (
	RoleDirector    = 1
	RoleLeader      = 2
	RoleMediator    = 3
	RoleAdmin       = 4

	RoleCodeDirector = "ROLE_DIRECTOR"
	RoleCodeLeader   = "ROLE_LEADER"
	RoleCodeMediator = "ROLE_MEDIATOR"
	RoleCodeAdmin    = "ROLE_ADMIN"

	OrgTypeCenter   = 1
	OrgTypeStreet   = 2
	OrgTypeCommunity = 3
	OrgTypeVillage  = 4

	CaseStatusPending    = 10
	CaseStatusMediating  = 20
	CaseStatusWaiting    = 30
	CaseStatusApproving  = 40
	CaseStatusClosed     = 50
	CaseStatusCancelled  = 99

	CaseLevelExtraUrgent = 1
	CaseLevelUrgent      = 2
	CaseLevelNormal      = 3
	CaseLevelCommon      = 4

	CaseSourceKiosk   = 1
	CaseSourceMiniApp = 2
	CaseSourcePhone   = 3
	CaseSourceWindow  = 4
	CaseSourceTransfer = 5

	MediationResultPending    = 0
	MediationResultSuccess    = 1
	MediationResultFail       = 2
	MediationResultPartial    = 3

	ApprovalStatusPending    = 10
	ApprovalStatusPassed     = 20
	ApprovalStatusRejected   = 30
	ApprovalStatusCancelled  = 40

	ApprovalActionPass       = 1
	ApprovalActionReject     = 2
	ApprovalActionReturn     = 3
	ApprovalActionAddSign    = 4
	ApprovalActionTransfer   = 5
	ApprovalActionRefuse     = 6

	UrgeTypeUser     = 1
	UrgeTypeLeader   = 2
	UrgeTypeSystem   = 3
	UrgeTypeEscalate = 4

	FileTypeImage    = 1
	FileTypeVideo    = 2
	FileTypeAudio    = 3
	FileTypeDocument = 4
	FileTypeOther    = 5

	EvidenceCategoryUncategorized = 0
	EvidenceCategoryIDCard        = 1
	EvidenceCategoryContract      = 2
	EvidenceCategoryReceipt       = 3
	EvidenceCategoryPhoto         = 4
	EvidenceCategoryChatRecord    = 5
	EvidenceCategoryMedia         = 6
	EvidenceCategoryInvoice       = 7
	EvidenceCategoryCertificate   = 8
	EvidenceCategoryOther         = 9

	AIEvidenceProcessNotStart = 0
	AIEvidenceProcessing     = 1
	AIEvidenceProcessDone    = 2
	AIEvidenceProcessFailed  = 3

	VideoStatusNotStarted = 10
	VideoStatusRunning    = 20
	VideoStatusEnded      = 30
	VideoStatusCancelled  = 40

	VideoRecordStatusIdle     = 0
	VideoRecordStatusRecording = 1
	VideoRecordStatusStopped  = 2
	VideoRecordStatusFailed   = 3

	VideoQueueStatusWaiting  = 1
	VideoQueueStatusEntered  = 2
	VideoQueueStatusCancelled = 3
	VideoQueueStatusTimeout  = 4

	VideoMinutesStatusGenerated = 1
	VideoMinutesStatusApproved  = 2
	VideoMinutesStatusRevoked   = 3

	KioskStatusOffline = 0
	KioskStatusOnline  = 1
	KioskStatusFault   = 2

	RedisKeyPrefixUser      = "user:"
	RedisKeyPrefixToken     = "token:"
	RedisKeyPrefixCase      = "case:"
	RedisKeyPrefixLock      = "lock:"
	RedisKeyPrefixRateLimit = "ratelimit:"
	RedisKeyPrefixQueue     = "queue:"
	RedisKeyPrefixJudicial  = "judicial:"
	RedisKeyPrefixVideo     = "video:"
	RedisKeyPrefixBC        = "bc:"
	RedisKeyPrefixEsign     = "esign:"

	RedisExpireToken     = 86400
	RedisExpireUser      = 3600
	RedisExpireCase      = 1800
	RedisExpireLock      = 300
	RedisExpireRateLimit = 60

	MQTopicCaseCreate     = "dispute_case_create"
	MQTopicCaseAssign     = "dispute_case_assign"
	MQTopicCaseUrge       = "dispute_case_urge"
	MQTopicCaseStatus     = "dispute_case_status"
	MQTopicApprovalNotify = "dispute_approval_notify"
	MQTopicNotification   = "dispute_notification"
	MQTopicDeadLetter     = "dispute_dead_letter"
	MQTopicAIProcess      = "dispute_ai_process"
	MQTopicJudicialSubmit = "judicial_confirmation_submit"
	MQTopicJudicialStatus = "judicial_confirmation_status"
	MQTopicJudicialSync   = "judicial_confirmation_sync"
	MQTopicJudicialSeal   = "judicial_confirmation_seal"
	MQTopicJudicialRemind = "judicial_confirmation_remind"
	MQTopicVideoRecord   = "dispute_video_record"
	MQTopicVideoQueue    = "dispute_video_queue"
	MQTopicEsignNotify   = "dispute_esign_notify"
	MQTopicBlockchainStore = "dispute_blockchain_store"
	MQTopicTransferCreate  = "dispute_transfer_create"
	MQTopicTransferReceive = "dispute_transfer_receive"
	MQTopicTransferUrge    = "dispute_transfer_urge"
	MQTopicTransferTimeout = "dispute_transfer_timeout"
	MQTopicTransferComplete = "dispute_transfer_complete"
	MQTopicEvidenceClassify = "dispute_evidence_classify"
	MQTopicExportPassword   = "data_export_password"

	MQTagAll     = "*"
	MQTagSms     = "sms"
	MQTagWechat  = "wechat"
	MQTagApp     = "app"
	MQTagEmail   = "email"

	ESIndexCase     = "dispute_case"
	ESIndexRecord   = "dispute_mediation_record"
	ESIndexLog      = "operation_log"

	MinIOPathEvidence    = "evidence"
	MinIOPathVideo       = "video"
	MinIOPathDocument    = "document"
	MinIOPathSignature   = "signature"
	MinIOPathAvatar      = "avatar"
	MinIOPathJudicial    = "judicial"
	MinIOPathCallback    = "callback"
	MinIOPathExport      = "export"

	ExportTypeCase        = 1
	ExportTypePerformance = 2
	ExportTypeEvidence    = 3
	ExportTypeOther       = 4

	ExportStatusProcessing = 10
	ExportStatusSuccess    = 20
	ExportStatusFailed     = 30

	ExportPasswordSmsStatusPending = 0
	ExportPasswordSmsStatusSent    = 1
	ExportPasswordSmsStatusFailed  = 2

	ExportAESKeyLength  = 32
	ExportExpireDays    = 7

	EsignStatusDraft     = 0
	EsignStatusPending   = 10
	EsignStatusSigning   = 20
	EsignStatusCompleted = 30
	EsignStatusExpired   = 40
	EsignStatusRevoked   = 50

	EsignSignerStatusPending  = 0
	EsignSignerStatusSigned   = 1
	EsignSignerStatusRejected = 2

	EsignNotifyStatusNone     = 0
	EsignNotifyStatusSMS      = 1
	EsignNotifyStatusWechat   = 2
	EsignNotifyStatusAll      = 3

	BCStatusPending    = 0
	BCStatusOnChain    = 1
	BCStatusFailed     = 2
	BCStatusVerified   = 3

	BCTypeMediationProtocol = "mediation_protocol"
	BCTypeEsignDocument     = "esign_document"
	BCTypeEvidence          = "evidence"

	AITypeSummary   = 1
	AITypeSuggestion = 2
	AITypeRisk      = 3

	SummaryTypeMediation = 1
	SummaryTypeApproval  = 2
	SummaryTypeRisk      = 3

	PerformancePeriodMonth    = 1
	PerformancePeriodQuarter  = 2
	PerformancePeriodYear     = 3

	TransferStatusPending     = 10
	TransferStatusReceived    = 20
	TransferStatusProcessing  = 30
	TransferStatusCompleted   = 40
	TransferStatusRejected    = 50
	TransferStatusCancelled   = 99

	TransferUrgeTypeUser     = 1
	TransferUrgeTypeLeader   = 2
	TransferUrgeTypeSystem   = 3

	CaseLibraryStatusDisabled  = 0
	CaseLibraryStatusActive    = 1
	CaseLibraryStatusArchived  = 2

	CaseLibraryVectorNotStart = 0
	CaseLibraryVectorProcessing = 1
	CaseLibraryVectorDone     = 2
	CaseLibraryVectorFailed   = 3

	CaseLibraryDifficultySimple  = 1
	CaseLibraryDifficultyNormal  = 2
	CaseLibraryDifficultyComplex = 3
	CaseLibraryDifficultyHard    = 4

	CaseLibraryQuoteTypeTactics  = 1
	CaseLibraryQuoteTypeStrategy = 2
	CaseLibraryQuoteTypeFull     = 3

	CaseLibraryArchiveReasonUnused     = 1
	CaseLibraryArchiveReasonManual     = 2
	CaseLibraryArchiveReasonLowScore   = 3

	CaseLibraryArchiveMonths = 12

	TransferUrgeSourceSystem = 1
	TransferUrgeSourceManual = 2

	LegalAidOrgTypeCenter      = 1
	LegalAidOrgTypeLawFirm     = 2
	LegalAidOrgTypeNotary      = 3
	LegalAidOrgTypeForensic    = 4

	LegalAidOrgLevelCity    = 1
	LegalAidOrgLevelDistrict = 2
	LegalAidOrgLevelStreet   = 3

	LegalAidApplyStatusPending  = 10
	LegalAidApplyStatusApproved = 20
	LegalAidApplyStatusRejected = 30
	LegalAidApplyStatusCanceled = 40

	LegalAidIncomeLevelLowIncome  = 1
	LegalAidIncomeLevelLow        = 2
	LegalAidIncomeLevelNormal     = 3

	LegalAidTransferStatusPending  = 10
	LegalAidTransferStatusAccepted = 20
	LegalAidTransferStatusRejected = 30
	LegalAidTransferStatusClosed   = 40

	LegalAidConsultTypeText   = 1
	LegalAidConsultTypeVoice  = 2
	LegalAidConsultTypeVideo  = 3

	LegalAidConsultStatusPending   = 10
	LegalAidConsultStatusOngoing   = 20
	LegalAidConsultStatusCompleted = 30
	LegalAidConsultStatusCanceled  = 40
	LegalAidConsultStatusTimeout   = 50

	LegalAidSenderTypeUser   = 1
	LegalAidSenderTypeLawyer = 2

	LegalAidMessageTypeText  = 1
	LegalAidMessageTypeImage = 2
	LegalAidMessageTypeVoice = 3
	LegalAidMessageTypeVideo = 4
	LegalAidMessageTypeFile  = 5

	LegalAidFreeConsultDuration = 1800

	RedisKeyPrefixLegalAid = "legalaid:"
	MinIOPathLegalAid      = "legalaid"

	MediationTemplateCategoryNeighborhood = "neighborhood"
	MediationTemplateCategoryWage         = "wage"
	MediationTemplateCategoryProperty     = "property"
	MediationTemplateCategoryContract     = "contract"
	MediationTemplateCategoryFamily       = "family"
	MediationTemplateCategoryTraffic      = "traffic"
	MediationTemplateCategoryConsumer     = "consumer"
	MediationTemplateCategoryOther        = "other"

	CounselorStatusDisabled = 0
	CounselorStatusEnabled  = 1

	CounselorConsultTypeVideo  = 1
	CounselorConsultTypeVoice  = 2
	CounselorConsultTypeOffline = 3

	CounselorAppointmentStatusPending    = 10
	CounselorAppointmentStatusConfirmed  = 20
	CounselorAppointmentStatusOngoing    = 30
	CounselorAppointmentStatusCompleted  = 40
	CounselorAppointmentStatusCancelled  = 50
	CounselorAppointmentStatusExpired    = 60

	CounselorAppointmentSourceAdmin    = 1
	CounselorAppointmentSourceSelf     = 2
	CounselorAppointmentSourceSystem   = 3

	CounselorEmergencyLevelNormal   = 0
	CounselorEmergencyLevelAttention = 1
	CounselorEmergencyLevelUrgent    = 2
	CounselorEmergencyLevelHighRisk  = 3

	CounselorScheduleTypeRest       = 1
	CounselorScheduleTypeAppointment = 2
	CounselorScheduleTypeOther      = 3

	RedisKeyPrefixCounselor = "counselor:"
	RedisKeyPrefixAppointment = "appointment:"

	MinIOPathCounselor = "counselor"
)

var MediationTemplateCategoryMap = map[string]string{
	MediationTemplateCategoryNeighborhood: "邻里纠纷",
	MediationTemplateCategoryWage:         "欠薪纠纷",
	MediationTemplateCategoryProperty:     "物业纠纷",
	MediationTemplateCategoryContract:     "合同纠纷",
	MediationTemplateCategoryFamily:       "家庭纠纷",
	MediationTemplateCategoryTraffic:      "交通事故",
	MediationTemplateCategoryConsumer:     "消费纠纷",
	MediationTemplateCategoryOther:        "其他",
}

var CaseStatusMap = map[int]string{
	CaseStatusPending:   "待分派",
	CaseStatusMediating: "调解中",
	CaseStatusWaiting:   "待审批",
	CaseStatusApproving: "审批中",
	CaseStatusClosed:    "已结案",
	CaseStatusCancelled: "已取消",
}

var CaseLevelMap = map[int]string{
	CaseLevelExtraUrgent: "特急",
	CaseLevelUrgent:      "紧急",
	CaseLevelNormal:      "一般",
	CaseLevelCommon:      "普通",
}

var RoleMap = map[int]string{
	RoleDirector: "主任",
	RoleLeader:   "组长",
	RoleMediator: "调解员",
	RoleAdmin:    "管理员",
}

var ApprovalActionMap = map[int]string{
	ApprovalActionPass:     "通过",
	ApprovalActionReject:   "驳回",
	ApprovalActionReturn:   "退回修改",
	ApprovalActionAddSign:  "加签",
	ApprovalActionTransfer: "转审",
	ApprovalActionRefuse:   "拒绝",
}

var JudicialStatusMap = map[int]string{
	10: "已提交",
	20: "审核中",
	30: "已确认",
	40: "已驳回",
	50: "已失效",
}

var JudicialActionTypeMap = map[int]string{
	10: "提交申请",
	20: "法院受理",
	30: "审核通过",
	40: "审核驳回",
	50: "已签章",
	60: "确认书送达",
	70: "履行提醒",
	80: "失效提醒",
	90: "已履行",
	99: "已失效",
}

var TransferStatusMap = map[int]string{
	TransferStatusPending:    "待接收",
	TransferStatusReceived:   "已接收",
	TransferStatusProcessing: "处理中",
	TransferStatusCompleted:  "已办结",
	TransferStatusRejected:   "已驳回",
	TransferStatusCancelled:  "已取消",
}

var TransferDeptTypeMap = map[string]string{
	"HR": "人社局",
	"MS": "市监局",
	"PS": "公安局",
	"CT": "法院",
	"OT": "其他部门",
}

var CaseLibraryStatusMap = map[int]string{
	CaseLibraryStatusDisabled: "禁用",
	CaseLibraryStatusActive:   "启用",
	CaseLibraryStatusArchived: "已归档",
}

var CaseLibraryDifficultyMap = map[int]string{
	CaseLibraryDifficultySimple:  "简单",
	CaseLibraryDifficultyNormal:  "一般",
	CaseLibraryDifficultyComplex: "复杂",
	CaseLibraryDifficultyHard:    "疑难",
}

var CaseLibraryArchiveReasonMap = map[int]string{
	CaseLibraryArchiveReasonUnused:   "超1年未使用",
	CaseLibraryArchiveReasonManual:   "手动归档",
	CaseLibraryArchiveReasonLowScore: "评分过低",
}

var EvidenceCategoryMap = map[int]string{
	EvidenceCategoryUncategorized: "未分类",
	EvidenceCategoryIDCard:        "身份证件",
	EvidenceCategoryContract:      "合同协议",
	EvidenceCategoryReceipt:       "收据凭证",
	EvidenceCategoryPhoto:         "现场照片",
	EvidenceCategoryChatRecord:    "聊天记录",
	EvidenceCategoryMedia:         "录音录像",
	EvidenceCategoryInvoice:       "发票票据",
	EvidenceCategoryCertificate:   "证件证明",
	EvidenceCategoryOther:         "其他材料",
}

var LegalAidOrgTypeMap = map[int]string{
	LegalAidOrgTypeCenter:   "法律援助中心",
	LegalAidOrgTypeLawFirm:  "律师事务所",
	LegalAidOrgTypeNotary:   "公证处",
	LegalAidOrgTypeForensic: "司法鉴定所",
}

var LegalAidOrgLevelMap = map[int]string{
	LegalAidOrgLevelCity:     "市级",
	LegalAidOrgLevelDistrict: "区级",
	LegalAidOrgLevelStreet:   "街道级",
}

var LegalAidApplyStatusMap = map[int]string{
	LegalAidApplyStatusPending:  "待审核",
	LegalAidApplyStatusApproved: "审核通过",
	LegalAidApplyStatusRejected: "审核驳回",
	LegalAidApplyStatusCanceled: "已撤销",
}

var LegalAidIncomeLevelMap = map[int]string{
	LegalAidIncomeLevelLowIncome: "低保",
	LegalAidIncomeLevelLow:       "低收入",
	LegalAidIncomeLevelNormal:    "普通",
}

var LegalAidTransferStatusMap = map[int]string{
	LegalAidTransferStatusPending:  "待受理",
	LegalAidTransferStatusAccepted: "已受理",
	LegalAidTransferStatusRejected: "已驳回",
	LegalAidTransferStatusClosed:   "已办结",
}

var LegalAidConsultTypeMap = map[int]string{
	LegalAidConsultTypeText:  "文字咨询",
	LegalAidConsultTypeVoice: "语音咨询",
	LegalAidConsultTypeVideo: "视频咨询",
}

var LegalAidConsultStatusMap = map[int]string{
	LegalAidConsultStatusPending:   "待开始",
	LegalAidConsultStatusOngoing:   "进行中",
	LegalAidConsultStatusCompleted: "已完成",
	LegalAidConsultStatusCanceled:  "已取消",
	LegalAidConsultStatusTimeout:   "已超时",
}

var CounselorStatusMap = map[int]string{
	CounselorStatusDisabled: "停用",
	CounselorStatusEnabled:  "启用",
}

var CounselorConsultTypeMap = map[int]string{
	CounselorConsultTypeVideo:   "线上视频",
	CounselorConsultTypeVoice:   "线上语音",
	CounselorConsultTypeOffline: "线下面谈",
}

var CounselorAppointmentStatusMap = map[int]string{
	CounselorAppointmentStatusPending:   "待确认",
	CounselorAppointmentStatusConfirmed: "已确认",
	CounselorAppointmentStatusOngoing:   "咨询中",
	CounselorAppointmentStatusCompleted: "已完成",
	CounselorAppointmentStatusCancelled: "已取消",
	CounselorAppointmentStatusExpired:   "已过期",
}

var CounselorAppointmentSourceMap = map[int]string{
	CounselorAppointmentSourceAdmin:  "管理员创建",
	CounselorAppointmentSourceSelf:   "自助预约",
	CounselorAppointmentSourceSystem: "系统推荐",
}

var CounselorEmergencyLevelMap = map[int]string{
	CounselorEmergencyLevelNormal:    "普通",
	CounselorEmergencyLevelAttention: "关注",
	CounselorEmergencyLevelUrgent:    "紧急",
	CounselorEmergencyLevelHighRisk:  "高危",
}

var CounselorScheduleTypeMap = map[int]string{
	CounselorScheduleTypeRest:       "休息",
	CounselorScheduleTypeAppointment: "已有预约",
	CounselorScheduleTypeOther:      "其他事务",
}

var ExportTypeMap = map[int]string{
	ExportTypeCase:        "案件数据",
	ExportTypePerformance: "考核数据",
	ExportTypeEvidence:    "证据数据",
	ExportTypeOther:       "其他",
}

var ExportStatusMap = map[int]string{
	ExportStatusProcessing: "导出中",
	ExportStatusSuccess:    "导出成功",
	ExportStatusFailed:   "导出失败",
}

var ExportPasswordSmsStatusMap = map[int]string{
	ExportPasswordSmsStatusPending: "未发送",
	ExportPasswordSmsStatusSent:    "已发送",
	ExportPasswordSmsStatusFailed:  "发送失败",
}

