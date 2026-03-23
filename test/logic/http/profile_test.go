package logic_http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	oapi "github.com/Fiagram/standalone/internal/generated/openapi"
	webhook_handler "github.com/Fiagram/standalone/internal/handler/chatbot"
	logic_account "github.com/Fiagram/standalone/internal/logic/account"
	logic_http "github.com/Fiagram/standalone/internal/logic/http"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---------------------------------------------------------------------------
// Mock: logic_account.Account
// ---------------------------------------------------------------------------

type mockAccountLogic struct {
	getAccountFn            func(ctx context.Context, params logic_account.GetAccountParams) (logic_account.GetAccountOutput, error)
	updateAccountInfoFn     func(ctx context.Context, params logic_account.UpdateAccountInfoParams) (logic_account.UpdateAccountInfoOutput, error)
	checkAccountValidFn     func(ctx context.Context, params logic_account.CheckAccountValidParams) (logic_account.CheckAccountValidOutput, error)
	updateAccountPasswordFn func(ctx context.Context, params logic_account.UpdateAccountPasswordParams) (logic_account.UpdateAccountPasswordOutput, error)
}

func (m *mockAccountLogic) CreateAccount(_ context.Context, _ logic_account.CreateAccountParams) (logic_account.CreateAccountOutput, error) {
	return logic_account.CreateAccountOutput{}, nil
}
func (m *mockAccountLogic) CheckAccountValid(ctx context.Context, params logic_account.CheckAccountValidParams) (logic_account.CheckAccountValidOutput, error) {
	if m.checkAccountValidFn != nil {
		return m.checkAccountValidFn(ctx, params)
	}
	return logic_account.CheckAccountValidOutput{}, nil
}
func (m *mockAccountLogic) IsUsernameTaken(_ context.Context, _ logic_account.IsUsernameTakenParams) (logic_account.IsUsernameTakenOutput, error) {
	return logic_account.IsUsernameTakenOutput{}, nil
}
func (m *mockAccountLogic) GetAccount(ctx context.Context, params logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
	if m.getAccountFn != nil {
		return m.getAccountFn(ctx, params)
	}
	return logic_account.GetAccountOutput{}, nil
}
func (m *mockAccountLogic) GetAccountAll(_ context.Context, _ logic_account.GetAccountAllParams) (logic_account.GetAccountAllOutput, error) {
	return logic_account.GetAccountAllOutput{}, nil
}
func (m *mockAccountLogic) GetAccountList(_ context.Context, _ logic_account.GetAccountListParams) (logic_account.GetAccountListOutput, error) {
	return logic_account.GetAccountListOutput{}, nil
}
func (m *mockAccountLogic) UpdateAccountInfo(ctx context.Context, params logic_account.UpdateAccountInfoParams) (logic_account.UpdateAccountInfoOutput, error) {
	if m.updateAccountInfoFn != nil {
		return m.updateAccountInfoFn(ctx, params)
	}
	return logic_account.UpdateAccountInfoOutput{}, nil
}
func (m *mockAccountLogic) UpdateAccountPassword(ctx context.Context, params logic_account.UpdateAccountPasswordParams) (logic_account.UpdateAccountPasswordOutput, error) {
	if m.updateAccountPasswordFn != nil {
		return m.updateAccountPasswordFn(ctx, params)
	}
	return logic_account.UpdateAccountPasswordOutput{}, nil
}
func (m *mockAccountLogic) DeleteAccount(_ context.Context, _ logic_account.DeleteAccountParams) error {
	return nil
}
func (m *mockAccountLogic) DeleteAccountByUsername(_ context.Context, _ logic_account.DeleteAccountByUsernameParams) error {
	return nil
}

// ---------------------------------------------------------------------------
// Mock: dao_database.ChatbotWebhookAccessor
// ---------------------------------------------------------------------------

type mockWebhookAccessor struct {
	createWebhookFn        func(ctx context.Context, webhook dao_database.ChatbotWebhook) (uint64, error)
	getWebhookFn           func(ctx context.Context, id uint64) (dao_database.ChatbotWebhook, error)
	getWebhooksByAccountFn func(ctx context.Context, accountId uint64, limit, offset int) ([]dao_database.ChatbotWebhook, error)
	updateWebhookFn        func(ctx context.Context, webhook dao_database.ChatbotWebhook) error
	deleteWebhookFn        func(ctx context.Context, id uint64) error
}

