package workflow

type ApprovalRecord struct {
	Id               int64  `json:"id" th:"1,optional"`
	CaseId           int64  `json:"case_id" th:"2,optional"`
	WorkflowId       int64  `json:"workflow_id" th:"3,optional"`
	WorkflowName     string `json:"workflow_name" th:"4,optional"`
	NodeType         int32  `json:"node_type" th:"5,optional"`
	NodeName         string `json:"node_name" th:"6,optional"`
	ApproverId       int64  `json:"approver_id" th:"7,optional"`
	ApproverName     string `json:"approver_name" th:"8,optional"`
	Status           int32  `json:"status" th:"9,optional"`
	Remark           string `json:"remark" th:"10,optional"`
	ApproveAction    int32  `json:"approve_action" th:"11,optional"`
	ActionName       string `json:"action_name" th:"12,optional"`
	ApprovedAt       string `json:"approved_at" th:"13,optional"`
	SortOrder        int32  `json:"sort_order" th:"14,optional"`
	SignUserId       int64  `json:"sign_user_id" th:"15,optional"`
	SignUserName     string `json:"sign_user_name" th:"16,optional"`
	TransferUserId   int64  `json:"transfer_user_id" th:"17,optional"`
	TransferUserName string `json:"transfer_user_name" th:"18,optional"`
	Level            int32  `json:"level" th:"19,optional"`
	Deadline         string `json:"deadline" th:"20,optional"`
	TimeoutLevel     int32  `json:"timeout_level" th:"21,optional"`
	CreatedAt        string `json:"created_at" th:"22,optional"`
	UpdatedAt        string `json:"updated_at" th:"23,optional"`
}

type WorkflowDefinition struct {
	Id            int64  `json:"id" th:"1,optional"`
	Name          string `json:"name" th:"2,optional"`
	Code          string `json:"code" th:"3,optional"`
	DisputeTypeId int64  `json:"dispute_type_id" th:"4,optional"`
	Description   string `json:"description" th:"5,optional"`
	FlowConfig    string `json:"flow_config" th:"6,optional"`
	Version       int32  `json:"version" th:"7,optional"`
	Status        int32  `json:"status" th:"8,optional"`
	CreatedBy     int64  `json:"created_by" th:"9,optional"`
	CreatedAt     string `json:"created_at" th:"10,optional"`
}

type SubmitApprovalRequest struct {
	CaseId     int64 `json:"case_id" th:"1,optional"`
	UserId     int64 `json:"user_id" th:"2,optional"`
	WorkflowId int64 `json:"workflow_id" th:"3,optional"`
}

type SubmitApprovalResponse struct {
	Code    int32           `json:"code" th:"1,optional"`
	Message string          `json:"message" th:"2,optional"`
	Record  *ApprovalRecord `json:"record" th:"3,optional"`
}

type ProcessApprovalRequest struct {
	ApprovalId   int64  `json:"approval_id" th:"1,optional"`
	UserId       int64  `json:"user_id" th:"2,optional"`
	Action       int32  `json:"action" th:"3,optional"`
	Remark       string `json:"remark" th:"4,optional"`
	TargetUserId int64  `json:"target_user_id" th:"5,optional"`
}

type ProcessApprovalResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type GetApprovalProgressRequest struct {
	CaseId int64 `json:"case_id" th:"1,optional"`
}

type GetApprovalProgressResponse struct {
	Code    int32             `json:"code" th:"1,optional"`
	Message string            `json:"message" th:"2,optional"`
	Records []*ApprovalRecord `json:"records" th:"3,optional"`
}

type GetApprovalListRequest struct {
	UserId   int64 `json:"user_id" th:"1,optional"`
	Page     int32 `json:"page" th:"2,optional"`
	PageSize int32 `json:"page_size" th:"3,optional"`
	Status   int32 `json:"status" th:"4,optional"`
}

type GetApprovalListResponse struct {
	Code    int32             `json:"code" th:"1,optional"`
	Message string            `json:"message" th:"2,optional"`
	Records []*ApprovalRecord `json:"records" th:"3,optional"`
	Total   int64             `json:"total" th:"4,optional"`
}

type ProcessTimeoutUpgradeResponse struct {
	Code           int32  `json:"code" th:"1,optional"`
	Message        string `json:"message" th:"2,optional"`
	ProcessedCount int32  `json:"processed_count" th:"3,optional"`
}
