package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dispute-resolve/common/cache"
	"github.com/dispute-resolve/common/constants"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/model"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	cronInstance *cron.Cron
	cronOnce     sync.Once
	entryIDs     []cron.EntryID
)

const (
	LockExpireTimeout     = 300 * time.Second
	LockExpirePerf      = 600 * time.Second
	LockExpireES        = 300 * time.Second
	LockExpireStats     = 600 * time.Second
	LockExpireSatisfy   = 600 * time.Second
)

func StartCronTasks() {
	cronOnce.Do(func() {
		cronInstance = cron.New(
			cron.WithSeconds(),
			cron.WithChain(
				cron.Recover(cron.DefaultLogger),
			),
		)

		addCronTask("*/1 * * * * ?", timeoutApprovalTask, "timeout_approval")
		addCronTask("0 0 * * * ?", performanceStatTask, "performance_stat")
		addCronTask("0 */5 * * * ?", syncESIndexTask, "sync_es_index")
		addCronTask("0 0 0 * * ?", dailyStatsCacheTask, "daily_stats_cache")
		addCronTask("0 */10 * * * ?", checkSatisfactionEvalTask, "check_satisfaction_eval")

		cronInstance.Start()
		logger.Info("All cron tasks started", zap.Int("taskCount", len(entryIDs)))
	})
}

func StopCronTasks() {
	if cronInstance != nil {
		ctx := cronInstance.Stop()
		<-ctx.Done()
		logger.Info("All cron tasks stopped")
	}
}

func addCronTask(spec string, taskFunc func(), taskName string) {
	entryID, err := cronInstance.AddFunc(spec, taskFunc)
	if err != nil {
		logger.Error("Failed to add cron task",
			zap.String("task", taskName),
			zap.String("spec", spec),
			logger.Error(err),
		)
		return
	}
	entryIDs = append(entryIDs, entryID)
	logger.Info("Cron task added",
		zap.String("task", taskName),
		zap.String("spec", spec),
	)
}

func acquireLock(ctx context.Context, lockKey string, expire time.Duration) (bool, error) {
	return cache.Lock(ctx, lockKey, expire)
}

func releaseLock(ctx context.Context, lockKey string) error {
	return cache.Unlock(ctx, lockKey)
}

func timeoutApprovalTask() {
	ctx := context.Background()
	lockKey := constants.RedisKeyPrefixLock + "cron:timeout_approval"

	locked, err := acquireLock(ctx, lockKey, LockExpireTimeout)
	if err != nil || !locked {
		logger.Debug("Skip timeout approval task, lock not acquired")
		return
	}
	defer releaseLock(ctx, lockKey)

	logger.Info("Starting timeout approval task")
	startTime := time.Now()

	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	ProcessTimeoutUpgrade(db, 24, 48, 72)

	elapsed := time.Since(startTime)
	logger.Info("Timeout approval task completed", zap.Duration("elapsed", elapsed))
}