func (m *mockWebhookAccessor) CreateWebhook(ctx context.Context, webhook dao_database.ChatbotWebhook) (uint64, error) {
	if m.createWebhookFn != nil {
		return m.createWebhookFn(ctx, webhook)
	}
	return 0, nil
}
func (m *mockWebhookAccessor) GetWebhook(ctx context.Context, id uint64) (dao_database.ChatbotWebhook, error) {
	if m.getWebhookFn != nil {
		return m.getWebhookFn(ctx, id)
	}
	return dao_database.ChatbotWebhook{}, nil
}
func (m *mockWebhookAccessor) GetWebhooksByAccountId(ctx context.Context, accountId uint64, limit, offset int) ([]dao_database.ChatbotWebhook, error) {
	if m.getWebhooksByAccountFn != nil {
		return m.getWebhooksByAccountFn(ctx, accountId, limit, offset)
	}
	return nil, nil
}
func (m *mockWebhookAccessor) UpdateWebhook(ctx context.Context, webhook dao_database.ChatbotWebhook) error {
	if m.updateWebhookFn != nil {
		return m.updateWebhookFn(ctx, webhook)
	}
	return nil
}
func (m *mockWebhookAccessor) DeleteWebhook(ctx context.Context, id uint64) error {
	if m.deleteWebhookFn != nil {
		return m.deleteWebhookFn(ctx, id)
	}
	return nil
}
func (m *mockWebhookAccessor) WithExecutor(_ dao_database.Executor) dao_database.ChatbotWebhookAccessor {
	return m
}

// ---------------------------------------------------------------------------
// Mock: dao_database.AccountRoleAccessor
// ---------------------------------------------------------------------------

type mockAccountRoleAccessor struct {
	getRoleByIdFn        func(ctx context.Context, id uint8) (dao_database.AccountRole, error)
	getRoleByNameFn      func(ctx context.Context, name string) (dao_database.AccountRole, error)
	getRoleByAccountIdFn func(ctx context.Context, accountId uint64) (dao_database.AccountRole, error)
}

func (m *mockAccountRoleAccessor) GetRoleById(ctx context.Context, id uint8) (dao_database.AccountRole, error) {
	if m.getRoleByIdFn != nil {
		return m.getRoleByIdFn(ctx, id)
	}
	return dao_database.AccountRole{}, nil
}
func (m *mockAccountRoleAccessor) GetRoleByName(ctx context.Context, name string) (dao_database.AccountRole, error) {
	if m.getRoleByNameFn != nil {
		return m.getRoleByNameFn(ctx, name)
	}
	return dao_database.AccountRole{}, nil
}
func (m *mockAccountRoleAccessor) GetRoleByAccountId(ctx context.Context, accountId uint64) (dao_database.AccountRole, error) {
	if m.getRoleByAccountIdFn != nil {
		return m.getRoleByAccountIdFn(ctx, accountId)
	}
	return dao_database.AccountRole{}, nil
}
func (m *mockAccountRoleAccessor) WithExecutor(_ dao_database.Executor) dao_database.AccountRoleAccessor {
	return m
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestProfileLogic(
	accountLogic logic_account.Account,
	webhookAccessor dao_database.ChatbotWebhookAccessor,
	accountRoleAccessor dao_database.AccountRoleAccessor,
) logic_http.ProfileLogic {
	logger := zap.NewNop()
	signalCh := webhook_handler.NewCreatedWebhookChan()

	if accountRoleAccessor == nil {
		accountRoleAccessor = &mockAccountRoleAccessor{}
	}

	return logic_http.NewProfileLogic(webhookAccessor, accountRoleAccessor, signalCh, accountLogic, logger)
}

// newGinContext creates a gin.Context backed by httptest for the given method, path, and optional body.
func newGinContext(method, path string, body any) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		b, _ := json.Marshal(body)
		req = httptest.NewRequest(method, path, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	c.Request = req
	return c, w
}

// setAccountId sets the "accountId" key in the gin context, mimicking the auth middleware.
func setAccountId(c *gin.Context, id uint64) {
	c.Set("accountId", id)
}

// ---------------------------------------------------------------------------
// Tests: GetProfileMe
// ---------------------------------------------------------------------------

func TestGetProfileMe_Success(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, params logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			require.Equal(t, uint64(42), params.AccountId)
			return logic_account.GetAccountOutput{
				AccountId: 42,
				AccountInfo: logic_account.AccountInfo{
					Username:    "johndoe",
					Fullname:    "John Doe",
					Email:       "john@example.com",
					PhoneNumber: "1234567890",
					Role:        logic_account.Member,
				},
			}, nil
		},
	}
	roleAccessor := &mockAccountRoleAccessor{
		getRoleByIdFn: func(_ context.Context, id uint8) (dao_database.AccountRole, error) {
			require.Equal(t, uint8(logic_account.Member), id)
			return dao_database.AccountRole{Id: id, Name: "member"}, nil
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, roleAccessor)

	c, w := newGinContext(http.MethodGet, "/profile/me", nil)
	setAccountId(c, 42)
	pl.GetProfileMe(c)

	require.Equal(t, http.StatusOK, w.Code)

	var resp oapi.ProfileMeResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, "johndoe", resp.Account.Username)
	require.Equal(t, "John Doe", resp.Account.Fullname)
	require.Equal(t, "john@example.com", resp.Account.Email)
	require.Equal(t, oapi.Role("member"), resp.Account.Role)
}

