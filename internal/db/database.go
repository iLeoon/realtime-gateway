package db

import (
	"context"
	"fmt"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Main entry for our pool connection.
func Connect(conf *config.Config) (*pgxpool.Pool, error) {
	var psqlInfo string

	var (
		host     = conf.DBHost
		port     = conf.DBPort
		user     = conf.DBUser
		password = conf.DBPassword
		dbname   = conf.DBName
	)

	if conf.DatabaseURL != "" {
		psqlInfo = conf.DatabaseURL
	} else {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
			"password=%s dbname=%s sslmode=disable",
			host, port, user, password, dbname)

	}
	parseConfig, parseErr := pgxpool.ParseConfig(psqlInfo)
	if parseErr != nil {
		return nil, parseErr
	}

	parseConfig.MaxConns = 20
	parseConfig.MinConns = 5
	parseConfig.MaxConnIdleTime = time.Hour * 1
	parseConfig.MaxConnLifetime = time.Hour
	parseConfig.HealthCheckPeriod = 1 * time.Minute
	parseConfig.ConnConfig.ConnectTimeout = 60 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), parseConfig)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return nil, err
	}

	logger.Info("Database pool connection is ready.")

	s := pool.Stat()

	logger.Info("Pool health", "total", s.MaxConns(), "idle", s.IdleConns(), "acquired", s.AcquireCount())
	return pool, nil
}
