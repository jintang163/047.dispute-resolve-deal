package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type GiftServiceImpl struct{}

func NewGiftService() service.GiftService {
	return &GiftServiceImpl{}
}

func (s *GiftServiceImpl) GetGiftList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("gift g").
		Select("g.*, gc.name as category_name")

	if categoryID, ok := query["categoryId"].(float64); ok && int64(categoryID) > 0 {
		db = db.Where("g.category_id = ?", int64(categoryID))
	}
	if status, ok := query["status"].(float64); ok && int(status) >= 0 {
		db = db.Where("g.status = ?", int(status))
	}
	if keyword, ok := query["keyword"].(string); ok && keyword != "" {
		db = db.Where("g.name LIKE ? OR g.gift_no LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if isHot, ok := query["isHot"].(float64); ok && int(isHot) > 0 {
		db = db.Where("g.is_hot = ?", int(isHot))
	}
	if isNew, ok := query["isNew"].(float64); ok && int(isNew) > 0 {
		db = db.Where("g.is_new = ?", int(isNew))
	}
	if minPoints, ok := query["minPoints"].(float64); ok && int(minPoints) > 0 {
		db = db.Where("g.points_required >= ?", int(minPoints))
	}
	if maxPoints, ok := query["maxPoints"].(float64); ok && int(maxPoints) > 0 {
		db = db.Where("g.points_required <= ?", int(maxPoints))
	}

	db = db.Joins("LEFT JOIN gift_category gc ON gc.id = g.category_id")

	var total int64
	db.Count(&total)

	page := intOrDefault(query, "page", 1)
	pageSize := intOrDefault(query, "pageSize", 20)
	offset := (page - 1) * pageSize

	sort := "g.sort_order ASC, g.created_at DESC"
	if sortBy, ok := query["sortBy"].(string); ok && sortBy != "" {
		switch sortBy {
		case "points_asc":
			sort = "g.points_required ASC"
		case "points_desc":
			sort = "g.points_required DESC"
		case "sold_desc":
			sort = "g.sold_count DESC"
		case "created_desc":
			sort = "g.created_at DESC"
		}
	}

	var gifts []map[string]interface{}
	result := db.Order(sort).Offset(offset).Limit(pageSize).Find(&gifts)

	return gifts, total, result.Error
}

func (s *GiftServiceImpl) GetGiftDetail(ctx context.Context, id int64) (map[string]interface{}, error) {
	var detail map[string]interface{}
	result := database.GetDB().Table("gift g").
		Select("g.*, gc.name as category_name").
		Joins("LEFT JOIN gift_category gc ON gc.id = g.category_id").
		Where("g.id = ?", id).
		First(&detail)
	return detail, result.Error
}

func (s *GiftServiceImpl) CreateGift(ctx context.Context, req map[string]interface{}) (int64, error) {
	giftNo := fmt.Sprintf("GF%s", utils.GenerateID())

	gift := model.Gift{
		GiftNo:         giftNo,
		CategoryID:     int64(req["categoryId"].(float64)),
		Name:           req["name"].(string),
		Description:    stringOrDefault(req, "description", ""),
		ImageURL:       stringOrDefault(req, "imageUrl", ""),
		BannerImageURL: stringOrDefault(req, "bannerImageUrl", ""),
		PointsRequired: int(req["pointsRequired"].(float64)),
		OriginalPrice:  floatOrDefault(req, "originalPrice", 0),
		Stock:          intOrDefault(req, "stock", 0),
		ExchangeLimit:  intOrDefault(req, "exchangeLimit", 0),
		Status:         int32(intOrDefault(req, "status", 1)),
		IsHot:          int32(intOrDefault(req, "isHot", 0)),
		IsNew:          int32(intOrDefault(req, "isNew", 0)),
		IsRecommend:    int32(intOrDefault(req, "isRecommend", 0)),
		SortOrder:      intOrDefault(req, "sortOrder", 0),
		Weight:         floatOrDefault(req, "weight", 0),
		Freight:        floatOrDefault(req, "freight", 0),
		VirtualType:    int32(intOrDefault(req, "virtualType", 0)),
	}

	result := database.GetDB().Create(&gift)
	return gift.ID, result.Error
}

func (s *GiftServiceImpl) UpdateGift(ctx context.Context, id int64, req map[string]interface{}) error {
	updates := make(map[string]interface{})
	for k, v := range req {
		updates[utils.CamelToSnake(k)] = v
	}
	return database.GetDB().Table("gift").Where("id = ?", id).Updates(updates).Error
}

func (s *GiftServiceImpl) DeleteGift(ctx context.Context, id int64) error {
	return database.GetDB().Table("gift").Where("id = ?", id).Update("status", 0).Error
}

