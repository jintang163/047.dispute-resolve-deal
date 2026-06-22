package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/model"
	"github.com/dispute-resolve/common/utils"
	"github.com/dispute-resolve/gateway/service"
)

type PointsServiceImpl struct{}

func NewPointsService() service.PointsService {
	return &PointsServiceImpl{}
}

func (s *PointsServiceImpl) GetPointsSummary(ctx context.Context, memberID int64) (map[string]interface{}, error) {
	var account model.GridMemberPointsAccount
	result := database.GetDB().Where("member_id = ?", memberID).First(&account)
	if result.Error != nil || account.ID == 0 {
		account = model.GridMemberPointsAccount{
			MemberID:        memberID,
			Balance:         0,
			TotalEarned:     0,
			TotalSpent:      0,
			Level:           1,
			LevelName:       "初级网格员",
			NextLevelPoints: 1000,
		}
		database.GetDB().Create(&account)
	}

	today := time.Now().Format("2006-01-02")
	monthStart := time.Now().Format("2006-01-01")

	var todayEarned int64
	database.GetDB().Table("points_record").
		Where("member_id = ? AND type = 1 AND DATE(created_at) = ?", memberID, today).
		Select("IFNULL(SUM(points), 0)").
		Scan(&todayEarned)

	var monthEarned int64
	database.GetDB().Table("points_record").
		Where("member_id = ? AND type = 1 AND DATE(created_at) >= ?", memberID, monthStart).
		Select("IFNULL(SUM(points), 0)").
		Scan(&monthEarned)

	levelProgress := 0.0
	if account.NextLevelPoints > 0 {
		levelProgress = float64(account.TotalEarned) / float64(account.NextLevelPoints) * 100
		if levelProgress > 100 {
			levelProgress = 100
		}
	}

	return map[string]interface{}{
		"memberId":        account.MemberID,
		"memberName":      account.MemberName,
		"balance":         account.Balance,
		"todayEarned":     todayEarned,
		"monthEarned":     monthEarned,
		"totalEarned":     account.TotalEarned,
		"totalSpent":      account.TotalSpent,
		"level":           account.Level,
		"levelName":       account.LevelName,
		"nextLevelPoints": account.NextLevelPoints,
		"levelProgress":   levelProgress,
		"checkinDays":     account.CheckinDays,
	}, nil
}

func (s *PointsServiceImpl) AddPoints(ctx context.Context, memberID int64, points int, businessType, businessNo string, description string) error {
	if points <= 0 {
		return nil
	}

	tx := database.GetDB().Begin()

	var account model.GridMemberPointsAccount
	result := tx.Where("member_id = ?", memberID).First(&account)
	if result.Error != nil || account.ID == 0 {
		var member model.GridMember
		database.GetDB().Where("id = ?", memberID).First(&member)
		account = model.GridMemberPointsAccount{
			MemberID:        memberID,
			MemberName:      member.RealName,
			MemberNo:        member.MemberNo,
			OrganizationID:  member.OrganizationID,
			Balance:         points,
			TotalEarned:     points,
			Level:           1,
			LevelName:       "初级网格员",
			NextLevelPoints: 1000,
		}
		if err := tx.Create(&account).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		updates := map[string]interface{}{
			"balance":      account.Balance + points,
			"total_earned": account.TotalEarned + points,
		}

		newTotal := account.TotalEarned + points
		level := 1
		levelName := "初级网格员"
		nextLevelPoints := 1000

		switch {
		case newTotal >= 10000:
			level = 4
			levelName = "专家网格员"
			nextLevelPoints = 100000
		case newTotal >= 5000:
			level = 3
			levelName = "高级网格员"
			nextLevelPoints = 10000
		case newTotal >= 1000:
			level = 2
			levelName = "中级网格员"
			nextLevelPoints = 5000
		}

		updates["level"] = level
		updates["level_name"] = levelName
		updates["next_level_points"] = nextLevelPoints

		if err := tx.Model(&account).Updates(updates).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	recordNo := fmt.Sprintf("PR%s", utils.GenerateID())
	expireDate := time.Now().AddDate(1, 0, 0)

	record := model.PointsRecord{
		RecordNo:      recordNo,
		MemberID:      memberID,
		MemberName:    account.MemberName,
		OrganizationID: account.OrganizationID,
		Type:          1,
		TypeName:      "获得",
		BusinessType:  businessType,
		BusinessNo:    businessNo,
		Points:        points,
		BalanceBefore: account.Balance - points,
		BalanceAfter:  account.Balance,
		Description:   description,
		ExpireDate:    &expireDate,
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return err
	}

	var member model.GridMember
	tx.Model(&member).Where("id = ?", memberID).Updates(map[string]interface{}{
		"points":       account.Balance,
		"total_points": account.TotalEarned,
	})

	tx.Commit()
	return nil
}

func (s *PointsServiceImpl) DeductPoints(ctx context.Context, memberID int64, points int, businessType, businessNo string, description string) error {
	if points <= 0 {
		return nil
	}

	tx := database.GetDB().Begin()

	var account model.GridMemberPointsAccount
	result := tx.Where("member_id = ?", memberID).First(&account)
	if result.Error != nil || account.ID == 0 {
		tx.Rollback()
		return fmt.Errorf("points account not found")
	}

	if account.Balance < points {
		tx.Rollback()
		return fmt.Errorf("insufficient points")
	}

	balanceBefore := account.Balance
	account.Balance -= points
	account.TotalSpent += points

	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		return err
	}

	recordNo := fmt.Sprintf("PR%s", utils.GenerateID())
	record := model.PointsRecord{
		RecordNo:      recordNo,
		MemberID:      memberID,
		MemberName:    account.MemberName,
		OrganizationID: account.OrganizationID,
		Type:          2,
		TypeName:      "消费",
		BusinessType:  businessType,
		BusinessNo:    businessNo,
		Points:        -points,
		BalanceBefore: balanceBefore,
		BalanceAfter:  account.Balance,
		Description:   description,
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return err
	}

	var member model.GridMember
	tx.Model(&member).Where("id = ?", memberID).Updates(map[string]interface{}{
		"points": account.Balance,
	})

	tx.Commit()
	return nil
}

