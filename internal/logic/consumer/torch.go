package logic_consumer

import (
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

type TorchLogic interface {
	Setup(sarama.ConsumerGroupSession) error
	Cleanup(sarama.ConsumerGroupSession) error
	ConsumeClaim(sarama.ConsumerGroupSession, sarama.ConsumerGroupClaim) error
}

var TorchLogicTopic = "torch"

var _ sarama.ConsumerGroupHandler = (*torchLogic)(nil)

type torchLogic struct {
	logger *zap.Logger
}

func NewTorchLogic(
	logger *zap.Logger,
) TorchLogic {
	return &torchLogic{
		logger: logger,
	}
}

func (l *torchLogic) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (l *torchLogic) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (l *torchLogic) ConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				fmt.Println("message channel closed")
				return nil
			}

			fmt.Printf(
				"topic=%s partition=%d offset=%d key=%s value=%s",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				string(msg.Key),
				string(msg.Value),
			)

			// Mark message as processed so the offset can be committed.
			session.MarkMessage(msg, "")

		case <-session.Context().Done():
			// Important during rebalance/shutdown.
			return nil
		}
	}
}
