package model

type LawArticle struct {
	BaseModel
	Title     string `gorm:"size:200;not null" json:"title"`
	Content   string `gorm:"type:text;not null" json:"content"`
	Category  string `gorm:"size:100;index" json:"category"`
	LawName   string `gorm:"size:100;index" json:"lawName"`
	ArticleNo string `gorm:"size:50" json:"articleNo"`
	Keywords  string `gorm:"size:500" json:"keywords"`
	VectorID  string `gorm:"size:100;index" json:"vectorId"`
	Status    int32  `gorm:"default:1;index" json:"status"`
	CreatedBy int64  `json:"createdBy"`
}

func (LawArticle) TableName() string {
	return "law_article"
}
