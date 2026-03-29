package handler

import (
	"context"
	"fmt"

	"github.com/Fiagram/standalone/internal/configs"
	"github.com/Fiagram/standalone/internal/logger"
	logic_torch "github.com/Fiagram/standalone/internal/logic/consumer"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type Consumer interface {
	Start(ctx context.Context) error
}

type consumer struct {
	consumerGroup sarama.ConsumerGroup
	handlersMap   map[string]sarama.ConsumerGroupHandler
	logger        *zap.Logger
}

func NewConsumer(
	messQueueConfig configs.MessageQueue,
	torchLogic logic_torch.TorchLogic,
	logger *zap.Logger,
) (Consumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = messQueueConfig.ClientID
	saramaConfig.Metadata.Full = true
	consumerGroup, err := sarama.NewConsumerGroup(
		messQueueConfig.Addresses,
		messQueueConfig.ClientID,
		saramaConfig,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}

	// Register handlers from different topics
	handlersMap := make(map[string]sarama.ConsumerGroupHandler)
	handlersMap[logic_torch.TorchLogicTopic] = torchLogic

	return &consumer{
		consumerGroup: consumerGroup,
		handlersMap:   handlersMap,
		logger:        logger,
	}, nil
}

func (c *consumer) Start(ctx context.Context) error {
	logger := logger.LoggerWithContext(ctx, c.logger)

	for topic, handlerFunc := range c.handlersMap {
		go func(topic string, handlerFunc sarama.ConsumerGroupHandler) {
			if err := c.consumerGroup.Consume(
				ctx,
				[]string{topic},
				handlerFunc,
			); err != nil {
				logger.
					With(zap.String("queue_name", topic)).
					With(zap.Error(err)).
					Error("failed to consume message from queue")
			}
		}(topic, handlerFunc)
	}

	<-ctx.Done()
	return nil
}
