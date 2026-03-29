package dao_message_queue

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/Fiagram/standalone/internal/configs"
	"github.com/Fiagram/standalone/internal/logger"
)

type Producer interface {
	Produce(ctx context.Context, queueName string, payload []byte) error
}

type producer struct {
	saramaSyncProducer sarama.SyncProducer
	logger             *zap.Logger
}

func NewDaoMessageQueueProducer(
	mqConfig configs.MessageQueue,
	logger *zap.Logger,
) (Producer, error) {
	saramaSyncProducer, err := sarama.NewSyncProducer(mqConfig.Addresses, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama sync producer: %w", err)
	}

	return &producer{
		saramaSyncProducer: saramaSyncProducer,
		logger:             logger,
	}, nil
}

func newSaramaConfig(mqConfig configs.MessageQueue) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Retry.Max = 1
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true
	return saramaConfig
}

func (c producer) Produce(ctx context.Context, queueName string, payload []byte) error {
	logger := logger.LoggerWithContext(ctx, c.logger).
		With(zap.String("queue_name", queueName)).
		With(zap.ByteString("payload", payload))

	if _, _, err := c.saramaSyncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: queueName,
		Value: sarama.ByteEncoder(payload),
	}); err != nil {
		logger.With(zap.Error(err)).Error("failed to produce message")
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}
