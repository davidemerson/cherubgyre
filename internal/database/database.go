package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/dev3mike/go-api-swagger-boilerplate/internal/environments"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

var DB *bun.DB

func Connect(logger *slog.Logger) error {
	dbConnection := environments.Env.DbConnection
	sqldb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dbConnection)))

	DB = bun.NewDB(sqldb, pgdialect.New())

	if err := DB.Ping(); err != nil {
		panic(fmt.Errorf("❌ Failed to connect to database: %w", err))
	}

	logger.Info("✅ Connected to the database successfully!")
	return nil
}
