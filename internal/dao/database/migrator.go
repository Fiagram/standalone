package dao_database

import (
	"context"
	"database/sql"
	"embed"

	"github.com/Fiagram/standalone/internal/logger"
	"github.com/Fiagram/standalone/internal/utils"
	migrate "github.com/rubenv/sql-migrate"
	"go.uber.org/zap"
)

var (
	//go:embed migrations/mysql/*.sql
	migrationDirectoryMysql embed.FS
)

type Migration interface {
	Up(ctx context.Context) error
	Down(ctx context.Context) error
}

type migrator struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewMigrator(
	db *sql.DB,
	logger *zap.Logger,
) Migration {
	return &migrator{
		db:     db,
		logger: logger,
	}
}

func (m migrator) migrate(ctx context.Context, direction migrate.MigrationDirection) error {
	logger := logger.LoggerWithContext(ctx, m.logger).
		With(zap.String("direction", utils.If(direction == migrate.Up, "up", "down")))

	applied_migrations, err := migrate.ExecContext(ctx, m.db, "mysql",
		migrate.EmbedFileSystemMigrationSource{
			FileSystem: migrationDirectoryMysql,
			Root:       "migrations/mysql",
		},
		direction)
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to execute migration")
		return err
	}

	logger.With(zap.Int("applied_migrations", applied_migrations)).
		Info("successfully executed database migrations")
	return nil
}

func (m migrator) Down(ctx context.Context) error {
	return m.migrate(ctx, migrate.Down)
}

func (m migrator) Up(ctx context.Context) error {
	return m.migrate(ctx, migrate.Up)
}
