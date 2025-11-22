package config

import (
	"github.com/joho/godotenv"
	"os"
)

//loads the config variables

func Load() *Config {

	err := godotenv.Load()
	if err != nil {
		panic("An Error while trying to get the env file")
	}

	c := &Config{}
	c.TCPServer.Port = getEnv("TCP_SERVER_PORT")

	return c
}

func getEnv(key string) string {

	if val, exists := os.LookupEnv(key); exists {
		return val
	}

	panic("An Error while trying to load the env")
}
