namespace go dispute

struct DisputeCase {
    1: i64 id
    2: string case_no
    3: string title
    4: i64 type_id
    5: string type_name
    6: i32 case_level
    7: string description
    8: string expected_solution
    9: i32 source
    10: i64 applicant_id
    11: string applicant_name
    12: string applicant_idcard
    13: string applicant_mobile
    14: string respondent_name
    15: string respondent_idcard
    16: string respondent_mobile
    17: double longitude
    18: double latitude
    19: string address
    20: i64 organization_id
    21: i64 mediator_id
    22: string mediator_name
    23: i32 status
    24: i32 urgency_level
    25: i64 created_by
    26: string created_at
    27: string updated_at
}

struct Evidence {
    1: i64 id
    2: i64 case_id
    3: i32 file_type
    4: string file_name
    5: string file_url
    6: i64 file_size
    7: string remark
    8: i32 sort_order
    9: i32 upload_from
    10: i64 uploaded_by
    11: string created_at
}

struct DisputeHistory {
    1: i64 id
    2: i64 case_id
    3: i32 action_type
    4: string action_name
    5: string remark
    6: i64 operator_id
    7: string operator_name
    8: string created_at
}

struct DisputeType {
    1: i64 id
    2: string name
    3: string code
    4: i64 parent_id
    5: i32 level
    6: string description
    7: i32 sort_order
    8: i32 status
}

struct GetDisputeListRequest {
    1: i64 user_id
    2: i32 role
    3: i64 organization_id
    4: i32 page
    5: i32 page_size
    6: i32 status
    7: i64 type_id
    8: string keyword
}

struct GetDisputeListResponse {
    1: i32 code
    2: string message
    3: list<DisputeCase> cases
    4: i64 total
}

struct GetDisputeDetailRequest {
    1: i64 case_id
    2: i64 user_id
    3: i32 role
}

struct GetDisputeDetailResponse {
    1: i32 code
    2: string message
    3: DisputeCase case_info
}

struct CreateDisputeRequest {
    1: DisputeCase dispute
    2: list<Evidence> evidence
}

struct CreateDisputeResponse {
    1: i32 code
    2: string message
    3: i64 case_id
    4: string case_no
}

struct KioskCreateDisputeRequest {
    1: DisputeCase dispute
    2: list<Evidence> evidence
    3: i64 device_id
}

struct AssignDisputeRequest {
    1: i64 case_id
    2: i64 mediator_id
    3: i64 assignor_id
}

struct AssignDisputeResponse {
    1: i32 code
    2: string message
}

struct UrgeDisputeRequest {
    1: i64 case_id
    2: i64 user_id
    3: i32 urge_type
    4: string remark
}

struct UrgeDisputeResponse {
    1: i32 code
    2: string message
}

struct UpdateDisputeStatusRequest {
    1: i64 case_id
    2: i32 status
    3: i64 user_id
    4: string remark
}

struct UpdateDisputeStatusResponse {
    1: i32 code
    2: string message
}

struct GetDisputeHistoryRequest {
    1: i64 case_id
}

struct GetDisputeHistoryResponse {
    1: i32 code
    2: string message
    3: list<DisputeHistory> history
}

struct GetDisputeProgressRequest {
    1: string case_no
    2: string idcard
}

struct GetDisputeProgressResponse {
    1: i32 code
    2: string message
    3: DisputeCase case_info
    4: list<DisputeHistory> history
}

struct GetDisputeTypesResponse {
    1: i32 code
    2: string message
    3: list<DisputeType> types
}

struct UploadEvidenceRequest {
    1: i64 case_id
    2: i32 file_type
    3: string file_name
    4: string file_url
    5: i64 file_size
    6: string remark
    7: i32 upload_from
    8: i64 user_id
}

struct UploadEvidenceResponse {
    1: i32 code
    2: string message
    3: Evidence evidence
}

struct GetEvidenceListRequest {
    1: i64 case_id
    2: i32 page
    3: i32 page_size
}

struct GetEvidenceListResponse {
    1: i32 code
    2: string message
    3: list<Evidence> evidence
    4: i64 total
}

struct DeleteEvidenceRequest {
    1: i64 evidence_id
    2: i64 user_id
}

struct DeleteEvidenceResponse {
    1: i32 code
    2: string message
}

struct BatchDeleteEvidenceRequest {
    1: list<i64> evidence_ids
    2: i64 user_id
}

struct BatchDeleteEvidenceResponse {
    1: i32 code
    2: string message
    3: i64 deleted_count
}

struct UpdateEvidenceRemarkRequest {
    1: i64 evidence_id
    2: string remark
    3: i64 user_id
}

struct UpdateEvidenceRemarkResponse {
    1: i32 code
    2: string message
}

service DisputeService {
    GetDisputeListResponse GetDisputeList(1: GetDisputeListRequest request)
    GetDisputeDetailResponse GetDisputeDetail(1: GetDisputeDetailRequest request)
    CreateDisputeResponse CreateDispute(1: CreateDisputeRequest request)
    CreateDisputeResponse KioskCreateDispute(1: KioskCreateDisputeRequest request)
    CreateDisputeResponse MiniAppCreateDispute(1: CreateDisputeRequest request)
    AssignDisputeResponse AssignDispute(1: AssignDisputeRequest request)
    UrgeDisputeResponse UrgeDispute(1: UrgeDisputeRequest request)
    UpdateDisputeStatusResponse UpdateDisputeStatus(1: UpdateDisputeStatusRequest request)
    GetDisputeHistoryResponse GetDisputeHistory(1: GetDisputeHistoryRequest request)
    GetDisputeProgressResponse GetDisputeProgress(1: GetDisputeProgressRequest request)
    GetDisputeTypesResponse GetDisputeTypes()
    UploadEvidenceResponse UploadEvidence(1: UploadEvidenceRequest request)
    GetEvidenceListResponse GetEvidenceList(1: GetEvidenceListRequest request)
    DeleteEvidenceResponse DeleteEvidence(1: DeleteEvidenceRequest request)
    BatchDeleteEvidenceResponse BatchDeleteEvidence(1: BatchDeleteEvidenceRequest request)
    UpdateEvidenceRemarkResponse UpdateEvidenceRemark(1: UpdateEvidenceRemarkRequest request)
}
