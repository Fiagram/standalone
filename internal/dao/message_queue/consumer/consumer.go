package dao_message_queue

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/Fiagram/standalone/internal/configs"
	"github.com/Fiagram/standalone/internal/logger"
)

type HandlerFunc func(ctx context.Context, queueName string, payload []byte) error

type consumerHandler struct {
	handlerFunc HandlerFunc
}

var _ sarama.ConsumerGroupHandler = (*consumerHandler)(nil)

func newConsumerHandler(
	handlerFunc HandlerFunc,
) *consumerHandler {
	return &consumerHandler{
		handlerFunc: handlerFunc,
	}
}

func (h consumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h consumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := h.handlerFunc(session.Context(), message.Topic, message.Value); err != nil {
			return err
		}
	}
	session.Commit()
	return nil
}

type Consumer interface {
	RegisterHandler(queueName string, handlerFunc HandlerFunc)
	Start(ctx context.Context) error
}

type consumer struct {
	saramaConsumer            sarama.ConsumerGroup
	logger                    *zap.Logger
	queueNameToHandlerFuncMap map[string]HandlerFunc
}

func NewDaoMessageQueueConsumer(
	mqConfig configs.MessageQueue,
	logger *zap.Logger,
) (Consumer, error) {
	saramaConsumer, err := sarama.NewConsumerGroup(mqConfig.Addresses, mqConfig.ClientID, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}

	return &consumer{
		saramaConsumer:            saramaConsumer,
		logger:                    logger,
		queueNameToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}

func newSaramaConfig(mqConfig configs.MessageQueue) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true
	return saramaConfig
}

func (c *consumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) {
	c.queueNameToHandlerFuncMap[queueName] = handlerFunc
}

func (c consumer) Start(ctx context.Context) error {
	logger := logger.LoggerWithContext(ctx, c.logger)

	for queueName, handlerFunc := range c.queueNameToHandlerFuncMap {
		go func(queueName string, handlerFunc HandlerFunc) {
			if err := c.saramaConsumer.Consume(
				ctx,
				[]string{queueName},
				newConsumerHandler(handlerFunc),
			); err != nil {
				logger.
					With(zap.String("queue_name", queueName)).
					With(zap.Error(err)).
					Error("failed to consume message from queue")
			}
		}(queueName, handlerFunc)
	}

	<-ctx.Done()
	return nil
}
