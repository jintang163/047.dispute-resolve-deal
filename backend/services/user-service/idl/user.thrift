namespace go user

struct User {
    1: i64 id
    2: string username
    3: string real_name
    4: i32 role
    5: string avatar
    6: string mobile
    7: string email
    8: i64 organization_id
    9: i32 status
    10: string openid
    11: string created_at
    12: string updated_at
}

struct Organization {
    1: i64 id
    2: string name
    3: string code
    4: i64 parent_id
    5: i32 level
    6: i32 sort_order
    7: string leader
    8: string contact
    9: string address
    10: double longitude
    11: double latitude
    12: i32 status
}

struct LoginRequest {
    1: string username
    2: string password
}

struct LoginResponse {
    1: i32 code
    2: string message
    3: User user
    4: string token
}

struct GetUserRequest {
    1: i64 user_id
}

struct GetUserResponse {
    1: i32 code
    2: string message
    3: User user
}

struct GetUserListRequest {
    1: i32 page
    2: i32 page_size
    3: string keyword
    4: i32 role
    5: i64 organization_id
}

struct GetUserListResponse {
    1: i32 code
    2: string message
    3: list<User> users
    4: i64 total
}

struct CreateUserRequest {
    1: User user
}

struct CreateUserResponse {
    1: i32 code
    2: string message
    3: i64 user_id
}

struct UpdateUserRequest {
    1: User user
}

struct UpdateUserResponse {
    1: i32 code
    2: string message
}

struct DeleteUserRequest {
    1: i64 user_id
}

struct DeleteUserResponse {
    1: i32 code
    2: string message
}

struct ChangePasswordRequest {
    1: i64 user_id
    2: string old_password
    3: string new_password
}

struct ChangePasswordResponse {
    1: i32 code
    2: string message
}

struct ResetPasswordRequest {
    1: i64 user_id
    2: string new_password
}

struct ResetPasswordResponse {
    1: i32 code
    2: string message
}

struct GetOrganizationTreeRequest {
    1: i64 organization_id
}

struct GetOrganizationTreeResponse {
    1: i32 code
    2: string message
    3: list<Organization> organizations
}

struct GetMediatorListRequest {
    1: i64 organization_id
    2: i64 specialty_id
}

struct GetMediatorListResponse {
    1: i32 code
    2: string message
    3: list<User> mediators
}

service UserService {
    LoginResponse Login(1: LoginRequest request)
    LoginResponse KioskLogin(1: string device_no)
    LoginResponse MiniAppLogin(1: string openid)
    GetUserResponse GetUserInfo(1: GetUserRequest request)
    ChangePasswordResponse ChangePassword(1: ChangePasswordRequest request)
    GetUserListResponse GetUserList(1: GetUserListRequest request)
    GetUserResponse GetUserDetail(1: GetUserRequest request)
    CreateUserResponse CreateUser(1: CreateUserRequest request)
    UpdateUserResponse UpdateUser(1: UpdateUserRequest request)
    DeleteUserResponse DeleteUser(1: DeleteUserRequest request)
    ResetPasswordResponse ResetPassword(1: ResetPasswordRequest request)
    GetOrganizationTreeResponse GetOrganizationTree(1: GetOrganizationTreeRequest request)
    GetMediatorListResponse GetMediatorList(1: GetMediatorListRequest request)
}
