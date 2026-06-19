package trtc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/database"
	"github.com/dispute-resolve/common/logger"
	"github.com/dispute-resolve/common/mq"
	"github.com/dispute-resolve/common/utils"
	"go.uber.org/zap"
)

type VideoQueueService struct {
	cfg       *config.TRTCConfig
	queueKey  string
}

var videoQueueService *VideoQueueService

func InitVideoQueueService() {
	cfg := config.GetConfig()
	videoQueueService = &VideoQueueService{
		cfg:      &cfg.TRTC,
		queueKey: "video:mediation:queue",
	}
	logger.Info("Video queue service initialized")
}

func GetVideoQueueService() *VideoQueueService {
	if videoQueueService == nil {
		InitVideoQueueService()
	}
	return videoQueueService
}

type QueueItem struct {
	CaseID       int64  `json:"caseId"`
	CaseNo       string `json:"caseNo"`
	CaseTitle    string `json:"caseTitle"`
	MediatorID   int64  `json:"mediatorId"`
	MediatorName string `json:"mediatorName"`
	PartyName    string `json:"partyName"`
	PartyPhone   string `json:"partyPhone"`
	PartyUserID  int64  `json:"partyUserId"`
	Priority     int    `json:"priority"`
	EnqueueTime  int64  `json:"enqueueTime"`
}

func (s *VideoQueueService) Enqueue(ctx context.Context, item *QueueItem) (int, error) {
	rdb := database.GetRedisClient()

	queueLen, err := rdb.ZCard(ctx, s.queueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("get queue length failed: %w", err)
	}

	maxSize := s.cfg.MaxQueueSize
	if maxSize == 0 {
		maxSize = 50
	}

	if int(queueLen) >= maxSize {
		return 0, fmt.Errorf("排队已满，请稍后再试")
	}

	data, _ := json.Marshal(item)
	score := float64(item.Priority)*1e12 + float64(item.EnqueueTime)

	err = rdb.ZAdd(ctx, s.queueKey, fmt.Sprintf("%d:%s", item.CaseID, item.PartyUserID), float64(score)).Err()
	if err != nil {
		return 0, fmt.Errorf("enqueue failed: %w", err)
	}

	cacheKey := fmt.Sprintf("video:queue:item:%d:%d", item.CaseID, item.PartyUserID)
	rdb.Set(ctx, cacheKey, string(data), 24*time.Hour)

	position, _ := s.GetPosition(ctx, item.CaseID, item.PartyUserID)

	s.notifyQueueStatus(item.PartyPhone, item.PartyName, position)

	queueRecord := map[string]interface{}{
		"id":             utils.GenerateID(),
		"case_id":        item.CaseID,
		"case_no":        item.CaseNo,
		"mediator_id":    item.MediatorID,
		"mediator_name":  item.MediatorName,
		"party_name":     item.PartyName,
		"party_phone":    item.PartyPhone,
		"party_user_id":  item.PartyUserID,
		"priority":       item.Priority,
		"status":         1,
		"enqueue_time":   time.Unix(item.EnqueueTime, 0),
	}
	database.GetDB().Table("video_queue").Create(queueRecord)

	logger.Info("User enqueued",
		zap.Int64("caseId", item.CaseID),
		zap.Int64("userId", item.PartyUserID),
		zap.Int("position", position),
	)

	return position, nil
}

func (s *VideoQueueService) Dequeue(ctx context.Context) (*QueueItem, error) {
	rdb := database.GetRedisClient()

	results, err := rdb.ZRange(ctx, s.queueKey, 0, 0).Result()
	if err != nil || len(results) == 0 {
		return nil, nil
	}

	member := results[0]
	err = rdb.ZRem(ctx, s.queueKey, member).Err()
	if err != nil {
		return nil, fmt.Errorf("dequeue failed: %w", err)
	}

	cacheKey := fmt.Sprintf("video:queue:item:%s", member)
	data, err := rdb.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, fmt.Errorf("get queue item data failed: %w", err)
	}
	rdb.Del(ctx, cacheKey)

	var item QueueItem
	if err := json.Unmarshal([]byte(data), &item); err != nil {
		return nil, fmt.Errorf("unmarshal queue item failed: %w", err)
	}

	database.GetDB().Table("video_queue").
		Where("case_id = ? AND party_user_id = ? AND status = 1", item.CaseID, item.PartyUserID).
		Updates(map[string]interface{}{
			"status":      2,
			"dequeue_time": time.Now(),
		})

	return &item, nil
}

