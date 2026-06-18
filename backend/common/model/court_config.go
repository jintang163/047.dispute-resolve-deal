package model

import "time"

const (
	CourtLevelSupreme    = 1
	CourtLevelHigh       = 2
	CourtLevelIntermediate = 3
	CourtLevelBasic      = 4
	CourtLevelTribunal   = 5
)

type CourtConfig struct {
	BaseModel
	CourtCode         string `gorm:"size:64;uniqueIndex;not null" json:"courtCode"`
	CourtName         string `gorm:"size:128;not null" json:"courtName"`
	CourtLevel        int32  `gorm:"default:3;index" json:"courtLevel"`
	JurisdictionArea  string `gorm:"size:256" json:"jurisdictionArea"`
	Address           string `gorm:"size:256" json:"address"`
	Contact           string `gorm:"size:64" json:"contact"`
	Phone             string `gorm:"size:20" json:"phone"`

	APIEndpoint    string `gorm:"size:256" json:"apiEndpoint"`
	APIAppID       string `gorm:"size:128" json:"apiAppId"`
	APISecret      string `gorm:"size:256" json:"apiSecret"`
	APIPublicKey   string `gorm:"type:text" json:"apiPublicKey"`

	SealCertNo     string `gorm:"size:64" json:"sealCertNo"`
	SealImageURL   string `gorm:"size:512" json:"sealImageUrl"`

	OrganizationID   int64  `gorm:"column:org_id;index" json:"organizationId"`
	OrganizationName string `gorm:"column:org_name;size:100" json:"organizationName"`
	SortOrder        int    `json:"sortOrder"`
	Status           int32  `gorm:"default:1;index" json:"status"`

	OrgID   int64  `gorm:"-" json:"orgId"`
	OrgName string `gorm:"-" json:"orgName"`
}

func (c *CourtConfig) AfterFind(tx interface{}) error {
	c.OrgID = c.OrganizationID
	c.OrgName = c.OrganizationName
	return nil
}

func (CourtConfig) TableName() string {
	return "court_config"
}
