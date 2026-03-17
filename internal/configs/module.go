package configs

import "go.uber.org/fx"

var Module = fx.Module(
	"config",
	fx.Provide(
		NewConfig,
		GetConfigHttp,
		GetConfigHttpCORS,
		GetConfigHttpCookie,
		GetConfigLog,
		GetConfigCacheClient,
		GetConfigDatabaseClient,
		GetConfigAuth,
		GetConfigAuthHash,
		GetConfigAuthToken,
	),
)
