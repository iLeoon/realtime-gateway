package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// the entry point for all the config variables
var (
	ErrLoadEnvFile = errors.New("can't find the env file")
	ErrRetriveKey  = errors.New("no key with that name exists")
)

func Load() (*Config, error) {

	err := godotenv.Load(".env")
	if err != nil {
		return nil, fmt.Errorf("%w:%v", ErrLoadEnvFile, err)
	}

	c := &Config{}
	c.TCPServer.Port, err = getEnv("TCP_SERVER_PORT")
	if err != nil {
		return nil, err
	}

	c.Websocket.Port, err = getEnv("WEBSOKCET_SERVR_PORT")
	if err != nil {
		return nil, err
	}
	return c, nil
}

func getEnv(key string) (string, error) {

	if val, exists := os.LookupEnv(key); exists {
		return val, nil
	}
	return "", fmt.Errorf("%w:%v", ErrRetriveKey, key)
}
