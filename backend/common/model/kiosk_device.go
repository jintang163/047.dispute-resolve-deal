package model

import "time"

type KioskDevice struct {
	BaseModel
	DeviceNo      string     `gorm:"size:50;uniqueIndex;not null" json:"deviceNo"`
	DeviceName    string     `gorm:"size:100;not null" json:"deviceName"`
	OrgID         int64      `gorm:"index" json:"orgId"`
	OrgName       string     `gorm:"size:100" json:"orgName"`
	Location      string     `gorm:"size:255" json:"location"`
	Status        int32      `gorm:"default:1;index" json:"status"`
	LastHeartbeat *time.Time `json:"lastHeartbeat"`
	IPAddress     string     `gorm:"size:50" json:"ipAddress"`
	Longitude     float64    `json:"longitude"`
	Latitude      float64    `json:"latitude"`
	BindUserID    int64      `gorm:"index" json:"bindUserId"`
	CreatedBy     int64      `json:"createdBy"`
}

type Kiosk = KioskDevice

func (KioskDevice) TableName() string {
	return "kiosk_device"
}
