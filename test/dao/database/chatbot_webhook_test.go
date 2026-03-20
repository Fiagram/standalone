package dao_database_test

import (
	"context"
	"testing"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"github.com/stretchr/testify/require"
)

// helper: create an account and return its ID + cleanup func
func createTestAccount(t *testing.T) (uint64, func()) {
	t.Helper()
	aAsor := dao_database.NewAccountAccessor(sqlDb, logger)
	acc := RandomAccount()
	id, err := aAsor.CreateAccount(context.Background(), acc)
	require.NoError(t, err)
	require.NotZero(t, id)
	return id, func() {
		_ = aAsor.DeleteAccount(context.Background(), id)
	}
}

// helper: create a webhook under the given account and return its ID
func createTestWebhook(t *testing.T, accountId uint64, name, url string) uint64 {
	t.Helper()
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	id, err := wAsor.CreateWebhook(context.Background(), dao_database.ChatbotWebhook{
		OfAccountId: accountId,
		Name:        name,
		Url:         url,
	})
	require.NoError(t, err)
	require.NotZero(t, id)
	return id
}

// --- CreateWebhook ---

func TestCreateWebhook(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	id, err := wAsor.CreateWebhook(context.Background(), dao_database.ChatbotWebhook{
		OfAccountId: accountId,
		Name:        "test-webhook",
		Url:         "https://example.com/hook",
	})

	require.NoError(t, err)
	require.NotZero(t, id)

	// cleanup
	_ = wAsor.DeleteWebhook(context.Background(), id)
}

