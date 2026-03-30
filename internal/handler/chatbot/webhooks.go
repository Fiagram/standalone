package handler

import (
	"context"

	"github.com/Fiagram/standalone/internal/logger"
	logic_chatbot "github.com/Fiagram/standalone/internal/logic/chatbot"
	"go.uber.org/zap"
)

type WebhookServer interface {
	Start(ctx context.Context) error
}

type webhookServer struct {
	createdWebhookChan logic_chatbot.CreatedWebhookChan
	torchSignal        chan logic_chatbot.TorchSignal
	webhooksLogic      logic_chatbot.WebhooksLogic
	logger             *zap.Logger
}

func NewWebhookServer(
	createdWebhookChan logic_chatbot.CreatedWebhookChan,
	torchSignal chan logic_chatbot.TorchSignal,
	webhooksLogic logic_chatbot.WebhooksLogic,
	logger *zap.Logger,
) WebhookServer {
	return &webhookServer{
		createdWebhookChan: createdWebhookChan,
		torchSignal:        torchSignal,
		webhooksLogic:      webhooksLogic,
		logger:             logger,
	}
}

func (w *webhookServer) Start(ctx context.Context) error {
	logger := logger.LoggerWithContext(ctx, w.logger)
	logger.Info("WebhookServer started, listening for signals")

	for {
		select {
		case <-ctx.Done():
			logger.Info("WebhookServer stopped")
			return ctx.Err()
		case createdWebhookSignal := <-w.createdWebhookChan:
			w.webhooksLogic.HandleCreatedWebhookSignal(ctx, createdWebhookSignal)
		case torchSignal := <-w.torchSignal:
			w.webhooksLogic.HandleTorchSignal(ctx, torchSignal)
		}
	}
}
