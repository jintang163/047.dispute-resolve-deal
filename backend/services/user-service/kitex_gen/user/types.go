package user

type User struct {
	Id             int64  `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	Username       string `thrift:"username,2" frugal:"2,default,string" json:"username"`
	RealName       string `thrift:"real_name,3" frugal:"3,default,string" json:"realName"`
	Role           int32  `thrift:"role,4" frugal:"4,default,i32" json:"role"`
	Avatar         string `thrift:"avatar,5" frugal:"5,default,string" json:"avatar"`
	Mobile         string `thrift:"mobile,6" frugal:"6,default,string" json:"mobile"`
	Email          string `thrift:"email,7" frugal:"7,default,string" json:"email"`
	OrganizationId int64  `thrift:"organization_id,8" frugal:"8,default,i64" json:"organizationId"`
	Status         int32  `thrift:"status,9" frugal:"9,default,i32" json:"status"`
	Openid         string `thrift:"openid,10" frugal:"10,default,string" json:"openid"`
	CreatedAt      string `thrift:"created_at,11" frugal:"11,default,string" json:"createdAt"`
	UpdatedAt      string `thrift:"updated_at,12" frugal:"12,default,string" json:"updatedAt"`
}

type Organization struct {
	Id           int64   `thrift:"id,1" frugal:"1,default,i64" json:"id"`
	Name         string  `thrift:"name,2" frugal:"2,default,string" json:"name"`
	Code         string  `thrift:"code,3" frugal:"3,default,string" json:"code"`
	ParentId     int64   `thrift:"parent_id,4" frugal:"4,default,i64" json:"parentId"`
	Level        int32   `thrift:"level,5" frugal:"5,default,i32" json:"level"`
	SortOrder    int32   `thrift:"sort_order,6" frugal:"6,default,i32" json:"sortOrder"`
	Leader       string  `thrift:"leader,7" frugal:"7,default,string" json:"leader"`
	Contact      string  `thrift:"contact,8" frugal:"8,default,string" json:"contact"`
	Address      string  `thrift:"address,9" frugal:"9,default,string" json:"address"`
	Longitude    float64 `thrift:"longitude,10" frugal:"10,default,double" json:"longitude"`
	Latitude     float64 `thrift:"latitude,11" frugal:"11,default,double" json:"latitude"`
	Status       int32   `thrift:"status,12" frugal:"12,default,i32" json:"status"`
}

type LoginRequest struct {
	Username string `thrift:"username,1" frugal:"1,default,string" json:"username"`
	Password string `thrift:"password,2" frugal:"2,default,string" json:"password"`
}

type LoginResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	User    *User  `thrift:"user,3" frugal:"3,default,*User" json:"user,omitempty"`
	Token   string `thrift:"token,4" frugal:"4,default,string" json:"token"`
}

type GetUserRequest struct {
	UserId int64 `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
}

type GetUserResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	User    *User  `thrift:"user,3" frugal:"3,default,*User" json:"user,omitempty"`
}

type GetUserListRequest struct {
	Page           int32  `thrift:"page,1" frugal:"1,default,i32" json:"page"`
	PageSize       int32  `thrift:"page_size,2" frugal:"2,default,i32" json:"pageSize"`
	Keyword        string `thrift:"keyword,3" frugal:"3,default,string" json:"keyword"`
	Role           int32  `thrift:"role,4" frugal:"4,default,i32" json:"role"`
	OrganizationId int64  `thrift:"organization_id,5" frugal:"5,default,i64" json:"organizationId"`
}

type GetUserListResponse struct {
	Code    int32   `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string  `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Users   []*User `thrift:"users,3" frugal:"3,default,list<*User>" json:"users,omitempty"`
	Total   int64   `thrift:"total,4" frugal:"4,default,i64" json:"total"`
}

type CreateUserRequest struct {
	User *User `thrift:"user,1" frugal:"1,default,*User" json:"user,omitempty"`
}

type CreateUserResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
	UserId  int64  `thrift:"user_id,3" frugal:"3,default,i64" json:"userId"`
}

type UpdateUserRequest struct {
	User *User `thrift:"user,1" frugal:"1,default,*User" json:"user,omitempty"`
}

type UpdateUserResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type DeleteUserRequest struct {
	UserId int64 `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
}

type DeleteUserResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type ChangePasswordRequest struct {
	UserId      int64  `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
	OldPassword string `thrift:"old_password,2" frugal:"2,default,string" json:"oldPassword"`
	NewPassword string `thrift:"new_password,3" frugal:"3,default,string" json:"newPassword"`
}

type ChangePasswordResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type ResetPasswordRequest struct {
	UserId      int64  `thrift:"user_id,1" frugal:"1,default,i64" json:"userId"`
	NewPassword string `thrift:"new_password,2" frugal:"2,default,string" json:"newPassword"`
}

type ResetPasswordResponse struct {
	Code    int32  `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message string `thrift:"message,2" frugal:"2,default,string" json:"message"`
}

type GetOrganizationTreeRequest struct {
	OrganizationId int64 `thrift:"organization_id,1" frugal:"1,default,i64" json:"organizationId"`
}

type GetOrganizationTreeResponse struct {
	Code          int32           `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message       string          `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Organizations []*Organization `thrift:"organizations,3" frugal:"3,default,list<*Organization>" json:"organizations,omitempty"`
}

type GetMediatorListRequest struct {
	OrganizationId int64 `thrift:"organization_id,1" frugal:"1,default,i64" json:"organizationId"`
	SpecialtyId    int64 `thrift:"specialty_id,2" frugal:"2,default,i64" json:"specialtyId"`
}

type GetMediatorListResponse struct {
	Code      int32   `thrift:"code,1" frugal:"1,default,i32" json:"code"`
	Message   string  `thrift:"message,2" frugal:"2,default,string" json:"message"`
	Mediators []*User `thrift:"mediators,3" frugal:"3,default,list<*User>" json:"mediators,omitempty"`
}
