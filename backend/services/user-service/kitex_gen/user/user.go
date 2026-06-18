package user

type User struct {
	Id             int64  `json:"id" th:"1,optional"`
	Username       string `json:"username" th:"2,optional"`
	RealName       string `json:"real_name" th:"3,optional"`
	Role           int32  `json:"role" th:"4,optional"`
	Avatar         string `json:"avatar" th:"5,optional"`
	Mobile         string `json:"mobile" th:"6,optional"`
	Email          string `json:"email" th:"7,optional"`
	OrganizationId int64  `json:"organization_id" th:"8,optional"`
	Status         int32  `json:"status" th:"9,optional"`
	Openid         string `json:"openid" th:"10,optional"`
	CreatedAt      string `json:"created_at" th:"11,optional"`
	UpdatedAt      string `json:"updated_at" th:"12,optional"`
}

type Organization struct {
	Id           int64   `json:"id" th:"1,optional"`
	Name         string  `json:"name" th:"2,optional"`
	Code         string  `json:"code" th:"3,optional"`
	ParentId     int64   `json:"parent_id" th:"4,optional"`
	Level        int32   `json:"level" th:"5,optional"`
	SortOrder    int32   `json:"sort_order" th:"6,optional"`
	Leader       string  `json:"leader" th:"7,optional"`
	Contact      string  `json:"contact" th:"8,optional"`
	Address      string  `json:"address" th:"9,optional"`
	Longitude    float64 `json:"longitude" th:"10,optional"`
	Latitude     float64 `json:"latitude" th:"11,optional"`
	Status       int32   `json:"status" th:"12,optional"`
}

type LoginRequest struct {
	Username string `json:"username" th:"1,optional"`
	Password string `json:"password" th:"2,optional"`
}

type LoginResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
	User    *User  `json:"user" th:"3,optional"`
	Token   string `json:"token" th:"4,optional"`
}

type GetUserRequest struct {
	UserId int64 `json:"user_id" th:"1,optional"`
}

type GetUserResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
	User    *User  `json:"user" th:"3,optional"`
}

type GetUserListRequest struct {
	Page           int32  `json:"page" th:"1,optional"`
	PageSize       int32  `json:"page_size" th:"2,optional"`
	Keyword        string `json:"keyword" th:"3,optional"`
	Role           int32  `json:"role" th:"4,optional"`
	OrganizationId int64  `json:"organization_id" th:"5,optional"`
}

type GetUserListResponse struct {
	Code    int32   `json:"code" th:"1,optional"`
	Message string  `json:"message" th:"2,optional"`
	Users   []*User `json:"users" th:"3,optional"`
	Total   int64   `json:"total" th:"4,optional"`
}

type CreateUserRequest struct {
	User *User `json:"user" th:"1,optional"`
}

type CreateUserResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
	UserId  int64  `json:"user_id" th:"3,optional"`
}

type UpdateUserRequest struct {
	User *User `json:"user" th:"1,optional"`
}

type UpdateUserResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type DeleteUserRequest struct {
	UserId int64 `json:"user_id" th:"1,optional"`
}

type DeleteUserResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type ChangePasswordRequest struct {
	UserId      int64  `json:"user_id" th:"1,optional"`
	OldPassword string `json:"old_password" th:"2,optional"`
	NewPassword string `json:"new_password" th:"3,optional"`
}

type ChangePasswordResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type ResetPasswordRequest struct {
	UserId      int64  `json:"user_id" th:"1,optional"`
	NewPassword string `json:"new_password" th:"2,optional"`
}

type ResetPasswordResponse struct {
	Code    int32  `json:"code" th:"1,optional"`
	Message string `json:"message" th:"2,optional"`
}

type GetOrganizationTreeRequest struct {
	OrganizationId int64 `json:"organization_id" th:"1,optional"`
}

type GetOrganizationTreeResponse struct {
	Code          int32           `json:"code" th:"1,optional"`
	Message       string          `json:"message" th:"2,optional"`
	Organizations []*Organization `json:"organizations" th:"3,optional"`
}

type GetMediatorListRequest struct {
	OrganizationId int64 `json:"organization_id" th:"1,optional"`
	SpecialtyId    int64 `json:"specialty_id" th:"2,optional"`
}

type GetMediatorListResponse struct {
	Code      int32   `json:"code" th:"1,optional"`
	Message   string  `json:"message" th:"2,optional"`
	Mediators []*User `json:"mediators" th:"3,optional"`
}
