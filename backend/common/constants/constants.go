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
)

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
