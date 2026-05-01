package configs_test

import (
	"log"
	"os"
	"testing"
	"time"

	"net/http"

	"github.com/Fiagram/standalone/internal/configs"
	"github.com/stretchr/testify/require"
	yaml_pkg "gopkg.in/yaml.v3"
)

var config configs.Config

func TestMain(m *testing.M) {
	cfg, err := configs.NewConfig("")
	if err != nil {
		log.Fatal("failed to init config default")
	}
	config = cfg
	os.Exit(m.Run())
}

func TestAuth(t *testing.T) {
	require.Equal(t, 24*time.Hour, config.Auth.Token.RefreshTokenTTL)
	require.Equal(t, 15*time.Minute, config.Auth.Token.AccessTokenTTL)
	require.Equal(t, "secret_token_in_here", config.Auth.Token.Secret)
	require.Equal(t, 720*time.Hour, config.Auth.Token.RefreshTokenLongTTL)
	require.Equal(t, 10, config.Auth.Hash.Cost)
}

func TestHttp(t *testing.T) {
	require.Equal(t, "0.0.0.0", config.Http.Address)
	require.Equal(t, "8080", config.Http.Port)

	cors := config.Http.CORS
	require.True(t, cors.IsEnable)
	require.Equal(t, []string{
		"http://localhost:3000",
		"https://fiagram.io.vn",
		"https://app.fiagram.io.vn",
	}, cors.AllowOrigins)
	require.Equal(t, []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}, cors.AllowMethods)
	require.Equal(t, []string{"Origin", "Content-Type", "Authorization"}, cors.AllowHeaders)
	require.Equal(t, []string{"Content-Length"}, cors.ExposeHeaders)
	require.True(t, cors.AllowCredentials)
	require.Equal(t, 12*time.Hour, cors.MaxAge)
}

func TestHttpCookie(t *testing.T) {
	cookie := config.Http.Cookie
	require.Equal(t, "localhost", cookie.Domain)
	require.Equal(t, "/", cookie.Path)
	require.Equal(t, "none", cookie.SameSiteMode)
	require.Equal(t, http.SameSiteNoneMode, cookie.SameSite())
	require.True(t, cookie.Secure)
	require.True(t, cookie.HttpOnly)
}

func TestLog(t *testing.T) {
	require.Equal(t, "debug", config.Log.Level)
}

func TestCacheClient(t *testing.T) {
	require.Equal(t, "redis", string(config.CacheClient.Type))
	require.Equal(t, "127.0.0.1", config.CacheClient.Address)
	require.Equal(t, "6379", config.CacheClient.Port)
	require.Equal(t, "", config.CacheClient.Username)
	require.Equal(t, "", config.CacheClient.Password)
}

func TestDatabaseClient(t *testing.T) {
	require.Equal(t, "mysql", string(config.DatabaseClient.Type))
	require.Equal(t, "127.0.0.1", config.DatabaseClient.Address)
	require.Equal(t, 3306, config.DatabaseClient.Port)
	require.Equal(t, "root", config.DatabaseClient.Username)
	require.Equal(t, "root", config.DatabaseClient.Password)
	require.Equal(t, "fiagram", config.DatabaseClient.Database)
}

func TestStrategyAlertQuota_DefaultConfig(t *testing.T) {
	// Verifies the embedded local.yaml values are loaded correctly.
	q := config.Strategy.AlertQuota
	require.Equal(t, 1, q.Free)
	require.Equal(t, 10, q.Pro)
	require.Equal(t, 0, q.Max) // "*" in local.yaml → 0 (unlimited)
}

func TestAlertQuota_UnmarshalYAML_Star(t *testing.T) {
	yaml := `free: 1
pro: 10
max: "*"`
	var q configs.AlertQuota
	require.NoError(t, unmarshalAlertQuota(yaml, &q))
	require.Equal(t, 1, q.Free)
	require.Equal(t, 10, q.Pro)
	require.Equal(t, 0, q.Max, "\"*\" should be treated as 0 (unlimited)")
}

func TestAlertQuota_UnmarshalYAML_IntegerMax(t *testing.T) {
	yaml := `free: 1
pro: 10
max: 5`
	var q configs.AlertQuota
	require.NoError(t, unmarshalAlertQuota(yaml, &q))
	require.Equal(t, 5, q.Max)
}

func TestAlertQuota_UnmarshalYAML_OmittedMax(t *testing.T) {
	yaml := `free: 1
pro: 10`
	var q configs.AlertQuota
	require.NoError(t, unmarshalAlertQuota(yaml, &q))
	require.Equal(t, 0, q.Max, "omitted max should default to 0 (unlimited)")
}

// unmarshalAlertQuota is a helper that decodes a YAML string into an AlertQuota.
func unmarshalAlertQuota(src string, q *configs.AlertQuota) error {
	return yaml_pkg.Unmarshal([]byte(src), q)
}