func TestGetProfileMe_NoAccountId(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	c, w := newGinContext(http.MethodGet, "/profile/me", nil)
	// Do NOT set accountId
	pl.GetProfileMe(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetProfileMe_AccountLogicError(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{}, errors.New("db down")
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, nil)

	c, w := newGinContext(http.MethodGet, "/profile/me", nil)
	setAccountId(c, 1)
	pl.GetProfileMe(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetProfileMe_RoleAccessorError(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{
				AccountId:   1,
				AccountInfo: logic_account.AccountInfo{Role: logic_account.Member},
			}, nil
		},
	}
	roleAccessor := &mockAccountRoleAccessor{
		getRoleByIdFn: func(_ context.Context, _ uint8) (dao_database.AccountRole, error) {
			return dao_database.AccountRole{}, errors.New("role lookup failed")
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, roleAccessor)

	c, w := newGinContext(http.MethodGet, "/profile/me", nil)
	setAccountId(c, 1)
	pl.GetProfileMe(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GetProfileWebhooks
// ---------------------------------------------------------------------------

func TestGetProfileWebhooks_Success_Defaults(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhooksByAccountFn: func(_ context.Context, accountId uint64, limit, offset int) ([]dao_database.ChatbotWebhook, error) {
			require.Equal(t, uint64(10), accountId)
			require.Equal(t, 20, limit) // default
			require.Equal(t, 0, offset) // default
			return []dao_database.ChatbotWebhook{
				{Id: 1, OfAccountId: 10, Name: "hook1", Url: "https://a.com"},
				{Id: 2, OfAccountId: 10, Name: "hook2", Url: "https://b.com"},
			}, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks", nil)
	setAccountId(c, 10)
	pl.GetProfileWebhooks(c, oapi.GetProfileWebhooksParams{})

	require.Equal(t, http.StatusOK, w.Code)

	var resp []oapi.Webhook
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp, 2)
	require.Equal(t, "hook1", resp[0].Name)
	require.Equal(t, "hook2", resp[1].Name)
}

func TestGetProfileWebhooks_CustomLimitOffset(t *testing.T) {
	limit := 5
	offset := 10
	webhookAcc := &mockWebhookAccessor{
		getWebhooksByAccountFn: func(_ context.Context, _ uint64, l, o int) ([]dao_database.ChatbotWebhook, error) {
			require.Equal(t, 5, l)
			require.Equal(t, 10, o)
			return nil, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks?limit=5&offset=10", nil)
	setAccountId(c, 1)
	pl.GetProfileWebhooks(c, oapi.GetProfileWebhooksParams{
		Limit:  &limit,
		Offset: &offset,
	})

	require.Equal(t, http.StatusOK, w.Code)
}

func TestGetProfileWebhooks_NoAccountId(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks", nil)
	pl.GetProfileWebhooks(c, oapi.GetProfileWebhooksParams{})

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetProfileWebhooks_AccessorError(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhooksByAccountFn: func(_ context.Context, _ uint64, _, _ int) ([]dao_database.ChatbotWebhook, error) {
			return nil, errors.New("query failed")
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks", nil)
	setAccountId(c, 1)
	pl.GetProfileWebhooks(c, oapi.GetProfileWebhooksParams{})

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetProfileWebhooks_Empty(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhooksByAccountFn: func(_ context.Context, _ uint64, _, _ int) ([]dao_database.ChatbotWebhook, error) {
			return []dao_database.ChatbotWebhook{}, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks", nil)
	setAccountId(c, 1)
	pl.GetProfileWebhooks(c, oapi.GetProfileWebhooksParams{})

	require.Equal(t, http.StatusOK, w.Code)

	var resp []oapi.Webhook
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Len(t, resp, 0)
}

// ---------------------------------------------------------------------------
// Tests: CreateProfileWebhook
// ---------------------------------------------------------------------------

func TestCreateProfileWebhook_Success(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		createWebhookFn: func(_ context.Context, wh dao_database.ChatbotWebhook) (uint64, error) {
			require.Equal(t, uint64(7), wh.OfAccountId)
			require.Equal(t, "my-hook", wh.Name)
			require.Equal(t, "https://example.com/hook", wh.Url)
			return 99, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	body := oapi.Webhook{Name: "my-hook", Url: "https://example.com/hook"}
	c, w := newGinContext(http.MethodPost, "/profile/webhooks", body)
	setAccountId(c, 7)
	pl.CreateProfileWebhook(c)

	require.Equal(t, http.StatusCreated, w.Code)

	var resp oapi.Webhook
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp.Id)
	require.Equal(t, int64(99), *resp.Id)
	require.Equal(t, "my-hook", resp.Name)
	require.Equal(t, "https://example.com/hook", resp.Url)
}

func TestCreateProfileWebhook_NoAccountId(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	body := oapi.Webhook{Name: "h", Url: "http://x.com"}
	c, w := newGinContext(http.MethodPost, "/profile/webhooks", body)
	pl.CreateProfileWebhook(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateProfileWebhook_InvalidBody(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	// Send invalid JSON (string instead of object)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/profile/webhooks",
		bytes.NewReader([]byte(`not json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	setAccountId(c, 1)
	pl.CreateProfileWebhook(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateProfileWebhook_AccessorError(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		createWebhookFn: func(_ context.Context, _ dao_database.ChatbotWebhook) (uint64, error) {
			return 0, errors.New("duplicate entry")
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	body := oapi.Webhook{Name: "dup", Url: "http://x.com"}
	c, w := newGinContext(http.MethodPost, "/profile/webhooks", body)
	setAccountId(c, 1)
	pl.CreateProfileWebhook(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: GetProfileWebhook
// ---------------------------------------------------------------------------

func TestGetProfileWebhook_Success(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, id uint64) (dao_database.ChatbotWebhook, error) {
			require.Equal(t, uint64(55), id)
			return dao_database.ChatbotWebhook{
				Id: 55, OfAccountId: 3, Name: "wh", Url: "https://wh.io",
			}, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks/55", nil)
	setAccountId(c, 3)
	pl.GetProfileWebhook(c, 55)

	require.Equal(t, http.StatusOK, w.Code)

	var resp oapi.Webhook
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp.Id)
	require.Equal(t, int64(55), *resp.Id)
	require.Equal(t, "wh", resp.Name)
}

func TestGetProfileWebhook_NoAccountId(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks/1", nil)
	pl.GetProfileWebhook(c, 1)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetProfileWebhook_NotFound(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{}, errors.New("sql: no rows")
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks/999", nil)
	setAccountId(c, 1)
	pl.GetProfileWebhook(c, 999)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetProfileWebhook_Forbidden(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{
				Id: 10, OfAccountId: 99, Name: "other", Url: "http://x.com",
			}, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodGet, "/profile/webhooks/10", nil)
	setAccountId(c, 1) // different from OfAccountId=99
	pl.GetProfileWebhook(c, 10)

	require.Equal(t, http.StatusForbidden, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: UpdateProfileWebhook
// ---------------------------------------------------------------------------

func TestUpdateProfileWebhook_Success(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, id uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{
				Id: id, OfAccountId: 5, Name: "old", Url: "http://old.com",
			}, nil
		},
		updateWebhookFn: func(_ context.Context, wh dao_database.ChatbotWebhook) error {
			require.Equal(t, uint64(20), wh.Id)
			require.Equal(t, "new-name", wh.Name)
			require.Equal(t, "https://new.com", wh.Url)
			return nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	body := oapi.Webhook{Name: "new-name", Url: "https://new.com"}
	c, w := newGinContext(http.MethodPut, "/profile/webhooks/20", body)
	setAccountId(c, 5)
	pl.UpdateProfileWebhook(c, 20)

	require.Equal(t, http.StatusOK, w.Code)

	var resp oapi.Webhook
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.NotNil(t, resp.Id)
	require.Equal(t, int64(20), *resp.Id)
	require.Equal(t, "new-name", resp.Name)
	require.Equal(t, "https://new.com", resp.Url)
}

func TestUpdateProfileWebhook_NoAccountId(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	body := oapi.Webhook{Name: "x", Url: "http://x.com"}
	c, w := newGinContext(http.MethodPut, "/profile/webhooks/1", body)
	pl.UpdateProfileWebhook(c, 1)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateProfileWebhook_NotFound(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{}, errors.New("not found")
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	body := oapi.Webhook{Name: "x", Url: "http://x.com"}
	c, w := newGinContext(http.MethodPut, "/profile/webhooks/999", body)
	setAccountId(c, 1)
	pl.UpdateProfileWebhook(c, 999)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateProfileWebhook_Forbidden(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{
				Id: 10, OfAccountId: 99, // owned by account 99
			}, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	body := oapi.Webhook{Name: "x", Url: "http://x.com"}
	c, w := newGinContext(http.MethodPut, "/profile/webhooks/10", body)
	setAccountId(c, 1) // not 99
	pl.UpdateProfileWebhook(c, 10)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateProfileWebhook_InvalidBody(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{Id: 1, OfAccountId: 1}, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/profile/webhooks/1",
		bytes.NewReader([]byte(`{bad json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	setAccountId(c, 1)
	pl.UpdateProfileWebhook(c, 1)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProfileWebhook_AccessorError(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{Id: 1, OfAccountId: 1}, nil
		},
		updateWebhookFn: func(_ context.Context, _ dao_database.ChatbotWebhook) error {
			return errors.New("db error")
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	body := oapi.Webhook{Name: "x", Url: "http://x.com"}
	c, w := newGinContext(http.MethodPut, "/profile/webhooks/1", body)
	setAccountId(c, 1)
	pl.UpdateProfileWebhook(c, 1)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: DeleteProfileWebhook
// ---------------------------------------------------------------------------

func TestDeleteProfileWebhook_Success(t *testing.T) {
	deleteCalled := false
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, id uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{Id: id, OfAccountId: 8}, nil
		},
		deleteWebhookFn: func(_ context.Context, id uint64) error {
			require.Equal(t, uint64(30), id)
			deleteCalled = true
			return nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, _ := newGinContext(http.MethodDelete, "/profile/webhooks/30", nil)
	setAccountId(c, 8)
	pl.DeleteProfileWebhook(c, 30)

	// c.Status() sets gin's internal writer status but does not flush to
	// httptest.ResponseRecorder when no body is written, so check the writer.
	require.Equal(t, http.StatusNoContent, c.Writer.Status())
	require.True(t, deleteCalled)
}

func TestDeleteProfileWebhook_NoAccountId(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	c, w := newGinContext(http.MethodDelete, "/profile/webhooks/1", nil)
	pl.DeleteProfileWebhook(c, 1)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestDeleteProfileWebhook_NotFound(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{}, errors.New("not found")
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodDelete, "/profile/webhooks/999", nil)
	setAccountId(c, 1)
	pl.DeleteProfileWebhook(c, 999)

	require.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteProfileWebhook_Forbidden(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{
				Id: 10, OfAccountId: 99,
			}, nil
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodDelete, "/profile/webhooks/10", nil)
	setAccountId(c, 1) // not 99
	pl.DeleteProfileWebhook(c, 10)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteProfileWebhook_AccessorError(t *testing.T) {
	webhookAcc := &mockWebhookAccessor{
		getWebhookFn: func(_ context.Context, _ uint64) (dao_database.ChatbotWebhook, error) {
			return dao_database.ChatbotWebhook{Id: 1, OfAccountId: 1}, nil
		},
		deleteWebhookFn: func(_ context.Context, _ uint64) error {
			return errors.New("db error")
		},
	}
	pl := newTestProfileLogic(&mockAccountLogic{}, webhookAcc, nil)

	c, w := newGinContext(http.MethodDelete, "/profile/webhooks/1", nil)
	setAccountId(c, 1)
	pl.DeleteProfileWebhook(c, 1)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// Tests: UpdateProfileMe
// ---------------------------------------------------------------------------

func strPtr(s string) *string { return &s }

func TestUpdateProfileMe_Success_AllFields(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, params logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			require.Equal(t, uint64(42), params.AccountId)
			return logic_account.GetAccountOutput{
				AccountId: 42,
				AccountInfo: logic_account.AccountInfo{
					Username:    "johndoe",
					Fullname:    "John Doe",
					Email:       "john@example.com",
					PhoneNumber: "1234567890",
					Role:        logic_account.Member,
				},
			}, nil
		},
		updateAccountInfoFn: func(_ context.Context, params logic_account.UpdateAccountInfoParams) (logic_account.UpdateAccountInfoOutput, error) {
			require.Equal(t, uint64(42), params.AccountId)
			require.Equal(t, "Jane Doe", params.UpdatedAccountInfo.Fullname)
			require.Equal(t, "jane@example.com", params.UpdatedAccountInfo.Email)
			require.Equal(t, "0987654321", params.UpdatedAccountInfo.PhoneNumber)
			// Username and Role should be preserved from original
			require.Equal(t, "johndoe", params.UpdatedAccountInfo.Username)
			require.Equal(t, logic_account.Member, params.UpdatedAccountInfo.Role)
			return logic_account.UpdateAccountInfoOutput{AccountId: 42}, nil
		},
	}
	roleAccessor := &mockAccountRoleAccessor{
		getRoleByIdFn: func(_ context.Context, id uint8) (dao_database.AccountRole, error) {
			return dao_database.AccountRole{Id: id, Name: "member"}, nil
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, roleAccessor)

	body := oapi.UpdateProfileMeRequest{
		Fullname:    strPtr("Jane Doe"),
		Email:       strPtr("jane@example.com"),
		PhoneNumber: strPtr("0987654321"),
	}
	c, w := newGinContext(http.MethodPut, "/profile/me", body)
	setAccountId(c, 42)
	pl.UpdateProfileMe(c)

	require.Equal(t, http.StatusOK, w.Code)

	var resp oapi.ProfileMeResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, "johndoe", resp.Account.Username)
	require.Equal(t, "Jane Doe", resp.Account.Fullname)
	require.Equal(t, "jane@example.com", resp.Account.Email)
	require.Equal(t, oapi.Role("member"), resp.Account.Role)
}

func TestUpdateProfileMe_Success_PartialUpdate(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{
				AccountId: 1,
				AccountInfo: logic_account.AccountInfo{
					Username:    "alice",
					Fullname:    "Alice Original",
					Email:       "alice@old.com",
					PhoneNumber: "1111111111",
					Role:        logic_account.Member,
				},
			}, nil
		},
		updateAccountInfoFn: func(_ context.Context, params logic_account.UpdateAccountInfoParams) (logic_account.UpdateAccountInfoOutput, error) {
			// Only fullname should change
			require.Equal(t, "Alice Updated", params.UpdatedAccountInfo.Fullname)
			require.Equal(t, "alice@old.com", params.UpdatedAccountInfo.Email)
			require.Equal(t, "1111111111", params.UpdatedAccountInfo.PhoneNumber)
			return logic_account.UpdateAccountInfoOutput{AccountId: 1}, nil
		},
	}
	roleAccessor := &mockAccountRoleAccessor{
		getRoleByIdFn: func(_ context.Context, id uint8) (dao_database.AccountRole, error) {
			return dao_database.AccountRole{Id: id, Name: "member"}, nil
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, roleAccessor)

	body := oapi.UpdateProfileMeRequest{
		Fullname: strPtr("Alice Updated"),
	}
	c, w := newGinContext(http.MethodPut, "/profile/me", body)
	setAccountId(c, 1)
	pl.UpdateProfileMe(c)

	require.Equal(t, http.StatusOK, w.Code)

	var resp oapi.ProfileMeResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	require.Equal(t, "Alice Updated", resp.Account.Fullname)
	require.Equal(t, "alice@old.com", resp.Account.Email)
}

func TestUpdateProfileMe_NoAccountId(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	body := oapi.UpdateProfileMeRequest{Fullname: strPtr("x")}
	c, w := newGinContext(http.MethodPut, "/profile/me", body)
	pl.UpdateProfileMe(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateProfileMe_InvalidBody(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/profile/me",
		bytes.NewReader([]byte(`not json`)))
	c.Request.Header.Set("Content-Type", "application/json")
	setAccountId(c, 1)
	pl.UpdateProfileMe(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProfileMe_GetAccountError(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{}, errors.New("db down")
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, nil)

	body := oapi.UpdateProfileMeRequest{Fullname: strPtr("x")}
	c, w := newGinContext(http.MethodPut, "/profile/me", body)
	setAccountId(c, 1)
	pl.UpdateProfileMe(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateProfileMe_UpdateAccountInfoError(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{
				AccountId:   1,
				AccountInfo: logic_account.AccountInfo{Role: logic_account.Member},
			}, nil
		},
		updateAccountInfoFn: func(_ context.Context, _ logic_account.UpdateAccountInfoParams) (logic_account.UpdateAccountInfoOutput, error) {
			return logic_account.UpdateAccountInfoOutput{}, errors.New("update failed")
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, nil)

	body := oapi.UpdateProfileMeRequest{Fullname: strPtr("x")}
	c, w := newGinContext(http.MethodPut, "/profile/me", body)
	setAccountId(c, 1)
	pl.UpdateProfileMe(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateProfileMe_RoleAccessorError(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{
				AccountId:   1,
				AccountInfo: logic_account.AccountInfo{Role: logic_account.Member},
			}, nil
		},
		updateAccountInfoFn: func(_ context.Context, _ logic_account.UpdateAccountInfoParams) (logic_account.UpdateAccountInfoOutput, error) {
			return logic_account.UpdateAccountInfoOutput{AccountId: 1}, nil
		},
	}
	roleAccessor := &mockAccountRoleAccessor{
		getRoleByIdFn: func(_ context.Context, _ uint8) (dao_database.AccountRole, error) {
			return dao_database.AccountRole{}, errors.New("role lookup failed")
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, roleAccessor)

	body := oapi.UpdateProfileMeRequest{Fullname: strPtr("x")}
	c, w := newGinContext(http.MethodPut, "/profile/me", body)
	setAccountId(c, 1)
	pl.UpdateProfileMe(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

// ===========================================================================
// UpdateProfilePassword
// ===========================================================================

func TestUpdateProfilePassword_Success(t *testing.T) {
	oldPw := "oldSecret123"
	newPw := "newSecret456"
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{
				AccountId:   1,
				AccountInfo: logic_account.AccountInfo{Username: "alice"},
			}, nil
		},
		checkAccountValidFn: func(_ context.Context, params logic_account.CheckAccountValidParams) (logic_account.CheckAccountValidOutput, error) {
			require.Equal(t, "alice", params.Username)
			require.Equal(t, oldPw, params.Password)
			return logic_account.CheckAccountValidOutput{AccountId: 1}, nil
		},
		updateAccountPasswordFn: func(_ context.Context, params logic_account.UpdateAccountPasswordParams) (logic_account.UpdateAccountPasswordOutput, error) {
			require.Equal(t, uint64(1), params.AccountId)
			require.Equal(t, newPw, params.Password)
			return logic_account.UpdateAccountPasswordOutput{}, nil
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, nil)

	body := oapi.UpdatePasswordRequest{OldPassword: &oldPw, NewPassword: &newPw}
	c, _ := newGinContext(http.MethodPut, "/profile/me/password", body)
	setAccountId(c, 1)
	pl.UpdateProfilePassword(c)

	require.Equal(t, http.StatusNoContent, c.Writer.Status())
}

func TestUpdateProfilePassword_NoAccountId(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	body := oapi.UpdatePasswordRequest{}
	c, w := newGinContext(http.MethodPut, "/profile/me/password", body)
	// Do NOT set accountId
	pl.UpdateProfilePassword(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestUpdateProfilePassword_InvalidBody(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	c, w := newGinContext(http.MethodPut, "/profile/me/password", "not-json")
	setAccountId(c, 1)
	pl.UpdateProfilePassword(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProfilePassword_MissingFields(t *testing.T) {
	pl := newTestProfileLogic(&mockAccountLogic{}, &mockWebhookAccessor{}, nil)

	body := oapi.UpdatePasswordRequest{} // both nil
	c, w := newGinContext(http.MethodPut, "/profile/me/password", body)
	setAccountId(c, 1)
	pl.UpdateProfilePassword(c)

	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProfilePassword_GetAccountError(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{}, errors.New("db down")
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, nil)

	oldPw := "old"
	newPw := "new"
	body := oapi.UpdatePasswordRequest{OldPassword: &oldPw, NewPassword: &newPw}
	c, w := newGinContext(http.MethodPut, "/profile/me/password", body)
	setAccountId(c, 1)
	pl.UpdateProfilePassword(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateProfilePassword_WrongOldPassword(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{
				AccountId:   1,
				AccountInfo: logic_account.AccountInfo{Username: "alice"},
			}, nil
		},
		checkAccountValidFn: func(_ context.Context, _ logic_account.CheckAccountValidParams) (logic_account.CheckAccountValidOutput, error) {
			return logic_account.CheckAccountValidOutput{AccountId: 0}, nil // wrong password
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, nil)

	oldPw := "wrong"
	newPw := "new"
	body := oapi.UpdatePasswordRequest{OldPassword: &oldPw, NewPassword: &newPw}
	c, w := newGinContext(http.MethodPut, "/profile/me/password", body)
	setAccountId(c, 1)
	pl.UpdateProfilePassword(c)

	require.Equal(t, http.StatusForbidden, w.Code)
}

func TestUpdateProfilePassword_CheckAccountValidError(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{
				AccountId:   1,
				AccountInfo: logic_account.AccountInfo{Username: "alice"},
			}, nil
		},
		checkAccountValidFn: func(_ context.Context, _ logic_account.CheckAccountValidParams) (logic_account.CheckAccountValidOutput, error) {
			return logic_account.CheckAccountValidOutput{}, errors.New("validation service down")
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, nil)

	oldPw := "old"
	newPw := "new"
	body := oapi.UpdatePasswordRequest{OldPassword: &oldPw, NewPassword: &newPw}
	c, w := newGinContext(http.MethodPut, "/profile/me/password", body)
	setAccountId(c, 1)
	pl.UpdateProfilePassword(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateProfilePassword_UpdatePasswordError(t *testing.T) {
	acctLogic := &mockAccountLogic{
		getAccountFn: func(_ context.Context, _ logic_account.GetAccountParams) (logic_account.GetAccountOutput, error) {
			return logic_account.GetAccountOutput{
				AccountId:   1,
				AccountInfo: logic_account.AccountInfo{Username: "alice"},
			}, nil
		},
		checkAccountValidFn: func(_ context.Context, _ logic_account.CheckAccountValidParams) (logic_account.CheckAccountValidOutput, error) {
			return logic_account.CheckAccountValidOutput{AccountId: 1}, nil
		},
		updateAccountPasswordFn: func(_ context.Context, _ logic_account.UpdateAccountPasswordParams) (logic_account.UpdateAccountPasswordOutput, error) {
			return logic_account.UpdateAccountPasswordOutput{}, errors.New("hash failure")
		},
	}
	pl := newTestProfileLogic(acctLogic, &mockWebhookAccessor{}, nil)

	oldPw := "old"
	newPw := "new"
	body := oapi.UpdatePasswordRequest{OldPassword: &oldPw, NewPassword: &newPw}
	c, w := newGinContext(http.MethodPut, "/profile/me/password", body)
	setAccountId(c, 1)
	pl.UpdateProfilePassword(c)

	require.Equal(t, http.StatusInternalServerError, w.Code)
}