func TestCreateWebhook_MissingAccountId(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	_, err := wAsor.CreateWebhook(context.Background(), dao_database.ChatbotWebhook{
		Name: "test-webhook",
		Url:  "https://example.com/hook",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

func TestCreateWebhook_MissingName(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	_, err := wAsor.CreateWebhook(context.Background(), dao_database.ChatbotWebhook{
		OfAccountId: 1,
		Url:         "https://example.com/hook",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

func TestCreateWebhook_MissingUrl(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	_, err := wAsor.CreateWebhook(context.Background(), dao_database.ChatbotWebhook{
		OfAccountId: 1,
		Name:        "test-webhook",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

// --- GetWebhook ---

func TestGetWebhook(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	ctx := context.Background()

	webhookId := createTestWebhook(t, accountId, "get-hook", "https://example.com/get")
	defer func() { _ = wAsor.DeleteWebhook(ctx, webhookId) }()

	webhook, err := wAsor.GetWebhook(ctx, webhookId)
	require.NoError(t, err)
	require.Equal(t, webhookId, webhook.Id)
	require.Equal(t, accountId, webhook.OfAccountId)
	require.Equal(t, "get-hook", webhook.Name)
	require.Equal(t, "https://example.com/get", webhook.Url)
	require.NotZero(t, webhook.CreatedAt)
	require.NotZero(t, webhook.UpdatedAt)
}

func TestGetWebhook_ZeroId(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	_, err := wAsor.GetWebhook(context.Background(), 0)

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

func TestGetWebhook_NotFound(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	_, err := wAsor.GetWebhook(context.Background(), 999999999)

	require.Error(t, err)
}

// --- GetWebhooksByAccountId ---

func TestGetWebhooksByAccountId(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	ctx := context.Background()

	id1 := createTestWebhook(t, accountId, "hook-1", "https://example.com/1")
	id2 := createTestWebhook(t, accountId, "hook-2", "https://example.com/2")
	id3 := createTestWebhook(t, accountId, "hook-3", "https://example.com/3")
	defer func() {
		_ = wAsor.DeleteWebhook(ctx, id1)
		_ = wAsor.DeleteWebhook(ctx, id2)
		_ = wAsor.DeleteWebhook(ctx, id3)
	}()

	webhooks, err := wAsor.GetWebhooksByAccountId(ctx, accountId, 20, 0)
	require.NoError(t, err)
	require.Len(t, webhooks, 3)

	// Verify ordering by id ASC
	require.Equal(t, id1, webhooks[0].Id)
	require.Equal(t, id2, webhooks[1].Id)
	require.Equal(t, id3, webhooks[2].Id)
}

func TestGetWebhooksByAccountId_WithLimitOffset(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	ctx := context.Background()

	id1 := createTestWebhook(t, accountId, "hook-a", "https://example.com/a")
	id2 := createTestWebhook(t, accountId, "hook-b", "https://example.com/b")
	id3 := createTestWebhook(t, accountId, "hook-c", "https://example.com/c")
	defer func() {
		_ = wAsor.DeleteWebhook(ctx, id1)
		_ = wAsor.DeleteWebhook(ctx, id2)
		_ = wAsor.DeleteWebhook(ctx, id3)
	}()

	// limit=2, offset=0 → first two
	webhooks, err := wAsor.GetWebhooksByAccountId(ctx, accountId, 2, 0)
	require.NoError(t, err)
	require.Len(t, webhooks, 2)
	require.Equal(t, id1, webhooks[0].Id)
	require.Equal(t, id2, webhooks[1].Id)

	// limit=2, offset=2 → last one
	webhooks2, err := wAsor.GetWebhooksByAccountId(ctx, accountId, 2, 2)
	require.NoError(t, err)
	require.Len(t, webhooks2, 1)
	require.Equal(t, id3, webhooks2[0].Id)
}

func TestGetWebhooksByAccountId_ZeroAccountId(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	_, err := wAsor.GetWebhooksByAccountId(context.Background(), 0, 20, 0)

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

func TestGetWebhooksByAccountId_Empty(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	webhooks, err := wAsor.GetWebhooksByAccountId(context.Background(), accountId, 20, 0)

	require.NoError(t, err)
	require.Empty(t, webhooks)
}

// --- UpdateWebhook ---

func TestUpdateWebhook(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	ctx := context.Background()

	webhookId := createTestWebhook(t, accountId, "old-name", "https://example.com/old")
	defer func() { _ = wAsor.DeleteWebhook(ctx, webhookId) }()

	err := wAsor.UpdateWebhook(ctx, dao_database.ChatbotWebhook{
		Id:   webhookId,
		Name: "new-name",
		Url:  "https://example.com/new",
	})
	require.NoError(t, err)

	updated, err := wAsor.GetWebhook(ctx, webhookId)
	require.NoError(t, err)
	require.Equal(t, "new-name", updated.Name)
	require.Equal(t, "https://example.com/new", updated.Url)
	require.Equal(t, accountId, updated.OfAccountId)
}

func TestUpdateWebhook_ZeroId(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	err := wAsor.UpdateWebhook(context.Background(), dao_database.ChatbotWebhook{
		Name: "name",
		Url:  "https://example.com",
	})

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

// --- DeleteWebhook ---

func TestDeleteWebhook(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	ctx := context.Background()

	webhookId := createTestWebhook(t, accountId, "to-delete", "https://example.com/del")

	err := wAsor.DeleteWebhook(ctx, webhookId)
	require.NoError(t, err)

	// Verify it's gone
	_, err = wAsor.GetWebhook(ctx, webhookId)
	require.Error(t, err)
}

func TestDeleteWebhook_ZeroId(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	err := wAsor.DeleteWebhook(context.Background(), 0)

	require.Error(t, err)
	require.ErrorIs(t, err, dao_database.ErrLackOfInfor)
}

func TestDeleteWebhook_NotFound(t *testing.T) {
	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	err := wAsor.DeleteWebhook(context.Background(), 999999999)

	require.Error(t, err)
}

// --- WithExecutor ---

func TestWebhookWithExecutor(t *testing.T) {
	accountId, cleanupAcc := createTestAccount(t)
	defer cleanupAcc()

	wAsor := dao_database.NewChatbotWebhookAccessor(sqlDb, logger)
	wAsor2 := wAsor.WithExecutor(sqlDb)
	require.NotNil(t, wAsor2)

	ctx := context.Background()
	webhookId := createTestWebhook(t, accountId, "exec-hook", "https://example.com/exec")
	defer func() { _ = wAsor.DeleteWebhook(ctx, webhookId) }()

	// Verify the new accessor works
	webhook, err := wAsor2.GetWebhook(ctx, webhookId)
	require.NoError(t, err)
	require.Equal(t, "exec-hook", webhook.Name)
}
