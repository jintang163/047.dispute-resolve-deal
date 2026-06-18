namespace go workflow

struct ApprovalRecord {
    1: i64 id
    2: i64 case_id
    3: i64 workflow_id
    4: string workflow_name
    5: i32 node_type
    6: string node_name
    7: i64 approver_id
    8: string approver_name
    9: i32 status
    10: string remark
    11: i32 approve_action
    12: string action_name
    13: string approved_at
    14: i32 sort_order
    15: i64 sign_user_id
    16: string sign_user_name
    17: i64 transfer_user_id
    18: string transfer_user_name
    19: i32 level
    20: string deadline
    21: i32 timeout_level
    22: string created_at
    23: string updated_at
}

struct WorkflowDefinition {
    1: i64 id
    2: string name
    3: string code
    4: i64 dispute_type_id
    5: string description
    6: string flow_config
    7: i32 version
    8: i32 status
    9: i64 created_by
    10: string created_at
}

struct SubmitApprovalRequest {
    1: i64 case_id
    2: i64 user_id
    3: i64 workflow_id
}

struct SubmitApprovalResponse {
    1: i32 code
    2: string message
    3: ApprovalRecord record
}

struct ProcessApprovalRequest {
    1: i64 approval_id
    2: i64 user_id
    3: i32 action
    4: string remark
    5: i64 target_user_id
}

struct ProcessApprovalResponse {
    1: i32 code
    2: string message
}

struct GetApprovalProgressRequest {
    1: i64 case_id
}

struct GetApprovalProgressResponse {
    1: i32 code
    2: string message
    3: list<ApprovalRecord> records
}

struct GetApprovalListRequest {
    1: i64 user_id
    2: i32 page
    3: i32 page_size
    4: i32 status
}

struct GetApprovalListResponse {
    1: i32 code
    2: string message
    3: list<ApprovalRecord> records
    4: i64 total
}

struct ProcessTimeoutUpgradeResponse {
    1: i32 code
    2: string message
    3: i32 processed_count
}

service WorkflowService {
    SubmitApprovalResponse SubmitApproval(1: SubmitApprovalRequest request)
    ProcessApprovalResponse ProcessApproval(1: ProcessApprovalRequest request)
    GetApprovalProgressResponse GetApprovalProgress(1: GetApprovalProgressRequest request)
    GetApprovalListResponse GetApprovalTodoList(1: GetApprovalListRequest request)
    GetApprovalListResponse GetApprovalDoneList(1: GetApprovalListRequest request)
    ProcessTimeoutUpgradeResponse ProcessTimeoutUpgrade()
}
