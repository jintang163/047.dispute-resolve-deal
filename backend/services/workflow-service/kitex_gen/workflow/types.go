package workflow

type ApprovalRecord struct {
	Id              int64  `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	CaseId          int64  `thrift:"case_id,2" frugal:"2,default,i64" json:"caseId"`
	WorkflowId      int64  `thrift:"workflow_id,3" frugal:"3,default,i64" json:"workflowId"`
	WorkflowName    string `thrift:"workflow_name,4" frugal:"4,default,string" json:"workflowName"`
	NodeType        int32  `thrift:"node_type,5" frugal:"5,default,i32" json:"nodeType"`
	NodeName        string `thrift:"node_name,6" frugal:"6,default,string" json:"nodeName"`
	ApproverId      int64  `thrift:"approver_id,7" frugal:"7,default,i64" json:"approverId"`
	ApproverName    string `thrift:"approver_name,8" frugal:"8,default,string" json:"approverName"`
	Status          int32  `thrift:"status,9" frugal:"9,default,i32" json:"status"`
	Remark          string `thrift:"remark,10" frugal:"10,default,string" json:"remark"`
	ApproveAction   int32  `thrift:"approve_action,11" frugal:"11,default,i32" json:"approveAction"`
	ActionName      string `thrift:"action_name,12" frugal:"12,default,string" json:"actionName"`
	ApprovedAt      string `thrift:"approved_at,13" frugal:"13,default,string" json:"approvedAt"`
	SortOrder       int32  `thrift:"sort_order,14" frugal:"14,default,i32" json:"sortOrder"`
	SignUserId      int64  `thrift:"sign_user_id,15" frugal:"15,default,i64" json:"signUserId"`
	SignUserName    string `thrift:"sign_user_name,16" frugal:"16,default,string" json:"signUserName"`
	TransferUserId  int64  `thrift:"transfer_user_id,17" frugal:"17,default,i64" json:"transferUserId"`
	TransferUserName string `thrift:"transfer_user_name,18" frugal:"18,default,string" json:"transferUserName"`
	Level           int32  `thrift:"level,19" frugal:"19,default,i32" json:"level"`
	Deadline        string `thrift:"deadline,20" frugal:"20,default,string" json:"deadline"`
	TimeoutLevel    int32  `thrift:"timeout_level,21" frugal:"21,default,i32" json:"timeoutLevel"`
	CreatedAt       string `thrift:"created_at,22" frugal:"22,default,string" json:"createdAt"`
	UpdatedAt       string `thrift:"updated_at,23" frugal:"23,default,string" json:"updatedAt"`
}

type WorkflowDefinition struct {
	Id             int64  `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	Name           string `thrift:"name,2" frugal:"2,default,string" json:"name"`
	Code           string `thrift:"code,3" frugal:"3,default,string" json:"code"`
	DisputeTypeId  int64  `thrift:"dispute_type_id,4" frugal:"4,default,i64" json:"disputeTypeId"`
	Description    string `thrift:"description,5" frugal:"5,default,string" json:"description"`
	FlowConfig     string `thrift:"flow_config,6" frugal:"6,default,string" json:"flowConfig"`
	Version        int32  `thrift:"version,7" frugal:"7,default,i32" json:"version"`
	Status         int32  `thrift:"status,8" frugal:"8,default,i32" json:"status"`
	CreatedBy      int64  `thrift:"created_by,9" frugal:"9,default,i64" json:"createdBy"`
	CreatedAt      string `thrift:"created_at,10" frugal:"10,default,string" json:"createdAt"`
}

type SubmitApprovalRequest struct {
	CaseId     int64 `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
	UserId     int64 `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
	WorkflowId int64 `thrift:"workflow_id,3" frugal:"3,default,i64" json:"workflowId"`
}

type SubmitApprovalResponse struct {
	Code    int32          `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string         `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Record  *ApprovalRecord `thrift:"record,3" frugal:"3,default,*ApprovalRecord" json:"record,omitempty"`
}

type ProcessApprovalRequest struct {
	ApprovalId   int64  `thrift:"approval_id,1" frugal:"1,default,i64" json:"approvalId"`
	UserId       int64  `thrift:"user_id,2" frugal:"2,default,i64" json:"userId"`
	Action       int32  `thrift:"action,3" frugal:"3,default,i32" json:"action"`
	Remark       string `thrift:"remark,4" frugal:"4,default,string" json:"remark"`
	TargetUserId int64  `thrift:"target_user_id,5" frugal:"5,default,i64" json:"targetUserId"`
}

type ProcessApprovalResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type GetApprovalProgressRequest struct {
	CaseId int64 `thrift:"case_id,1" frugal:"1,default,i64" json:"caseId"`
}

type GetApprovalProgressResponse struct {
	Code    int32            `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string           `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Records []*ApprovalRecord `thrift:"records,3" frugal:"3,default,list<*ApprovalRecord>" json:"records,omitempty"`
}

type GetApprovalListRequest struct {
	UserId   int64 `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
	Page     int32 `thrift:"page,2" frugal:"2,default,i32" json:"page"`
	PageSize int32 `thrift:"page_size,3" frugal:"3,default,i32" json:"pageSize"`
	Status   int32 `thrift:"status,4" frugal:"4,default,i32" json:"status"`
}

type GetApprovalListResponse struct {
	Code    int32            `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string           `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Records []*ApprovalRecord `thrift:"records,3" frugal:"3,default,list<*ApprovalRecord>" json:"records,omitempty"`
	Total   int64            `thrift:"total,4" frugal:"4,default,i64" json:"total"`
}

type ProcessTimeoutUpgradeResponse struct {
	Code           int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message        string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	ProcessedCount int32  `thrift:"processed_count,3" frugal:"3,default,i32" json:"processedCount"`
}