func (s *GiftServiceImpl) GetGiftCategories(ctx context.Context) ([]map[string]interface{}, error) {
	var categories []map[string]interface{}
	result := database.GetDB().Table("gift_category").
		Where("status = 1").
		Order("sort_order ASC, created_at DESC").
		Find(&categories)

	for i, cat := range categories {
		catID := int64(cat["id"].(float64))
		var count int64
		database.GetDB().Table("gift").
			Where("category_id = ? AND status = 1", catID).
			Count(&count)
		categories[i]["giftCount"] = count
	}

	return categories, result.Error
}

func (s *GiftServiceImpl) CreateGiftCategory(ctx context.Context, req map[string]interface{}) (int64, error) {
	category := model.GiftCategory{
		Name:        req["name"].(string),
		ParentID:    int64OrDefault(req, "parentId", 0),
		IconURL:     stringOrDefault(req, "iconUrl", ""),
		Description: stringOrDefault(req, "description", ""),
		SortOrder:   intOrDefault(req, "sortOrder", 0),
		Status:      int32(intOrDefault(req, "status", 1)),
	}
	result := database.GetDB().Create(&category)
	return category.ID, result.Error
}

func (s *GiftServiceImpl) UpdateGiftCategory(ctx context.Context, id int64, req map[string]interface{}) error {
	updates := make(map[string]interface{})
	for k, v := range req {
		updates[utils.CamelToSnake(k)] = v
	}
	return database.GetDB().Table("gift_category").Where("id = ?", id).Updates(updates).Error
}

func (s *GiftServiceImpl) DeleteGiftCategory(ctx context.Context, id int64) error {
	tx := database.GetDB().Begin()
	tx.Table("gift_category").Where("id = ? OR parent_id = ?", id, id).Update("status", 0)
	tx.Table("gift").Where("category_id IN (SELECT id FROM gift_category WHERE id = ? OR parent_id = ?)", id, id).Update("status", 0)
	tx.Commit()
	return nil
}

