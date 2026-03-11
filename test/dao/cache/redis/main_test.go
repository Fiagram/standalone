package dao_cache_test

import (
	"log"
	"os"
	"testing"

	"github.com/Fiagram/standalone/internal/configs"
	dao_cache "github.com/Fiagram/standalone/internal/dao/cache"

	"go.uber.org/zap"
)

var client dao_cache.Client

func TestMain(m *testing.M) {
	// Use the default config to test database connection
	config, err := configs.NewConfig("")
	if err != nil {
		log.Fatal("failed to init config default")
	}

	config.CacheClient.Type = configs.CacheTypeRedis
	logger := zap.NewNop()

	cl, err := dao_cache.NewDaoCache(config.CacheClient, logger)
	if err != nil {
		log.Fatal("failed to init redis cache")
	}
	client = cl

	os.Exit(m.Run())
}