func (s *VideoQueueService) GetPosition(ctx context.Context, caseID, userID int64) (int, error) {
	rdb := database.GetRedisClient()

	member := fmt.Sprintf("%d:%d", caseID, userID)
	rank, err := rdb.ZRank(ctx, s.queueKey, member).Result()
	if err != nil {
		return -1, nil
	}

	return int(rank) + 1, nil
}

func (s *VideoQueueService) GetQueueList(ctx context.Context) ([]*QueueItem, error) {
	rdb := database.GetRedisClient()

	members, err := rdb.ZRange(ctx, s.queueKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var items []*QueueItem
	for _, member := range members {
		cacheKey := fmt.Sprintf("video:queue:item:%s", member)
		data, err := rdb.Get(ctx, cacheKey).Result()
		if err != nil {
			continue
		}

		var item QueueItem
		if err := json.Unmarshal([]byte(data), &item); err != nil {
			continue
		}
		items = append(items, &item)
	}

	return items, nil
}

func (s *VideoQueueService) RemoveFromQueue(ctx context.Context, caseID, userID int64) error {
	rdb := database.GetRedisClient()

	member := fmt.Sprintf("%d:%d", caseID, userID)
	err := rdb.ZRem(ctx, s.queueKey, member).Err()
	if err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("video:queue:item:%s", member)
	rdb.Del(ctx, cacheKey)

	database.GetDB().Table("video_queue").
		Where("case_id = ? AND party_user_id = ? AND status = 1", caseID, userID).
		Updates(map[string]interface{}{
			"status":      3,
			"dequeue_time": time.Now(),
		})

	return nil
}

func (s *VideoQueueService) CheckAndNotify(ctx context.Context) error {
	rdb := database.GetRedisClient()

	activeRoomCount, err := rdb.Get(ctx, "video:active_rooms:count").Int()
	if err != nil {
		activeRoomCount = 0
	}

	if activeRoomCount < 5 {
		item, err := s.Dequeue(ctx)
		if err != nil {
			return err
		}
		if item != nil {
			s.notifyUserEnter(item.PartyPhone, item.PartyName, item.CaseID, item.CaseNo)

			s.notifyNextInQueue(ctx)
		}
	}

	return nil
}

func (s *VideoQueueService) notifyQueueStatus(phone, name string, position int) {
	msg := map[string]interface{}{
		"type":       "video_queue",
		"channel":    "sms",
		"phone":      phone,
		"templateCode": "VIDEO_QUEUE_NOTIFY",
		"params": map[string]interface{}{
			"name":     name,
			"position": position,
		},
		"timestamp": time.Now(),
	}
	mq.SendAsync("dispute_notification", msg)
}

func (s *VideoQueueService) notifyUserEnter(phone, name string, caseID int64, caseNo string) {
	msg := map[string]interface{}{
		"type":       "video_queue",
		"channel":    "sms",
		"phone":      phone,
		"templateCode": "VIDEO_QUEUE_ENTER",
		"params": map[string]interface{}{
			"name":    name,
			"caseNo":  caseNo,
			"caseId":  caseID,
		},
		"timestamp": time.Now(),
	}
	mq.SendAsync("dispute_notification", msg)
}

func (s *VideoQueueService) notifyNextInQueue(ctx context.Context) {
	items, err := s.GetQueueList(ctx)
	if err != nil || len(items) == 0 {
		return
	}

	for i, item := range items {
		position := i + 1
		if position <= 3 {
			msg := map[string]interface{}{
				"type":       "video_queue",
				"channel":    "sms",
				"phone":      item.PartyPhone,
				"templateCode": "VIDEO_QUEUE_POSITION",
				"params": map[string]interface{}{
					"name":     item.PartyName,
					"position": position,
				},
				"timestamp": time.Now(),
			}
			mq.SendAsync("dispute_notification", msg)
		}
	}
}

func (s *VideoQueueService) IncrementActiveRoomCount(ctx context.Context) {
	rdb := database.GetRedisClient()
	rdb.Incr(ctx, "video:active_rooms:count")
	rdb.Expire(ctx, "video:active_rooms:count", 24*time.Hour)
}

func (s *VideoQueueService) DecrementActiveRoomCount(ctx context.Context) {
	rdb := database.GetRedisClient()
	rdb.Decr(ctx, "video:active_rooms:count")
}
