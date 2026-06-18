package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type BaseQuery struct {
	Page     int    `form:"page,default=1" json:"page"`
	PageSize int    `form:"pageSize,default=10" json:"pageSize"`
	Keyword  string `form:"keyword" json:"keyword"`
	SortBy   string `form:"sortBy" json:"sortBy"`
	SortOrder string `form:"sortOrder,default=desc" json:"sortOrder"`
}

type DateRangeQuery struct {
	StartTime string `form:"startTime" json:"startTime"`
	EndTime   string `form:"endTime" json:"endTime"`
}

type IDListRequest struct {
	IDs []int64 `json:"ids" binding:"required"`
}

type StatusRequest struct {
	ID     int64 `json:"id" binding:"required"`
	Status int32 `json:"status" binding:"required"`
}

type UpdateStatusRequest struct {
	IDs    []int64 `json:"ids" binding:"required"`
	Status int32   `json:"status" binding:"required"`
}

func (q *BaseQuery) GetOffset() int {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	return (q.Page - 1) * q.PageSize
}

func (q *BaseQuery) GetLimit() int {
	if q.PageSize <= 0 {
		q.PageSize = 10
	}
	return q.PageSize
}

func (q *BaseQuery) GetSort() string {
	if q.SortBy == "" {
		return "id DESC"
	}
	return q.SortBy + " " + q.SortOrder
}
