package dispute

type DisputeCase struct {
	Id               int64   `json:"id" th:"1,optional"`
	CaseNo           string  `json:"case_no" th:"2,optional"`
	Title            string  `json:"title" th:"3,optional"`
	TypeId           int64   `json:"type_id" th:"4,optional"`
	TypeName         string  `json:"type_name" th:"5,optional"`
	CaseLevel        int32   `json:"case_level" th:"6,optional"`
	Description      string  `json:"description" th:"7,optional"`
	ExpectedSolution string  `json:"expected_solution" th:"8,optional"`
	Source           int32   `json:"source" th:"9,optional"`
	ApplicantId      int64   `json:"applicant_id" th:"10,optional"`
	ApplicantName    string  `json:"applicant_name" th:"11,optional"`
	ApplicantIdcard  string  `json:"applicant_idcard" th:"12,optional"`
	ApplicantMobile  string  `json:"applicant_mobile" th:"13,optional"`
	RespondentName   string  `json:"respondent_name" th:"14,optional"`
	RespondentIdcard string  `json:"respondent_idcard" th:"15,optional"`
	RespondentMobile string  `json:"respondent_mobile" th:"16,optional"`
	Longitude        float64 `json:"longitude" th:"17,optional"`
	Latitude         float64 `json:"latitude" th:"18,optional"`
	Address          string  `json:"address" th:"19,optional"`
	OrganizationId   int64   `json:"organization_id" th:"20,optional"`
	MediatorId       int64   `json:"mediator_id" th:"21,optional"`
	MediatorName     string  `json:"mediator_name" th:"22,optional"`
	Status           int32   `json:"status" th:"23,optional"`
	UrgencyLevel     int32   `json:"urgency_level" th:"24,optional"`
	CreatedBy        int64   `json:"created_by" th:"25,optional"`
	CreatedAt        string  `json:"created_at" th:"26,optional"`
	UpdatedAt        string  `json:"updated_at" th:"27,optional"`
}

type Evidence struct {
	Id         int64  `json:"id" th:"1,optional"`
	CaseId     int64  `json:"case_id" th:"2,optional"`
	FileType   int32  `json:"file_type" th:"3,optional"`
	FileName   string `json:"file_name" th:"4,optional"`
	FileUrl    string `json:"file_url" th:"5,optional"`
	FileSize   int64  `json:"file_size" th:"6,optional"`
	Remark     string `json:"remark" th:"7,optional"`
	SortOrder  int32  `json:"sort_order" th:"8,optional"`
	UploadFrom int32  `json:"upload_from" th:"9,optional"`
	UploadedBy int64  `json:"uploaded_by" th:"10,optional"`
	CreatedAt  string `json:"created_at" th:"11,optional"`
}

type DisputeHistory struct {
	Id           int64  `json:"id" th:"1,optional"`
	CaseId       int64  `json:"case_id" th:"2,optional"`
	ActionType   int32  `json:"action_type" th:"3,optional"`
	ActionName   string `json:"action_name" th:"4,optional"`
	Remark       string `json:"remark" th:"5,optional"`
	OperatorId   int64  `json:"operator_id" th:"6,optional"`
	OperatorName string `json:"operator_name" th:"7,optional"`
	CreatedAt    string `json:"created_at" th:"8,optional"`
}

type DisputeType struct {
	Id          int64  `json:"id" th:"1,optional"`
	Name        string `json:"name" th:"2,optional"`
	Code        string `json:"code" th:"3,optional"`
	ParentId    int64  `json:"parent_id" th:"4,optional"`
	Level       int32  `json:"level" th:"5,optional"`
	Description string `json:"description" th:"6,optional"`
	SortOrder   int32  `json:"sort_order" th:"7,optional"`
	Status      int32  `json:"status" th:"8,optional"`
}

type GetDisputeListRequest struct {
	UserId         int64  `json:"user_id" th:"1,optional"`
	Role           int32  `json:"role" th:"2,optional"`
	OrganizationId int64  `json:"organization_id" th:"3,optional"`
	Page           int32  `json:"page" th:"4,optional"`
	PageSize       int32  `json:"page_size" th:"5,optional"`
	Status         int32  `json:"status" th:"6,optional"`
	TypeId         int64  `json:"type_id" th:"7,optional"`
	Keyword        string `json:"keyword" th:"8,optional"`
}

type GetDisputeListResponse struct {
	Code    int32           `json:"code" th:"1,optional"`
	Message string          `json:"message" th:"2,optional"`
	Cases   []*DisputeCase  `json:"cases" th:"3,optional"`
	Total   int64           `json:"total" th:"4,optional"`
}

type GetDisputeDetailRequest struct {
	CaseId int64 `json:"case_id" th:"1,optional"`
	UserId int64 `json:"user_id" th:"2,optional"`
	Role   int32 `json:"role" th:"3,optional"`
}

type GetDisputeDetailResponse struct {
	Code     int32        `json:"code" th:"1,optional"`
	Message  string       `json:"message" th:"2,optional"`
	CaseInfo *DisputeCase `json:"case_info" th:"3,optional"`
}

