package config

import (
	"errors"
	"os"

	"github.com/iLeoon/chatserver/pkg/logger"
	"github.com/joho/godotenv"
)

// load the config variables
var (
	ErrLoadEnvFile = errors.New("can't find the env file")
	ErrRetriveKey  = errors.New("no key with that name exists")
)

func Load() *Config {

	err := godotenv.Load()
	if err != nil {
		panic("An Error while trying to get the env file")
	}

	logger.Info("Loaded the env file")

	c := &Config{}
	c.TCPServer.Port = getEnv("TCP_SERVER_PORT")
	c.Websocket.Port = getEnv("WEBSOKCET_SERVER_PORT")

	return c
}

func getEnv(key string) string {

	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	panic("An Error while trying to load the env")
}
