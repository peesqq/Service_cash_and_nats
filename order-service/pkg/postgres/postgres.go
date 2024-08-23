package postgres

import (
	"context"
	"fmt"
	"net"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

const (
	dbUser = "postgres"
	dbPass = "postgres"
	dbHost = "localhost"
	dbPort = "5432"
	dbName = "postgres"
)

func Connect(ctx context.Context /*, config.Postgres*/) (*sqlx.DB, error) {
	hostPort := net.JoinHostPort(dbHost, dbPort)
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s",
		dbUser,
		dbPass,
		hostPort,
		dbName,
	)
	dbClient, err := sqlx.ConnectContext(ctx, "pgx", dsn)
	if err != nil {
		return nil, err
	}
	return dbClient, nil
}