type CreateDisputeRequest struct {
	Dispute  *DisputeCase `json:"dispute" th:"1,optional"`
	Evidence []*Evidence  `json:"evidence" th:"2,optional"`
}

type CreateDisputeResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
	CaseId  int64  `json:"case_id" th:"3,optional"`
	CaseNo  string `json:"case_no" th:"4,optional"`
}

type KioskCreateDisputeRequest struct {
	Dispute  *DisputeCase `json:"dispute" th:"1,optional"`
	Evidence []*Evidence  `json:"evidence" th:"2,optional"`
	DeviceId int64        `json:"device_id" th:"3,optional"`
}

type AssignDisputeRequest struct {
	CaseId     int64 `json:"case_id" th:"1,optional"`
	MediatorId int64 `json:"mediator_id" th:"2,optional"`
	AssignorId int64 `json:"assignor_id" th:"3,optional"`
}

type AssignDisputeResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type UrgeDisputeRequest struct {
	CaseId   int64  `json:"case_id" th:"1,optional"`
	UserId   int64  `json:"user_id" th:"2,optional"`
	UrgeType int32  `json:"urge_type" th:"3,optional"`
	Remark   string `json:"remark" th:"4,optional"`
}

type UrgeDisputeResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type UpdateDisputeStatusRequest struct {
	CaseId int64  `json:"case_id" th:"1,optional"`
	Status int32  `json:"status" th:"2,optional"`
	UserId int64  `json:"user_id" th:"3,optional"`
	Remark string `json:"remark" th:"4,optional"`
}

type UpdateDisputeStatusResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type GetDisputeHistoryRequest struct {
	CaseId int64 `json:"case_id" th:"1,optional"`
}

type GetDisputeHistoryResponse struct {
	Code    int32             `json:"code" th:"1,optional"`
	Message string            `json:"message" th:"2,optional"`
	History []*DisputeHistory `json:"history" th:"3,optional"`
}

type GetDisputeProgressRequest struct {
	CaseNo string `json:"case_no" th:"1,optional"`
	Idcard string `json:"idcard" th:"2,optional"`
}

type GetDisputeProgressResponse struct {
	Code     int32             `json:"code" th:"1,optional"`
	Message  string            `json:"message" th:"2,optional"`
	CaseInfo *DisputeCase      `json:"case_info" th:"3,optional"`
	History  []*DisputeHistory `json:"history" th:"4,optional"`
}

type GetDisputeTypesResponse struct {
	Code    int32           `json:"code" th:"1,optional"`
	Message string          `json:"message" th:"2,optional"`
	Types   []*DisputeType  `json:"types" th:"3,optional"`
}

type UploadEvidenceRequest struct {
	CaseId     int64  `json:"case_id" th:"1,optional"`
	FileType   int32  `json:"file_type" th:"2,optional"`
	FileName   string `json:"file_name" th:"3,optional"`
	FileUrl    string `json:"file_url" th:"4,optional"`
	FileSize   int64  `json:"file_size" th:"5,optional"`
	Remark     string `json:"remark" th:"6,optional"`
	UploadFrom int32  `json:"upload_from" th:"7,optional"`
	UserId     int64  `json:"user_id" th:"8,optional"`
}

type UploadEvidenceResponse struct {
	Code     int32     `json:"code" th:"1,optional"`
	Message  string    `json:"message" th:"2,optional"`
	Evidence *Evidence `json:"evidence" th:"3,optional"`
}

type GetEvidenceListRequest struct {
	CaseId   int64 `json:"case_id" th:"1,optional"`
	Page     int32 `json:"page" th:"2,optional"`
	PageSize int32 `json:"page_size" th:"3,optional"`
}

type GetEvidenceListResponse struct {
	Code     int32       `json:"code" th:"1,optional"`
	Message  string      `json:"message" th:"2,optional"`
	Evidence []*Evidence `json:"evidence" th:"3,optional"`
	Total    int64       `json:"total" th:"4,optional"`
}

type DeleteEvidenceRequest struct {
	EvidenceId int64 `json:"evidence_id" th:"1,optional"`
	UserId     int64 `json:"user_id" th:"2,optional"`
}

type DeleteEvidenceResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type BatchDeleteEvidenceRequest struct {
	EvidenceIds []int64 `json:"evidence_ids" th:"1,optional"`
	UserId      int64   `json:"user_id" th:"2,optional"`
}

type BatchDeleteEvidenceResponse struct {
	Code         int32  `json:"code" th:"1,optional"`
	Message      string `json:"message" th:"2,optional"`
	DeletedCount int64  `json:"deleted_count" th:"3,optional"`
}

type UpdateEvidenceRemarkRequest struct {
	EvidenceId int64  `json:"evidence_id" th:"1,optional"`
	Remark     string `json:"remark" th:"2,optional"`
	UserId     int64  `json:"user_id" th:"3,optional"`
}

type UpdateEvidenceRemarkResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}
