package dao_database

import (
	"context"
	"errors"
	"time"

	"github.com/Fiagram/standalone/internal/logger"
	"go.uber.org/zap"
)

type ChatbotWebhook struct {
	Id          uint64    `json:"id"`
	OfAccountId uint64    `json:"of_account_id"`
	Name        string    `json:"name"`
	Url         string    `json:"url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ChatbotWebhookAccessor interface {
	CreateWebhook(ctx context.Context, webhook ChatbotWebhook) (uint64, error)

	GetWebhook(ctx context.Context, id uint64) (ChatbotWebhook, error)
	GetWebhooksAll(ctx context.Context) ([]ChatbotWebhook, error)
	GetWebhooksByAccountId(ctx context.Context, accountId uint64, limit, offset int) ([]ChatbotWebhook, error)

	UpdateWebhook(ctx context.Context, webhook ChatbotWebhook) error

	DeleteWebhook(ctx context.Context, id uint64) error

	WithExecutor(exec Executor) ChatbotWebhookAccessor
}

type chatbotWebhookAccessor struct {
	exec   Executor
	logger *zap.Logger
}

func NewChatbotWebhookAccessor(
	exec Executor,
	logger *zap.Logger,
) ChatbotWebhookAccessor {
	return &chatbotWebhookAccessor{
		exec:   exec,
		logger: logger,
	}
}

func (a chatbotWebhookAccessor) CreateWebhook(
	ctx context.Context,
	webhook ChatbotWebhook,
) (uint64, error) {
	if webhook.OfAccountId == 0 || webhook.Name == "" || webhook.Url == "" {
		return 0, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("webhook", webhook))
	const query = `INSERT INTO chatbot_webhooks
			(of_account_id, name, url)
			VALUES (?, ?, ?)`
	result, err := a.exec.ExecContext(ctx, query,
		webhook.OfAccountId,
		webhook.Name,
		webhook.Url,
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to create webhook")
		return 0, err
	}

	rowEfNum, err := result.RowsAffected()
	if rowEfNum != 1 || err != nil {
		errMsg := "failed to effect row"
		logger.With(zap.Int64("rowEfNum", rowEfNum)).
			With(zap.Error(err)).
			Error(errMsg)
		return 0, errors.New(errMsg)
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get last inserted id")
		return 0, err
	}

	return uint64(lastInsertedId), nil
}

func (a chatbotWebhookAccessor) GetWebhook(
	ctx context.Context,
	id uint64,
) (ChatbotWebhook, error) {
	if id == 0 {
		return ChatbotWebhook{}, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Uint64("webhook_id", id))
	const query = `SELECT id, of_account_id, name, url, created_at, updated_at
			FROM chatbot_webhooks WHERE id = ?`
	row := a.exec.QueryRowContext(ctx, query, id)

	var out ChatbotWebhook
	err := row.Scan(
		&out.Id,
		&out.OfAccountId,
		&out.Name,
		&out.Url,
		&out.CreatedAt,
		&out.UpdatedAt,
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get webhook by id")
		return ChatbotWebhook{}, err
	}

	return out, nil
}

func (a chatbotWebhookAccessor) GetWebhooksAll(
	ctx context.Context,
) ([]ChatbotWebhook, error) {
	logger := logger.LoggerWithContext(ctx, a.logger)
	const query = `SELECT id, of_account_id, name, url, created_at, updated_at
			FROM chatbot_webhooks ORDER BY id ASC`
	rows, err := a.exec.QueryContext(ctx, query)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get all webhooks")
		return nil, err
	}
	defer rows.Close()

	var webhooks []ChatbotWebhook
	for rows.Next() {
		var item ChatbotWebhook
		err = rows.Scan(
			&item.Id,
			&item.OfAccountId,
			&item.Name,
			&item.Url,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to scan webhook row")
			return nil, err
		}
		webhooks = append(webhooks, item)
	}

	return webhooks, nil
}

func (a chatbotWebhookAccessor) GetWebhooksByAccountId(
	ctx context.Context,
	accountId uint64,
	limit, offset int,
) ([]ChatbotWebhook, error) {
	if accountId == 0 {
		return nil, ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Uint64("account_id", accountId))
	const query = `SELECT id, of_account_id, name, url, created_at, updated_at
			FROM chatbot_webhooks WHERE of_account_id = ?
			ORDER BY id ASC LIMIT ? OFFSET ?`
	rows, err := a.exec.QueryContext(ctx, query, accountId, limit, offset)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to get webhooks by account id")
		return nil, err
	}
	defer rows.Close()

	var webhooks []ChatbotWebhook
	for rows.Next() {
		var item ChatbotWebhook
		err = rows.Scan(
			&item.Id,
			&item.OfAccountId,
			&item.Name,
			&item.Url,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			logger.With(zap.Error(err)).Error("failed to scan webhook row")
			return nil, err
		}
		webhooks = append(webhooks, item)
	}

	return webhooks, nil
}

func (a chatbotWebhookAccessor) UpdateWebhook(
	ctx context.Context,
	webhook ChatbotWebhook,
) error {
	if webhook.Id == 0 {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Any("webhook", webhook))
	const query = `UPDATE chatbot_webhooks SET
			name = ?, url = ?
			WHERE id = ?`

	result, err := a.exec.ExecContext(ctx, query,
		webhook.Name,
		webhook.Url,
		webhook.Id,
	)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to update webhook")
		return err
	}

	rowEfNum, err := result.RowsAffected()
	if rowEfNum != 1 || err != nil {
		errMsg := "failed to effect row"
		logger.With(zap.Int64("rowEfNum", rowEfNum)).
			With(zap.Error(err)).
			Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (a chatbotWebhookAccessor) DeleteWebhook(
	ctx context.Context,
	id uint64,
) error {
	if id == 0 {
		return ErrLackOfInfor
	}

	logger := logger.LoggerWithContext(ctx, a.logger).With(zap.Uint64("webhook_id", id))
	const query = `DELETE FROM chatbot_webhooks WHERE id = ?`
	result, err := a.exec.ExecContext(ctx, query, id)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to delete webhook")
		return err
	}

	rowEfNum, err := result.RowsAffected()
	if rowEfNum != 1 || err != nil {
		errMsg := "failed to effect row"
		logger.With(zap.Int64("rowEfNum", rowEfNum)).
			With(zap.Error(err)).
			Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (a chatbotWebhookAccessor) WithExecutor(
	exec Executor,
) ChatbotWebhookAccessor {
	return &chatbotWebhookAccessor{
		exec:   exec,
		logger: a.logger,
	}
}
