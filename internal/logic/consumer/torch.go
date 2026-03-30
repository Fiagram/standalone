package logic_consumer

import (
	"encoding/json"
	"fmt"

	logic_chatbot "github.com/Fiagram/standalone/internal/logic/chatbot"
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
	torchSignalChan chan logic_chatbot.TorchSignal
	logger          *zap.Logger
}

func NewTorchLogic(
	torchSignalChan chan logic_chatbot.TorchSignal,
	logger *zap.Logger,
) TorchLogic {
	return &torchLogic{
		torchSignalChan: torchSignalChan,
		logger:          logger,
	}
}

func (l torchLogic) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (l torchLogic) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (l torchLogic) ConsumeClaim(
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

			var incomingTorchSignal logic_chatbot.TorchSignal
			if err := json.Unmarshal(msg.Value, &incomingTorchSignal); err != nil {
				l.logger.Error("failed to unmarshal torch signal", zap.Error(err))
				continue
			}

			l.torchSignalChan <- incomingTorchSignal

			// Mark message as processed so the offset can be committed.
			session.MarkMessage(msg, "")

		case <-session.Context().Done():
			// Important during rebalance/shutdown.
			return nil
		}
	}
}
