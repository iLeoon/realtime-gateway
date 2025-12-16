package src

import (
	"context"
	"fmt"

	"github.com/iLeoon/realtime-gateway/internal/config"
	"github.com/iLeoon/realtime-gateway/pkg/logger"
	"github.com/jackc/pgx/v5"
)

func Connect(conf *config.Config) (*pgx.Conn, error) {

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

	conn, err := pgx.Connect(context.Background(), psqlInfo)
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())

	logger.Info("Database connected successfully")
	return conn, nil
}
