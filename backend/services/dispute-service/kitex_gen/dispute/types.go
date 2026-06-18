package dispute

type DisputeCase struct {
	Id                int64   `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	CaseNo            string  `thrift:"case_no,2" frugal:"2,default,string" json:"caseNo"`
	Title             string  `thrift:"title,3" frugal:"3,default,string" json:"title"`
	TypeId            int64   `thrift:"type_id,4" frugal:"4,default,i64" json:"typeId"`
	TypeName          string  `thrift:"type_name,5" frugal:"5,default,string" json:"typeName"`
	CaseLevel         int32   `thrift:"case_level,6" frugal:"6,default,i32" json:"caseLevel"`
	Description       string  `thrift:"description,7" frugal:"7,default,string" json:"description"`
	ExpectedSolution  string  `thrift:"expected_solution,8" frugal:"8,default,string" json:"expectedSolution"`
	Source            int32   `thrift:"source,9" frugal:"9,default,i32" json:"source"`
	ApplicantId       int64   `thrift:"applicant_id,10" frugal:"10,default,i64" json:"applicantId"`
	ApplicantName     string  `thrift:"applicant_name,11" frugal:"11,default,string" json:"applicantName"`
	ApplicantIdcard   string  `thrift:"applicant_idcard,12" frugal:"12,default,string" json:"applicantIdcard"`
	ApplicantMobile   string  `thrift:"applicant_mobile,13" frugal:"13,default,string" json:"applicantMobile"`
	RespondentName    string  `thrift:"respondent_name,14" frugal:"14,default,string" json:"respondentName"`
	RespondentIdcard  string  `thrift:"respondent_idcard,15" frugal:"15,default,string" json:"respondentIdcard"`
	RespondentMobile  string  `thrift:"respondent_mobile,16" frugal:"16,default,string" json:"respondentMobile"`
	Longitude         float64 `thrift:"longitude,17" frugal:"17,default,double" json:"longitude"`
	Latitude          float64 `thrift:"latitude,18" frugal:"18,default,double" json:"latitude"`
	Address           string  `thrift:"address,19" frugal:"19,default,string" json:"address"`
	OrganizationId    int64   `thrift:"organization_id,20" frugal:"20,default,i64" json:"organizationId"`
	MediatorId        int64   `thrift:"mediator_id,21" frugal:"21,default,i64" json:"mediatorId"`
	MediatorName      string  `thrift:"mediator_name,22" frugal:"22,default,string" json:"mediatorName"`
	Status            int32   `thrift:"status,23" frugal:"23,default,i32" json:"status"`
	UrgencyLevel      int32   `thrift:"urgency_level,24" frugal:"24,default,i32" json:"urgencyLevel"`
	CreatedBy         int64   `thrift:"created_by,25" frugal:"25,default,i64" json:"createdBy"`
	CreatedAt         string  `thrift:"created_at,26" frugal:"26,default,string" json:"createdAt"`
	UpdatedAt         string  `thrift:"updated_at,27" frugal:"27,default,string" json:"updatedAt"`
}

type Evidence struct {
	Id          int64  `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	CaseId      int64  `thrift:"case_id,2" frugal:"2,default,i64" json:"caseId"`
	FileType    int32  `thrift:"file_type,3" frugal:"3,default,i32" json:"fileType"`
	FileName    string `thrift:"file_name,4" frugal:"4,default,string" json:"fileName"`
	FileUrl     string `thrift:"file_url,5" frugal:"5,default,string" json:"fileUrl"`
	FileSize    int64  `thrift:"file_size,6" frugal:"6,default,i64" json:"fileSize"`
	Remark      string `thrift:"remark,7" frugal:"7,default,string" json:"remark"`
	SortOrder   int32  `thrift:"sort_order,8" frugal:"8,default,i32" json:"sortOrder"`
	UploadFrom  int32  `thrift:"upload_from,9" frugal:"9,default,i32" json:"uploadFrom"`
	UploadedBy  int64  `thrift:"uploaded_by,10" frugal:"10,default,i64" json:"uploadedBy"`
	CreatedAt   string `thrift:"created_at,11" frugal:"11,default,string" json:"createdAt"`
}

