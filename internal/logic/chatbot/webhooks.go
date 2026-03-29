package logic_chatbot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type WebhooksLogic interface {
	HandleCreatedWebhookSignal(ctx context.Context, signal CreatedWebhookSignal)
}

type webhooksLogic struct {
	webhookAsor dao_database.ChatbotWebhookAccessor
	logger      *zap.Logger
}

func NewWebhooksLogic(
	webhookAsor dao_database.ChatbotWebhookAccessor,
	logger *zap.Logger,
) WebhooksLogic {
	return &webhooksLogic{
		webhookAsor: webhookAsor,
		logger:      logger,
	}
}

func (l webhooksLogic) HandleCreatedWebhookSignal(ctx context.Context, signal CreatedWebhookSignal) {
	logger := logger.LoggerWithContext(ctx, l.logger).With(
		zap.Uint64("of_webhook_id", signal.OfWebhookId),
	)

	webhook, err := l.webhookAsor.GetWebhook(ctx, signal.OfWebhookId)
	if err != nil {
		logger.Error("failed to get webhook by id", zap.Error(err))
		return
	}

	payload := map[string]any{
		"content": fmt.Sprintf("Webhook \"%s\" has been registered successfully.", webhook.Name),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		logger.Error("failed to marshal webhook payload", zap.Error(err))
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhook.Url, bytes.NewReader(body))
	if err != nil {
		logger.Error("failed to create HTTP request", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error("failed to send webhook notification", zap.Error(err))
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		logger.Warn("webhook endpoint returned non-success status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("url", webhook.Url),
		)
		return
	}
}
