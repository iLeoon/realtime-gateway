package db

import (
	"context"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(conf *config.Config) (*pgxpool.Pool, error) {

	var (
		host     = conf.DBHost
		port     = conf.DBPort
		user     = conf.DBUser
		password = conf.DBPassword
		dbname   = conf.DBName
	)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	parseConfig, parseErr := pgxpool.ParseConfig(psqlInfo)
	if parseErr != nil {
		return nil, parseErr
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), parseConfig)

	if err != nil {
		return nil, err
	}

	logger.Info("Database pool connection is ready.")
	return pool, nil
}
