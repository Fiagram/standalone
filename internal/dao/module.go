package dao

import (
	dao_cache "github.com/Fiagram/standalone/internal/dao/cache"
	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	dao_mq_consumer "github.com/Fiagram/standalone/internal/dao/message_queue/consumer"
	dao_mq_producer "github.com/Fiagram/standalone/internal/dao/message_queue/producer"
	dao_strategy "github.com/Fiagram/standalone/internal/dao/strategy"
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
		dao_database.NewChatbotWebhookAccessor,

		dao_cache.NewDaoCache,
		dao_cache.NewDaoCacheUsernamesTaken,
		dao_cache.NewDaoCacheRefreshToken,

		dao_mq_consumer.NewDaoMessageQueueConsumer,
		dao_mq_producer.NewDaoMessageQueueProducer,

		dao_strategy.NewClient,
	),
)
