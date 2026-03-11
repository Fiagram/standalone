package dao_database_test

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/Fiagram/standalone/internal/configs"
	dao_database "github.com/Fiagram/standalone/internal/dao/database"
	"go.uber.org/zap"
)

var sqlDb *sql.DB
var logger *zap.Logger

func TestMain(m *testing.M) {
	// Use the default config to test database connection
	config, err := configs.NewConfig("")
	if err != nil {
		log.Fatal("failed to init config default")
	}

	logger = zap.NewNop()

	db, dbCleanup, err := dao_database.InitAndMigrateUpDatabase(config.DatabaseClient, logger)
	if err != nil {
		log.Fatal("failed to init and migrate up database")
	}
	defer dbCleanup()
	sqlDb = db

	os.Exit(m.Run())
}
