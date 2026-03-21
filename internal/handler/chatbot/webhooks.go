package handler

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

type WebhookServer interface {
	Start(ctx context.Context) error
}

type webhookServer struct {
	createdWebhookChan CreatedWebhookChan
	webhookAccessor    dao_database.ChatbotWebhookAccessor
	logger             *zap.Logger
}

func NewWebhookServer(
	createdWebhookChan CreatedWebhookChan,
	webhookAccessor dao_database.ChatbotWebhookAccessor,
	logger *zap.Logger,
) WebhookServer {
	return &webhookServer{
		createdWebhookChan: createdWebhookChan,
		webhookAccessor:    webhookAccessor,
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
		case signal := <-w.createdWebhookChan:
			w.handleCreatedWebhookSignal(ctx, signal)
		}
	}
}

func (w *webhookServer) handleCreatedWebhookSignal(ctx context.Context, signal CreatedWebhookSignal) {
	logger := logger.LoggerWithContext(ctx, w.logger).With(
		zap.Uint64("of_webhook_id", signal.OfWebhookId),
	)

	webhook, err := w.webhookAccessor.GetWebhook(ctx, signal.OfWebhookId)
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