type DisputeHistory struct {
	Id           int64  `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	CaseId       int64  `thrift:"case_id,2" frugal:"2,default,i64" json:"caseId"`
	ActionType   int32  `thrift:"action_type,3" frugal:"3,default,i32" json:"actionType"`
	ActionName   string `thrift:"action_name,4" frugal:"4,default,string" json:"actionName"`
	Remark       string `thrift:"remark,5" frugal:"5,default,string" json:"remark"`
	OperatorId   int64  `thrift:"operator_id,6" frugal:"6,default,i64" json:"operatorId"`
	OperatorName string `thrift:"operator_name,7" frugal:"7,default,string" json:"operatorName"`
	CreatedAt    string `thrift:"created_at,8" frugal:"8,default,string" json:"createdAt"`
}

type DisputeType struct {
	Id          int64  `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	Name        string `thrift:"name,2" frugal:"2,default,string" json:"name"`
	Code        string `thrift:"code,3" frugal:"3,default,string" json:"code"`
	ParentId    int64  `thrift:"parent_id,4" frugal:"4,default,i64" json:"parentId"`
	Level       int32  `thrift:"level,5" frugal:"5,default,i32" json:"level"`
	Description string `thrift:"description,6" frugal:"6,default,string" json:"description"`
	SortOrder   int32  `thrift:"sort_order,7" frugal:"7,default,i32" json:"sortOrder"`
	Status      int32  `thrift:"status,8" frugal:"8,default,i32" json:"status"`
}

type GetDisputeListRequest struct {
	UserId         int64  `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
	Role           int32  `thrift:"role,2" frugal:"2,default,i32" json:"role"`
	OrganizationId int64  `thrift:"organization_id,3" frugal:"3,default,i64" json:"organizationId"`
	Page           int32  `thrift:"page,4" frugal:"4,default,i32" json:"page"`
	PageSize       int32  `thrift:"page_size,5" frugal:"5,default,i32" json:"pageSize"`
	Status         int32  `thrift:"status,6" frugal:"6,default,i32" json:"status"`
	TypeId         int64  `thrift:"type_id,7" frugal:"7,default,i64" json:"typeId"`
	Keyword        string `thrift:"keyword,8" frugal:"8,default,string" json:"keyword"`
}

type GetDisputeListResponse struct {
	Code    int32           `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string          `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Cases   []*DisputeCase  `thrift:"cases,3" frugal:"3,default,list<*DisputeCase>" json:"cases,omitempty"`
	Total   int64           `thrift:"total,4" frugal:"4,default,i64" json:"total"`
}

type GetDisputeDetailRequest struct {
	CaseId int64 `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
	UserId int64 `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
	Role   int32 `thrift:"role,3" frugal:"3,default,i32" json:"role"`
}

type GetDisputeDetailResponse struct {
	Code     int32       `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message  string      `thrift:"message,2" frugal:"2,default,string" json:"message"`
	CaseInfo *DisputeCase `thrift:"case_info,3" frugal:"3,default,*DisputeCase" json:"caseInfo,omitempty"`
}

type CreateDisputeRequest struct {
	Dispute  *DisputeCase `thrift:"dispute,1" frugal:"1,default,*DisputeCase" json:"dispute,omitempty"`
	Evidence []*Evidence  `thrift:"evidence,2" frugal:"2,default,list<*Evidence>" json:"evidence,omitempty"`
}

type CreateDisputeResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	CaseId  int64  `thrift:"case_id,3" frugal:"3,default,i64" json:"caseId"`
	CaseNo  string `thrift:"case_no,4" frugal:"4,default,string" json:"caseNo"`
}

type KioskCreateDisputeRequest struct {
	Dispute  *DisputeCase `thrift:"dispute,1" frugal:"1,default,*DisputeCase" json:"dispute,omitempty"`
	Evidence []*Evidence  `thrift:"evidence,2" frugal:"2,default,list<*Evidence>" json:"evidence,omitempty"`
	DeviceId int64        `thrift:"device_id,3" frugal:"3,default,i64" json:"deviceId"`
}

type AssignDisputeRequest struct {
	CaseId     int64 `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
	MediatorId int64 `thrift:"mediator_id,2" frugal:"2,default,i64" json:"mediatorId"`
	AssignorId int64 `thrift:"assignor_id,3" frugal:"3,default,i64" json:"assignorId"`
}

type AssignDisputeResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type UrgeDisputeRequest struct {
	CaseId   int64  `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
	UserId   int64  `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
	UrgeType int32  `thrift:"urge_type,3" frugal:"3,default,i32" json:"urgeType"`
	Remark   string `thrift:"remark,4" frugal:"4,default,string" json:"remark"`
}

type UrgeDisputeResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type UpdateDisputeStatusRequest struct {
	CaseId int64  `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
	Status int32  `thrift:"status,2" frugal:"2,default,i32" json:"status"`
	UserId int64  `thrift:"user_id,3" frugal:"3,default,i64" json:"userId"`
	Remark string `thrift:"remark,4" frugal:"4,default,string" json:"remark"`
}

type UpdateDisputeStatusResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type GetDisputeHistoryRequest struct {
	CaseId int64 `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
}

type GetDisputeHistoryResponse struct {
	Code    int32             `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string            `thrift:"message,2" frugal:"2,default,string" json:"message"`
	History []*DisputeHistory `thrift:"history,3" frugal:"3,default,list<*DisputeHistory>" json:"history,omitempty"`
}

type GetDisputeProgressRequest struct {
	CaseNo string `thrift:"case_no,1" frugal:"1,default,string" json:"caseNo"`
	Idcard string `thrift:"idcard,2" frugal:"2,default,string" json:"idcard"`
}

type GetDisputeProgressResponse struct {
	Code     int32             `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message  string            `thrift:"message,2" frugal:"2,default,string" json:"message"`
	CaseInfo *DisputeCase      `thrift:"case_info,3" frugal:"3,default,*DisputeCase" json:"caseInfo,omitempty"`
	History  []*DisputeHistory `thrift:"history,4" frugal:"4,default,list<*DisputeHistory>" json:"history,omitempty"`
}

type GetDisputeTypesResponse struct {
	Code    int32           `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string          `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Types   []*DisputeType  `thrift:"types,3" frugal:"3,default,list<*DisputeType>" json:"types,omitempty"`
}

type UploadEvidenceRequest struct {
	CaseId     int64  `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
	FileType   int32  `thrift:"file_type,2" frugal:"2,default,i32" json:"fileType"`
	FileName   string `thrift:"file_name,3" frugal:"3,default,string" json:"fileName"`
	FileUrl    string `thrift:"file_url,4" frugal:"4,default,string" json:"fileUrl"`
	FileSize   int64  `thrift:"file_size,5" frugal:"5,default,i64" json:"fileSize"`
	Remark     string `thrift:"remark,6" frugal:"6,default,string" json:"remark"`
	UploadFrom int32  `thrift:"upload_from,7" frugal:"7,default,i32" json:"uploadFrom"`
	UserId     int64  `thrift:"user_id,8" frugal:"8,default,i64" json:"userId"`
}

type UploadEvidenceResponse struct {
	Code     int32     `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message  string    `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Evidence *Evidence `thrift:"evidence,3" frugal:"3,default,*Evidence" json:"evidence,omitempty"`
}

type GetEvidenceListRequest struct {
	CaseId   int64 `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
	Page     int32 `thrift:"page,2" frugal:"2,default,i32" json:"page"`
	PageSize int32 `thrift:"page_size,3" frugal:"3,default,i32" json:"pageSize"`
}

type GetEvidenceListResponse struct {
	Code     int32       `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message  string      `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Evidence []*Evidence `thrift:"evidence,3" frugal:"3,default,list<*Evidence>" json:"evidence,omitempty"`
	Total    int64       `thrift:"total,4" frugal:"4,default,i64" json:"total"`
}

type DeleteEvidenceRequest struct {
	EvidenceId int64 `thrift:"evidence_id,1" frugal:"1,default,i64" json:"evidenceId"`
	UserId     int64 `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
}

type DeleteEvidenceResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type BatchDeleteEvidenceRequest struct {
	EvidenceIds []int64 `thrift:"evidence_ids,1" frugal:"1,default,list<i64>" json:"evidenceIds"`
	UserId      int64   `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
}

type BatchDeleteEvidenceResponse struct {
	Code         int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message      string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	DeletedCount int64  `thrift:"deleted_count,3" frugal:"3,default,i64" json:"deletedCount"`
}

type UpdateEvidenceRemarkRequest struct {
	EvidenceId int64  `thrift:"evidence_id,1" frugal:"1,default,i64" json:"evidenceId"`
	Remark     string `thrift:"remark,2" frugal:"2,default,string" json:"remark"`
	UserId     int64  `thrift:"user_id,3" frugal:"3,default,i64" json:"userId"`
}

type UpdateEvidenceRemarkResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}
