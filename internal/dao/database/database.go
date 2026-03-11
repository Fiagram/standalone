package dao_database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Fiagram/standalone/internal/configs"
	"go.uber.org/fx"
	"go.uber.org/zap"

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
)

var ErrLackOfInfor = errors.New("lack of information")

type Executor interface {
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)

	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row

	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

// These are compile-time checks that verify
// our interface matches the API of *sql.DB and *sql.Tx.
var _ Executor = (*sql.DB)(nil)
var _ Executor = (*sql.Tx)(nil)

func InitAndMigrateUpDatabase(databaseConfig configs.DatabaseClient, logger *zap.Logger) (*sql.DB, func(), error) {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		databaseConfig.Username,
		databaseConfig.Password,
		databaseConfig.Address,
		databaseConfig.Port,
		databaseConfig.Database,
	)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed when connecting the database")
		return nil, nil, err
	}

	cleanupDb := func() {
		db.Close()
	}

	migrator := NewMigrator(db, logger)
	err = migrator.Up(context.Background())
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to execute database migration up")
		cleanupDb()
		return nil, nil, err
	}

	return db, cleanupDb, nil
}

func NewDaoDatabase(
	lc fx.Lifecycle,
	databaseConfig configs.DatabaseClient,
	logger *zap.Logger,
) (*sql.DB, error) {
	db, cleanup, err := InitAndMigrateUpDatabase(databaseConfig, logger)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			cleanup()
			return nil
		},
	})

	return db, nil
}

func NewDaoDatabaseExecutor(db *sql.DB) Executor {
	return db
}
