package db

import (
	"context"
	"fmt"
	"time"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect is the main entry for the pool connection.
func Connect(conf *config.Config) (*pgxpool.Pool, error) {
	var psqlInfo string

	var (
		host     = conf.DBHost
		port     = conf.DBPort
		user     = conf.DBUser
		password = conf.DBPassword
		dbName   = conf.DBName
	)

	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName)

	parseConfig, parseErr := pgxpool.ParseConfig(psqlInfo)
	if parseErr != nil {
		return nil, parseErr
	}

	parseConfig.MaxConns = 50
	parseConfig.MinConns = 10
	parseConfig.MaxConnIdleTime = time.Hour * 1
	parseConfig.MaxConnLifetime = time.Hour
	parseConfig.HealthCheckPeriod = 1 * time.Minute
	parseConfig.ConnConfig.ConnectTimeout = 60 * time.Second

	pool, err := pgxpool.NewWithConfig(context.Background(), parseConfig)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		return nil, err
	}

	log.Info.Printf("DBName: %q, message: Database connection is ready..\n", conf.DBName)

	s := pool.Stat()

	log.Info.Println("Pool health", "total", s.MaxConns(), "idle", s.IdleConns(), "acquired", s.AcquireCount())
	return pool, nil
}
