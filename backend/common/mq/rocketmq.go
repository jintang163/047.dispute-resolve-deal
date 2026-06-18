package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/dispute-resolve/common/config"
	"github.com/dispute-resolve/common/logger"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/consumer"
)

var (
	prod   rocketmq.Producer
	cons   rocketmq.PushConsumer
	once   sync.Once
	consOnce sync.Once
)

func InitProducer(cfg *config.RocketMQConfig) {
	once.Do(func() {
		var err error
		prod, err = rocketmq.NewProducer(
			producer.WithGroupName(cfg.GroupName),
			producer.WithNameServer(cfg.NameServer),
			producer.WithRetry(cfg.RetryTimes),
		)
		if err != nil {
			panic(fmt.Errorf("create rocketmq producer failed: %v", err))
		}

		if err = prod.Start(); err != nil {
			panic(fmt.Errorf("start rocketmq producer failed: %v", err))
		}

		logger.Info("RocketMQ producer started")
	})
}

func InitConsumer(cfg *config.RocketMQConfig, groupName string) rocketmq.PushConsumer {
	var err error
	cons, err = rocketmq.NewPushConsumer(
		consumer.WithGroupName(groupName),
		consumer.WithNameServer(cfg.NameServer),
	)
	if err != nil {
		panic(fmt.Errorf("create rocketmq consumer failed: %v", err))
	}
	return cons
}

func SendMessage(topic string, data interface{}, tags ...string) error {
	if prod == nil {
		return fmt.Errorf("rocketmq producer not initialized")
	}

	msgBody, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal message failed: %v", err)
	}

	tag := ""
	if len(tags) > 0 {
		tag = tags[0]
	}

	msg := &primitive.Message{
		Topic: topic,
		Body:  msgBody,
	}
	if tag != "" {
		msg.WithTag(tag)
	}

	msg.WithKeys([]string{fmt.Sprintf("%d", time.Now().UnixNano())})

	ctx := context.Background()
	result, err := prod.SendSync(ctx, msg)
	if err != nil {
		logger.Error("Send rocketmq message failed",
			logger.String("topic", topic),
			logger.String("tag", tag),
			logger.Error(err))
		return err
	}

	logger.Debug("Send rocketmq message success",
		logger.String("topic", topic),
		logger.String("tag", tag),
		logger.String("msgId", result.MsgID),
		logger.Int("queueId", int(result.Queue.QueueId)))

	return nil
}

func SendAsyncMessage(topic string, data interface{}, callback func(context.Context, *primitive.SendResult, error), tags ...string) error {
	if prod == nil {
		return fmt.Errorf("rocketmq producer not initialized")
	}

	msgBody, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal message failed: %v", err)
	}

	tag := ""
	if len(tags) > 0 {
		tag = tags[0]
	}

	msg := &primitive.Message{
		Topic: topic,
		Body:  msgBody,
	}
	if tag != "" {
		msg.WithTag(tag)
	}

	ctx := context.Background()
	err = prod.SendAsync(ctx, callback, msg)
	if err != nil {
		logger.Error("Send async rocketmq message failed",
			logger.String("topic", topic),
			logger.Error(err))
	}

	return err
}

func SendDelayMessage(topic string, data interface{}, delayLevel int, tags ...string) error {
	if prod == nil {
		return fmt.Errorf("rocketmq producer not initialized")
	}

	msgBody, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal message failed: %v", err)
	}

	tag := ""
	if len(tags) > 0 {
		tag = tags[0]
	}

	msg := &primitive.Message{
		Topic: topic,
		Body:  msgBody,
	}
	if tag != "" {
		msg.WithTag(tag)
	}
	msg.WithDelayTimeLevel(delayLevel)

	ctx := context.Background()
	result, err := prod.SendSync(ctx, msg)
	if err != nil {
		logger.Error("Send delay rocketmq message failed",
			logger.String("topic", topic),
			logger.Int("delayLevel", delayLevel),
			logger.Error(err))
		return err
	}

	logger.Debug("Send delay rocketmq message success",
		logger.String("topic", topic),
		logger.Int("delayLevel", delayLevel),
		logger.String("msgId", result.MsgID))

	return nil
}

func ShutdownProducer() {
	if prod != nil {
		prod.Shutdown()
		logger.Info("RocketMQ producer shutdown")
	}
}

func ShutdownConsumer() {
	if cons != nil {
		cons.Shutdown()
		logger.Info("RocketMQ consumer shutdown")
	}
}