func (s *PointsServiceImpl) GetPointsRecords(ctx context.Context, memberID int64, page, pageSize int) ([]map[string]interface{}, int64, error) {
	query := database.GetDB().Table("points_record").Where("member_id = ?", memberID)

	var total int64
	query.Count(&total)

	var records []map[string]interface{}
	offset := (page - 1) * pageSize
	result := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&records)

	return records, total, result.Error
}

func (s *PointsServiceImpl) GetPointsRules(ctx context.Context, ruleType string) ([]map[string]interface{}, error) {
	query := database.GetDB().Table("points_rule").Where("is_active = 1")
	if ruleType != "" {
		query = query.Where("rule_type = ?", ruleType)
	}

	var rules []map[string]interface{}
	result := query.Order("sort_order ASC").Find(&rules)
	return rules, result.Error
}

func (s *PointsServiceImpl) CreatePointsRule(ctx context.Context, req map[string]interface{}) error {
	rule := model.PointsRule{
		RuleCode:         req["ruleCode"].(string),
		RuleName:         req["ruleName"].(string),
		RuleType:         req["ruleType"].(string),
		Points:           int(req["points"].(float64)),
		MaxPointsPerDay:  int(req["maxPointsPerDay"].(float64)),
		MaxPointsPerMonth: int(req["maxPointsPerMonth"].(float64)),
		IsActive:         int(req["isActive"].(float64)),
		Description:      req["description"].(string),
		ExpireDays:       int(req["expireDays"].(float64)),
		SortOrder:        int(req["sortOrder"].(float64)),
	}
	return database.GetDB().Create(&rule).Error
}

func (s *PointsServiceImpl) UpdatePointsRule(ctx context.Context, id int64, req map[string]interface{}) error {
	updates := make(map[string]interface{})
	for k, v := range req {
		updates[k] = v
	}
	return database.GetDB().Table("points_rule").Where("id = ?", id).Updates(updates).Error
}

func (s *PointsServiceImpl) DeletePointsRule(ctx context.Context, id int64) error {
	return database.GetDB().Delete(&model.PointsRule{}, id).Error
}

