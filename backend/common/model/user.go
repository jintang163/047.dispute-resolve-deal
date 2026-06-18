package model

import "time"

type User struct {
	BaseModel
	Username       string `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Password       string `gorm:"size:100;not null" json:"-"`
	RealName       string `gorm:"size:50;not null" json:"realName"`
	Phone          string `gorm:"size:20;index" json:"phone"`
	Email          string `gorm:"size:100" json:"email"`
	IDCard         string `gorm:"size:18;index" json:"idCard"`
	Avatar         string `gorm:"size:255" json:"avatar"`
	Role           int    `gorm:"not null;index" json:"role"`
	RoleCode       string `gorm:"size:50" json:"roleCode"`
	OrganizationID int64  `gorm:"column:org_id;index" json:"organizationId"`
	OrganizationName string `gorm:"column:org_name;size:100" json:"organizationName"`
	OpenID         string `gorm:"size:100;index" json:"openId"`
	Specialty      string `gorm:"size:500" json:"specialty"`
	Description    string `gorm:"size:500" json:"description"`
	Status         int32  `gorm:"default:1;index" json:"status"`
	LastLoginAt    *time.Time `json:"lastLoginAt"`
	LastLoginIP    string `gorm:"size:50" json:"lastLoginIp"`
	LoginCount     int64  `gorm:"default:0" json:"loginCount"`
	CreatedBy      int64  `json:"createdBy"`
	UpdatedBy      int64  `json:"updatedBy"`

	OrgID   int64  `gorm:"-" json:"orgId"`
	OrgName string `gorm:"-" json:"orgName"`
}

func (u *User) AfterFind(tx interface{}) error {
	u.OrgID = u.OrganizationID
	u.OrgName = u.OrganizationName
	return nil
}

func (User) TableName() string {
	return "sys_user"
}
