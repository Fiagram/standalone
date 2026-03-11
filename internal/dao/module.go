package dao

import (
	dao_cache "github.com/Fiagram/standalone/internal/dao/cache"
	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"go.uber.org/fx"
)

var Module = fx.Module(
	"dao",
	fx.Provide(
		dao_database.NewDaoDatabase,
		dao_database.NewDaoDatabaseExecutor,
		dao_database.NewAccountAccessor,
		dao_database.NewAccountPasswordAccessor,
		dao_database.NewAccountRoleAccessor,

		dao_cache.NewDaoCache,
		dao_cache.NewDaoCacheUsernamesTaken,
		dao_cache.NewDaoCacheRefreshToken,
	),
)