func (s *PointsServiceImpl) ExchangeGift(ctx context.Context, memberID int64, giftID int64, quantity int, receiverName, receiverPhone, receiverAddress, remark string) (int64, error) {
	var gift model.Gift
	result := database.GetDB().Where("id = ?", giftID).First(&gift)
	if result.Error != nil {
		return 0, fmt.Errorf("gift not found")
	}

	if gift.Status != 1 {
		return 0, fmt.Errorf("gift is not available")
	}

	if gift.Stock < quantity {
		return 0, fmt.Errorf("insufficient stock")
	}

	totalPoints := gift.PointsRequired * quantity

	tx := database.GetDB().Begin()

	var account model.GridMemberPointsAccount
	result = tx.Where("member_id = ?", memberID).First(&account)
	if result.Error != nil || account.ID == 0 {
		tx.Rollback()
		return 0, fmt.Errorf("points account not found")
	}

	if account.Balance < totalPoints {
		tx.Rollback()
		return 0, fmt.Errorf("insufficient points")
	}

	if gift.ExchangeLimit > 0 {
		var exchangedCount int64
		tx.Table("gift_exchange").
			Where("member_id = ? AND gift_id = ? AND status != 50", memberID, giftID).
			Count(&exchangedCount)
		if int(exchangedCount)+quantity > gift.ExchangeLimit {
			tx.Rollback()
			return 0, fmt.Errorf("exceeded exchange limit")
		}
	}

	balanceBefore := account.Balance
	account.Balance -= totalPoints
	account.TotalSpent += totalPoints
	if err := tx.Save(&account).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	var member model.GridMember
	tx.Where("id = ?", memberID).First(&member)

	exchangeNo := fmt.Sprintf("EX%s", utils.GenerateID())
	exchange := model.GiftExchange{
		ExchangeNo:      exchangeNo,
		MemberID:        memberID,
		MemberName:      member.RealName,
		MemberPhone:     member.Phone,
		OrganizationID:  member.OrganizationID,
		GiftID:          giftID,
		GiftName:        gift.Name,
		GiftImage:       gift.ImageURL,
		GiftPoints:      gift.PointsRequired,
		Quantity:        quantity,
		TotalPoints:     totalPoints,
		ReceiverName:    receiverName,
		ReceiverPhone:   receiverPhone,
		ReceiverAddress: receiverAddress,
		Status:          10,
		Remark:          remark,
	}
	if err := tx.Create(&exchange).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	recordNo := fmt.Sprintf("PR%s", utils.GenerateID())
	record := model.PointsRecord{
		RecordNo:      recordNo,
		MemberID:      memberID,
		MemberName:    member.RealName,
		OrganizationID: member.OrganizationID,
		Type:          2,
		TypeName:      "消费",
		BusinessType:  "gift_exchange",
		BusinessNo:    exchangeNo,
		Points:        -totalPoints,
		BalanceBefore: balanceBefore,
		BalanceAfter:  account.Balance,
		Description:   fmt.Sprintf("兑换礼品: %s x%d", gift.Name, quantity),
	}
	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Model(&gift).Updates(map[string]interface{}{
		"stock":      gift.Stock - quantity,
		"sold_count": gift.SoldCount + quantity,
	})

	tx.Model(&member).Updates(map[string]interface{}{
		"points": account.Balance,
	})

	tx.Commit()

	return exchange.ID, nil
}

func (s *PointsServiceImpl) ProcessExpiredPoints(ctx context.Context) error {
	now := time.Now().Format("2006-01-02")

	var expiredRecords []model.PointsRecord
	database.GetDB().
		Where("type = 1 AND is_expired = 0 AND expire_date <= ? AND points > 0", now).
		Find(&expiredRecords)

	for _, record := range expiredRecords {
		tx := database.GetDB().Begin()

		var account model.GridMemberPointsAccount
		tx.Where("member_id = ?", record.MemberID).First(&account)
		if account.ID > 0 && account.Balance >= record.Points {
			balanceBefore := account.Balance
			account.Balance -= record.Points
			account.TotalExpired += record.Points
			tx.Save(&account)

			recordNo := fmt.Sprintf("PR%s", utils.GenerateID())
			newRecord := model.PointsRecord{
				RecordNo:      recordNo,
				MemberID:      record.MemberID,
				MemberName:    record.MemberName,
				OrganizationID: record.OrganizationID,
				Type:          3,
				TypeName:      "过期",
				BusinessType:  "points_expire",
				BusinessNo:    record.RecordNo,
				Points:        -record.Points,
				BalanceBefore: balanceBefore,
				BalanceAfter:  account.Balance,
				Description:   fmt.Sprintf("积分过期，原流水号: %s", record.RecordNo),
			}
			tx.Create(&newRecord)

			tx.Model(&record).Update("is_expired", 1)

			var member model.GridMember
			tx.Model(&member).Where("id = ?", record.MemberID).Updates(map[string]interface{}{
				"points": account.Balance,
			})
		}

		tx.Commit()
	}

	logger.Info("Processed expired points", logger.Int("count", len(expiredRecords)))
	return nil
}