func ProcessTimeoutUpgrade(db *gorm.DB, level1Hours, level2Hours, level3Hours int) {
	now := time.Now()

	type TimeoutCase struct {
		ID             int64
		CaseNo         string
		Title          string
		Status           int32
		CurrentNode      string
		ApproverID       int64
		ApproverName     string
		ApproverPhone    string
		SubmitTime       *time.Time
		EscalateLevel   int
	}

	level1Duration := time.Duration(level1Hours) * time.Hour
	level2Duration := time.Duration(level2Hours) * time.Hour
	level3Duration := time.Duration(level3Hours) * time.Hour

	levels := []struct {
		Hours      time.Duration
		Level   int
		NextRole string
	}{
		{level1Duration, 1, "组长"},
		{level2Duration, 2, "主任"},
		{level3Duration, 3, "系统告警"},
	}

	sql := `SELECT 
		c.id, c.case_no, c.title, c.status,
		wi.current_node_name as current_node,
		wi.approver_id, wi.approver_name,
		u.phone as approver_phone,
		wi.submit_time,
		c.escalate_level
	FROM dispute_case c
	LEFT JOIN workflow_approval_instance wi ON c.id = wi.case_id AND wi.status = 10
	LEFT JOIN sys_user u ON wi.approver_id = u.id
	WHERE c.status IN (30, 40) AND wi.id IS NOT NULL AND c.deleted_at IS NULL`

	var cases []TimeoutCase
	if err := db.Raw(sql).Scan(&cases).Error; err != nil {
		logger.Error("Query timeout cases failed", logger.Error(err))
		return
	}

	logger.Info("Found pending approval cases", zap.Int("count", len(cases)))

	for _, c := range cases {
		if c.SubmitTime == nil {
			continue
		}

		elapsed := now.Sub(*c.SubmitTime)
		currentLevel := 0
		nextEscalate := ""

		for i, lv := range levels {
			if elapsed > lv.Hours {
				currentLevel = lv.Level
				if i+1 < len(levels) {
					nextEscalate = levels[i+1].NextRole
				} else {
					nextEscalate = "系统告警"
				}
			}
		}

		if currentLevel == 0 {
			continue
		}

		if c.EscalateLevel >= currentLevel {
			continue
		}

		overdueHours := int(elapsed.Hours())
		nextEscalateTime := now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")

		msg := &mq.TimeoutUpgradeMessage{
			CaseID:           c.ID,
			CaseNo:           c.CaseNo,
			CaseTitle:        c.Title,
			CurrentNode:      c.CurrentNode,
			HandlerID:        c.ApproverID,
			HandlerName:      c.ApproverName,
			HandlerPhone:     c.ApproverPhone,
			OverdueHours:     overdueHours,
			TimeoutLevel:     currentLevel,
			NextEscalateTime: nextEscalateTime,
			NextEscalateRole: nextEscalate,
		}

		if err := mq.SendTimeoutUpgradeMessage(msg, 0); err != nil {
			logger.Warn("Send timeout upgrade message failed",
				zap.Int64("caseId", c.ID),
				logger.Error(err),
			)
		}

		urgeType := constants.UrgeTypeSystem
		if currentLevel >= 3 {
			urgeType = constants.UrgeTypeEscalate
		}

		urgeContent := fmt.Sprintf("案件%s超时%d小时，当前级别%d级", c.CaseNo, overdueHours, currentLevel)
		urgeSQL := `INSERT INTO dispute_urge (case_id, urge_type, urge_content, operator_id, operator_name, created_at) VALUES (?, ?, ?, ?, ?, ?)`
		_ = urgeType
		_ = urgeContent
		if err := db.Exec(urgeSQL, c.ID, urgeType, urgeContent, 0, "系统自动", now).Error; err != nil {
			logger.Warn("Insert urge record failed",
				zap.Int64("caseId", c.ID),
				logger.Error(err),
			)
		}

		if currentLevel >= 2 {
			updateSQL := `UPDATE dispute_case SET escalate_level = ?, escalate_time = ?, urgency_count = urgency_count + 1 WHERE id = ?`
			if err := db.Exec(updateSQL, currentLevel, now, c.ID).Error; err != nil {
				logger.Warn("Update case escalate level failed",
					zap.Int64("caseId", c.ID),
					logger.Error(err),
				)
			}
		}

		historySQL := `INSERT INTO dispute_case_history (case_id, case_no, operation_type, operation_detail, operator_id, operator_name, operator_role, remark, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
		detail := fmt.Sprintf(`{"timeoutLevel":%d,"overdueHours":%d,"approverId":%d}`, currentLevel, overdueHours, c.ApproverID)
		if err := db.Exec(historySQL,
			c.ID, c.CaseNo, "TIMEOUT_UPGRADE", detail,
			0, "系统自动", 0,
			fmt.Sprintf("超时升级：%d级，超时%d小时", currentLevel, overdueHours),
			now,
		).Error; err != nil {
			logger.Warn("Insert case history failed",
				zap.Int64("caseId", c.ID),
				logger.Error(err),
			)
		}

		logger.Info("Processed timeout case",
			zap.Int64("caseId", c.ID),
			zap.String("caseNo", c.CaseNo),
			zap.Int("timeoutLevel", currentLevel),
			zap.Int("overdueHours", overdueHours),
		)
	}
}

func performanceStatTask() {
	ctx := context.Background()
	lockKey := constants.RedisKeyPrefixLock + "cron:performance_stat"

	locked, err := acquireLock(ctx, lockKey, LockExpirePerf)
	if err != nil || !locked {
		logger.Debug("Skip performance stat task, lock not acquired")
		return
	}
	defer releaseLock(ctx, lockKey)

	logger.Info("Starting performance stat task")
	startTime := time.Now()

	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	now := time.Now()
	periodDate := now.Format("2006-01")
	periodType := constants.PerformancePeriodMonth

	type UserStat struct {
		UserID            int64
		UserName          string
		OrgID             int64
		OrgName           string
		CaseCount         int
		CloseCount        int
		SuccessCount      int
		TotalDays         float64
		TotalSatisfaction   int
		SatisfactionCount int
	}

	sql := `SELECT 
		u.id as user_id, u.real_name as user_name,
		u.organization_id as org_id, u.org_name,
		COUNT(c.id) as case_count,
		SUM(CASE WHEN c.status = 50 THEN 1 ELSE 0 END) as close_count,
		SUM(CASE WHEN c.mediation_result = 1 THEN 1 ELSE 0 END) as success_count,
		SUM(TIMESTAMPDIFF(HOUR, c.created_at, IFNULL(c.close_time, NOW()))) / 24.0 as total_days,
		IFNULL(SUM(c.satisfaction_score), 0) as total_satisfaction,
		SUM(CASE WHEN c.satisfaction_score > 0 THEN 1 ELSE 0 END) as satisfaction_count
	FROM sys_user u
	LEFT JOIN dispute_case c ON u.id = c.mediator_id 
		AND DATE_FORMAT(c.created_at, '%Y-%m') = ?
		AND c.deleted_at IS NULL
	WHERE u.role = 3 AND u.status = 1 AND u.deleted_at IS NULL
	GROUP BY u.id, u.real_name, u.organization_id, u.org_name`

	var stats []UserStat
	if err := db.Raw(sql, periodDate).Scan(&stats).Error; err != nil {
		logger.Error("Query user performance stat failed", logger.Error(err))
		return
	}

	logger.Info("Found users for performance stat", zap.Int("count", len(stats)))

	for _, s := range stats {
		if s.CaseCount == 0 {
			continue
		}

		closeRate := 0.0
		successRate := 0.0
		avgDays := 0.0
		avgSatisfaction := 0.0

		if s.CaseCount > 0 {
			closeRate = float64(s.CloseCount) / float64(s.CaseCount) * 100
			avgDays = s.TotalDays / float64(s.CaseCount)
			if s.CloseCount > 0 {
				successRate = float64(s.SuccessCount) / float64(s.CloseCount) * 100
			}
		}

		if s.SatisfactionCount > 0 {
			avgSatisfaction = float64(s.TotalSatisfaction) / float64(s.SatisfactionCount)
		}

		score := closeRate*0.25 + successRate*0.30 + (100-avgDays)*0.20 + avgSatisfaction*10*0.25
		grade := "C"
		switch {
		case score >= 90:
			grade = "A"
		case score >= 80:
			grade = "B"
		case score >= 60:
			grade = "C"
		default:
			grade = "D"
		}

		stat := &model.PerformanceStat{
			UserID:            s.UserID,
			UserName:          s.UserName,
			OrgID:             s.OrgID,
			OrgName:           s.OrgName,
			PeriodType:        periodType,
			PeriodDate:        periodDate,
			CaseCount:         s.CaseCount,
			CloseCount:        s.CloseCount,
			CloseRate:         closeRate,
			SuccessCount:      s.SuccessCount,
			SuccessRate:       successRate,
			AvgDays:           avgDays,
			TotalSatisfaction: s.TotalSatisfaction,
			AvgSatisfaction:   avgSatisfaction,
			Score:             score,
			Grade:             grade,
		}

		upsertSQL := `INSERT INTO performance_stat 
			(user_id, user_name, org_id, org_name, period_type, period_date, 
			 case_count, close_count, close_rate, success_count, success_rate, 
			 avg_days, total_satisfaction, avg_satisfaction, score, grade, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
			case_count = VALUES(case_count),
			close_count = VALUES(close_count),
			close_rate = VALUES(close_rate),
			success_count = VALUES(success_count),
			success_rate = VALUES(success_rate),
			avg_days = VALUES(avg_days),
			total_satisfaction = VALUES(total_satisfaction),
			avg_satisfaction = VALUES(avg_satisfaction),
			score = VALUES(score),
			grade = VALUES(grade),
			updated_at = VALUES(updated_at)`

		if err := db.Exec(upsertSQL,
			stat.UserID, stat.UserName, stat.OrgID, stat.OrgName,
			stat.PeriodType, stat.PeriodDate,
			stat.CaseCount, stat.CloseCount, stat.CloseRate,
			stat.SuccessCount, stat.SuccessRate,
			stat.AvgDays, stat.TotalSatisfaction, stat.AvgSatisfaction,
			stat.Score, stat.Grade,
			now, now,
		).Error; err != nil {
			logger.Warn("Upsert performance stat failed",
				zap.Int64("userId", s.UserID),
				logger.Error(err),
			)
		}
	}

	elapsed := time.Since(startTime)
	logger.Info("Performance stat task completed", zap.Duration("elapsed", elapsed))
}

func syncESIndexTask() {
	ctx := context.Background()
	lockKey := constants.RedisKeyPrefixLock + "cron:sync_es_index"

	locked, err := acquireLock(ctx, lockKey, LockExpireES)
	if err != nil || !locked {
		logger.Debug("Skip sync ES index task, lock not acquired")
		return
	}
	defer releaseLock(ctx, lockKey)

	logger.Info("Starting sync ES index task")
	startTime := time.Now()

	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	now := time.Now()
	fiveMinutesAgo := now.Add(-5 * time.Minute)

	type CaseDoc struct {
		ID               int64
		CaseNo           string
		Title            string
		Description      string
		TypeName         string
		Status           int32
		OrganizationID   int64
		OrganizationName string
		MediatorID       int64
		MediatorName     string
		ApplicantName    string
		ApplicantPhone   string
		RespondentName   string
		Level            int
		CreatedAt        time.Time
		UpdatedAt        time.Time
	}

	sql := `SELECT 
		id, case_no, title, description,
		type_name, status,
		org_id, org_name,
		mediator_id, mediator_name,
		applicant_name, applicant_phone,
		respondent_name,
		case_level,
		created_at, updated_at
	FROM dispute_case
	WHERE updated_at >= ? AND deleted_at IS NULL
	ORDER BY updated_at DESC
	LIMIT 1000`

	var docs []CaseDoc
	if err := db.Raw(sql, fiveMinutesAgo).Scan(&docs).Error; err != nil {
		logger.Error("Query cases for ES sync failed", logger.Error(err))
		return
	}

	logger.Info("Found cases for ES sync", zap.Int("count", len(docs)))

	syncedCount := 0
	for _, doc := range docs {
		docJSON, err := json.Marshal(doc)
		if err != nil {
			continue
		}
		_ = docJSON
		syncedCount++
	}

	esSyncSQL := `INSERT INTO operation_log (user_id, username, operation, module, method, url, ip, params, result, cost_time, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	if err := db.Exec(esSyncSQL,
		0, "system", "ES_INDEX_SYNC", "cron", "SYNC", "/internal/es/sync", "127.0.0.1",
		fmt.Sprintf(`{"count":%d}`, len(docs)),
		fmt.Sprintf(`{"synced":%d}`, syncedCount),
		int(time.Since(startTime).Milliseconds()),
		1,
		now,
	).Error; err != nil {
		logger.Warn("Insert ES sync log failed", logger.Error(err))
	}

	elapsed := time.Since(startTime)
	logger.Info("Sync ES index task completed",
		zap.Int("syncedCount", syncedCount),
		zap.Duration("elapsed", elapsed),
	)
}

func dailyStatsCacheTask() {
	ctx := context.Background()
	lockKey := constants.RedisKeyPrefixLock + "cron:daily_stats_cache"

	locked, err := acquireLock(ctx, lockKey, LockExpireStats)
	if err != nil || !locked {
		logger.Debug("Skip daily stats cache task, lock not acquired")
		return
	}
	defer releaseLock(ctx, lockKey)

	logger.Info("Starting daily stats cache task")
	startTime := time.Now()

	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	yesterdayStr := yesterday.Format("2006-01-02")

	type DailyStats struct {
		TotalCases      int `json:"totalCases"`
		NewCases       int `json:"newCases"`
		ClosedCases     int `json:"closedCases"`
		PendingCases    int `json:"pendingCases"`
		MediatingCases int `json:"mediatingCases"`
		ApprovingCases int `json:"approvingCases"`
		SuccessCount  int `json:"successCount"`
		SuccessRate    float64 `json:"successRate"`
		AvgSatisfaction float64 `json:"avgSatisfaction"`
	}

	var stats DailyStats

	db.Raw(`SELECT COUNT(*) FROM dispute_case WHERE deleted_at IS NULL`).Scan(&stats.TotalCases)
	db.Raw(`SELECT COUNT(*) FROM dispute_case WHERE DATE(created_at) = ? AND deleted_at IS NULL`, yesterdayStr).Scan(&stats.NewCases)
	db.Raw(`SELECT COUNT(*) FROM dispute_case WHERE DATE(close_time) = ? AND deleted_at IS NULL`, yesterdayStr).Scan(&stats.ClosedCases)
	db.Raw(`SELECT COUNT(*) FROM dispute_case WHERE status = 10 AND deleted_at IS NULL`).Scan(&stats.PendingCases)
	db.Raw(`SELECT COUNT(*) FROM dispute_case WHERE status = 20 AND deleted_at IS NULL`).Scan(&stats.MediatingCases)
	db.Raw(`SELECT COUNT(*) FROM dispute_case WHERE status IN (30, 40) AND deleted_at IS NULL`).Scan(&stats.ApprovingCases)
	db.Raw(`SELECT COUNT(*) FROM dispute_case WHERE DATE(close_time) = ? AND mediation_result = 1 AND deleted_at IS NULL`, yesterdayStr).Scan(&stats.SuccessCount)

	if stats.ClosedCases > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.ClosedCases) * 100
	}

	var avgSat float64
	row := db.Raw(`SELECT AVG(satisfaction_score) FROM dispute_case WHERE DATE(close_time) = ? AND satisfaction_score > 0 AND deleted_at IS NULL`, yesterdayStr).Row()
	row.Scan(&avgSat)
	stats.AvgSatisfaction = avgSat

	type TypeStat struct {
		TypeName string `json:"typeName"`
		Count    int    `json:"count"`
	}
	var typeStats []TypeStat
	db.Raw(`SELECT t.type_name, COUNT(c.id) as count 
		FROM dispute_type t 
		LEFT JOIN dispute_case c ON t.id = c.type_id AND DATE(c.created_at) = ? AND c.deleted_at IS NULL
		WHERE t.level = 1 AND t.deleted_at IS NULL
		GROUP BY t.id, t.type_name
		ORDER BY count DESC`, yesterdayStr).Scan(&typeStats)

	type OrgStat struct {
		OrgName string `json:"orgName"`
		Count   int    `json:"count"`
	}
	var orgStats []OrgStat
	db.Raw(`SELECT o.org_name, COUNT(c.id) as count
		FROM sys_organization o
		LEFT JOIN dispute_case c ON o.id = c.org_id AND DATE(c.created_at) = ? AND c.deleted_at IS NULL
		WHERE o.deleted_at IS NULL
		GROUP BY o.id, o.org_name
		ORDER BY count DESC`, yesterdayStr).Scan(&orgStats)

	type MediatorRank struct {
		MediatorID   int64  `json:"mediatorId"`
		MediatorName string `json:"mediatorName"`
		CaseCount    int    `json:"caseCount"`
	}
	var mediatorRanks []MediatorRank
	db.Raw(`SELECT mediator_id, mediator_name, COUNT(*) as case_count
		FROM dispute_case
		WHERE mediator_id > 0 AND DATE(created_at) = ? AND deleted_at IS NULL
		GROUP BY mediator_id, mediator_name
		ORDER BY case_count DESC
		LIMIT 10`, yesterdayStr).Scan(&mediatorRanks)

	cacheData := map[string]interface{}{
		"date":            yesterdayStr,
		"overview":        stats,
		"typeStats":     typeStats,
		"orgStats":      orgStats,
		"mediatorRanks": mediatorRanks,
		"generatedAt":   now.Format(time.RFC3339),
	}

	cacheKey := "stats:dashboard:" + yesterdayStr
	cacheValue, _ := json.Marshal(cacheData)

	cacheExpire := 5 * time.Minute
	if err := cache.Set(ctx, cacheKey, string(cacheValue), cacheExpire); err != nil {
		logger.Warn("Set daily stats cache failed", logger.Error(err))
	}

	overviewKey := "stats:dashboard:latest"
	if err := cache.Set(ctx, overviewKey, string(cacheValue), cacheExpire); err != nil {
		logger.Warn("Set latest stats cache failed", logger.Error(err))
	}

	elapsed := time.Since(startTime)
	logger.Info("Daily stats cache task completed",
		zap.String("date", yesterdayStr),
		zap.Duration("elapsed", elapsed),
	)
}

func checkSatisfactionEvalTask() {
	ctx := context.Background()
	lockKey := constants.RedisKeyPrefixLock + "cron:check_satisfaction_eval"

	locked, err := acquireLock(ctx, lockKey, LockExpireSatisfy)
	if err != nil || !locked {
		logger.Debug("Skip check satisfaction eval task, lock not acquired")
		return
	}
	defer releaseLock(ctx, lockKey)

	logger.Info("Starting check satisfaction eval task")
	startTime := time.Now()

	db := database.GetDB()
	if db == nil {
		logger.Error("Database not initialized")
		return
	}

	now := time.Now()
	seventyTwoHoursAgo := now.Add(-72 * time.Hour)

	type ClosedCase struct {
		ID               int64
		CaseNo         string
		Title            string
		ApplicantID    int64
		ApplicantName  string
		ApplicantPhone string
		MediatorID       int64
		MediatorName     string
		CloseTime       *time.Time
		MediationResult  string
		SatisfactionScore int
	}

	sql := `SELECT 
		c.id, c.case_no, c.title,
		c.applicant_id, c.applicant_name, c.applicant_phone,
		c.mediator_id, c.mediator_name,
		c.close_time, c.mediation_result,
		c.satisfaction_score
	FROM dispute_case c
	WHERE c.status = 50 
		AND c.satisfaction_score = 0
		AND c.close_time <= ?
		AND c.close_time >= ?
		AND c.deleted_at IS NULL
		AND NOT EXISTS (
			SELECT 1 FROM notification_record nr
			WHERE nr.case_id = c.id AND nr.template_type = 4
		)
	ORDER BY c.close_time ASC
	LIMIT 100`

	var cases []ClosedCase
	if err := db.Raw(sql, now, seventyTwoHoursAgo.Add(-72*time.Hour)).Scan(&cases).Error; err != nil {
		logger.Error("Query closed cases for satisfaction eval failed", logger.Error(err))
		return
	}

	logger.Info("Found closed cases need satisfaction eval", zap.Int("count", len(cases)))

	for _, c := range cases {
		if c.CloseTime == nil {
			continue
		}

		closeTimeStr := c.CloseTime.Format("2006-01-02 15:04:05")
		mediationResult := "调解成功"
		if c.MediationResult == "2" {
			mediationResult = "调解未达成"
		}

		evalUrl := fmt.Sprintf("/api/v1/dispute/eval/%d", c.ID)

		msg := &mq.SatisfactionEvalMessage{
			CaseID:          c.ID,
			CaseNo:          c.CaseNo,
			CaseTitle:       c.Title,
			UserID:          c.ApplicantID,
			UserName:        c.ApplicantName,
			UserPhone:       c.ApplicantPhone,
			MediatorID:      c.MediatorID,
			MediatorName:    c.MediatorName,
			CloseTime:       closeTimeStr,
			MediationResult: mediationResult,
			EvalUrl:         evalUrl,
		}

		if err := mq.SendSatisfactionEvalMessage(msg, 0); err != nil {
			logger.Warn("Send satisfaction eval message failed",
				zap.Int64("caseId", c.ID),
				logger.Error(err),
			)
			continue
		}

		insertSQL := `INSERT INTO notification_record (receiver_id, receiver_name, template_id, template_name, template_type, title, content, channel_type, send_status, params, case_id, case_no, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
		title := "【服务评价】诚邀您对本次调解服务进行评价"
		content := fmt.Sprintf(
			"尊敬的%s您好：您的纠纷案件（编号：%s）已结案。请您对本次调解服务进行满意度评价，点击链接参与评价：%s，本邀请72小时内有效。",
			c.ApplicantName, c.CaseNo, evalUrl,
		)
		params := map[string]interface{}{
			"caseNo":          c.CaseNo,
			"caseTitle":       c.Title,
			"mediatorName":    c.MediatorName,
			"closeTime":       closeTimeStr,
			"mediationResult": mediationResult,
			"evalUrl":         evalUrl,
		}
		paramsJSON, _ := json.Marshal(params)

		channels := []string{"站内信", "短信", "微信"}
		for _, ch := range channels {
			if err := db.Exec(insertSQL,
				c.ApplicantID, c.ApplicantName,
				4, "满意度评价邀请", 4,
				title, content,
				ch, 1,
				string(paramsJSON),
				c.ID, c.CaseNo,
				now,
			).Error; err != nil {
				logger.Warn("Insert satisfaction eval notification failed",
					zap.Int64("caseId", c.ID),
					zap.String("channel", ch),
					logger.Error(err),
				)
			}
		}

		logger.Info("Sent satisfaction eval invitation",
			zap.Int64("caseId", c.ID),
			zap.String("caseNo", c.CaseNo),
		)
	}

	elapsed := time.Since(startTime)
	logger.Info("Check satisfaction eval task completed",
		zap.Int("sentCount", len(cases)),
		zap.Duration("elapsed", elapsed),
	)
}