func (s *GiftServiceImpl) GetExchangeList(ctx context.Context, query map[string]interface{}) ([]map[string]interface{}, int64, error) {
	db := database.GetDB().Table("gift_exchange ge").
		Select("ge.*, g.name as gift_name, g.image_url as gift_image")

	if memberID, ok := query["memberId"].(float64); ok && int64(memberID) > 0 {
		db = db.Where("ge.member_id = ?", int64(memberID))
	}
	if status, ok := query["status"].(float64); ok && int(status) > 0 {
		db = db.Where("ge.status = ?", int(status))
	}
	if giftID, ok := query["giftId"].(float64); ok && int64(giftID) > 0 {
		db = db.Where("ge.gift_id = ?", int64(giftID))
	}
	if startDate, ok := query["startDate"].(string); ok && startDate != "" {
		db = db.Where("DATE(ge.created_at) >= ?", startDate)
	}
	if endDate, ok := query["endDate"].(string); ok && endDate != "" {
		db = db.Where("DATE(ge.created_at) <= ?", endDate)
	}
	if keyword, ok := query["keyword"].(string); ok && keyword != "" {
		db = db.Where("ge.exchange_no LIKE ? OR ge.member_name LIKE ? OR ge.receiver_name LIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}
	if orgID, ok := query["orgId"].(float64); ok && int64(orgID) > 0 {
		db = db.Where("ge.organization_id = ?", int64(orgID))
	}

	db = db.Joins("LEFT JOIN gift g ON g.id = ge.gift_id")

	var total int64
	db.Count(&total)

	page := intOrDefault(query, "page", 1)
	pageSize := intOrDefault(query, "pageSize", 20)
	offset := (page - 1) * pageSize

	var exchanges []map[string]interface{}
	result := db.Order("ge.created_at DESC").Offset(offset).Limit(pageSize).Find(&exchanges)

	return exchanges, total, result.Error
}

func (s *GiftServiceImpl) GetExchangeDetail(ctx context.Context, id int64) (map[string]interface{}, error) {
	var detail map[string]interface{}
	result := database.GetDB().Table("gift_exchange ge").
		Select("ge.*, g.name as gift_name, g.image_url as gift_image, g.description as gift_description").
		Joins("LEFT JOIN gift g ON g.id = ge.gift_id").
		Where("ge.id = ?", id).
		First(&detail)
	return detail, result.Error
}

func (s *GiftServiceImpl) AuditExchange(ctx context.Context, id int64, status int32, remark string) error {
	tx := database.GetDB().Begin()

	var exchange model.GiftExchange
	result := tx.Where("id = ?", id).First(&exchange)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if exchange.Status != 10 {
		tx.Rollback()
		return fmt.Errorf("exchange is not in pending status")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"status_name":  getExchangeStatusName(int(status)),
		"audit_time":   now,
		"audit_remark": remark,
	}

	if status == 30 {
		updates["status"] = 50
		updates["status_name"] = "已取消"
		updates["cancel_time"] = now
		updates["cancel_reason"] = remark

		pointsService := NewPointsService()
		err := pointsService.AddPoints(ctx, exchange.MemberID, exchange.TotalPoints,
			"exchange_refund", exchange.ExchangeNo,
			fmt.Sprintf("兑换退款: %s", exchange.GiftName))
		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Table("gift").Where("id = ?", exchange.GiftID).Updates(map[string]interface{}{
			"stock":      database.GetDB().Raw("stock + ?", exchange.Quantity),
			"sold_count": database.GetDB().Raw("sold_count - ?", exchange.Quantity),
		})
	}

	err := tx.Table("gift_exchange").Where("id = ?", id).Updates(updates).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (s *GiftServiceImpl) ShipExchange(ctx context.Context, id int64, expressCompany, expressNo string) error {
	return database.GetDB().Table("gift_exchange").Where("id = ?", id).Updates(map[string]interface{}{
		"status":          40,
		"status_name":     "已发货",
		"express_company": expressCompany,
		"express_no":      expressNo,
		"ship_time":       time.Now(),
	}).Error
}

func (s *GiftServiceImpl) ReceiveExchange(ctx context.Context, id int64) error {
	return database.GetDB().Table("gift_exchange").Where("id = ?", id).Updates(map[string]interface{}{
		"status":      60,
		"status_name": "已完成",
		"receive_time": time.Now(),
	}).Error
}

func (s *GiftServiceImpl) CancelExchange(ctx context.Context, id int64, reason string) error {
	tx := database.GetDB().Begin()

	var exchange model.GiftExchange
	result := tx.Where("id = ?", id).First(&exchange)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if exchange.Status >= 40 {
		tx.Rollback()
		return fmt.Errorf("cannot cancel shipped exchange")
	}

	now := time.Now()
	err := tx.Table("gift_exchange").Where("id = ?", id).Updates(map[string]interface{}{
		"status":        50,
		"status_name":   "已取消",
		"cancel_time":   now,
		"cancel_reason": reason,
	}).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	pointsService := NewPointsService()
	err = pointsService.AddPoints(ctx, exchange.MemberID, exchange.TotalPoints,
		"exchange_cancel", exchange.ExchangeNo,
		fmt.Sprintf("取消兑换退款: %s", exchange.GiftName))
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Table("gift").Where("id = ?", exchange.GiftID).Updates(map[string]interface{}{
		"stock":      database.GetDB().Raw("stock + ?", exchange.Quantity),
		"sold_count": database.GetDB().Raw("sold_count - ?", exchange.Quantity),
	})

	tx.Commit()
	return nil
}

func (s *GiftServiceImpl) GetMemberExchanges(ctx context.Context, memberID int64, page, pageSize int) ([]map[string]interface{}, int64, error) {
	query := make(map[string]interface{})
	query["memberId"] = float64(memberID)
	query["page"] = page
	query["pageSize"] = pageSize
	return s.GetExchangeList(ctx, query)
}

func (s *GiftServiceImpl) GetGiftStatistics(ctx context.Context) (map[string]interface{}, error) {
	var totalGifts int64
	database.GetDB().Table("gift").Where("status = 1").Count(&totalGifts)

	var totalExchanges int64
	database.GetDB().Table("gift_exchange").Count(&totalExchanges)

	var totalPoints int64
	database.GetDB().Table("gift_exchange").
		Where("status != 50").
		Select("IFNULL(SUM(total_points), 0)").
		Scan(&totalPoints)

	var pendingCount int64
	database.GetDB().Table("gift_exchange").Where("status = 10").Count(&pendingCount)

	var shippedCount int64
	database.GetDB().Table("gift_exchange").Where("status = 40").Count(&shippedCount)

	categoryStats := make(map[string]int64)
	var catResults []struct {
		CategoryID int64  `gorm:"column:category_id"`
		Count      int64  `gorm:"column:count"`
		Name       string `gorm:"column:name"`
	}
	database.GetDB().Table("gift_exchange ge").
		Select("ge.gift_id, COUNT(*) as count, g.category_id, gc.name").
		Joins("LEFT JOIN gift g ON g.id = ge.gift_id").
		Joins("LEFT JOIN gift_category gc ON gc.id = g.category_id").
		Where("ge.status != 50").
		Group("g.category_id").
		Scan(&catResults)

	for _, r := range catResults {
		if r.Name != "" {
			categoryStats[r.Name] = r.Count
		}
	}

	return map[string]interface{}{
		"totalGifts":     totalGifts,
		"totalExchanges": totalExchanges,
		"totalPoints":    totalPoints,
		"pendingCount":   pendingCount,
		"shippedCount":   shippedCount,
		"categoryStats":  categoryStats,
	}, nil
}

func getExchangeStatusName(status int) string {
	switch status {
	case 10:
		return "待审核"
	case 20:
		return "审核通过"
	case 30:
		return "审核拒绝"
	case 40:
		return "已发货"
	case 50:
		return "已取消"
	case 60:
		return "已完成"
	default:
		return "未知"
	}
}
