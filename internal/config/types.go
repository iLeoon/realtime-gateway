package config

// a file that defines the configuration variables

type Config struct {
	TCPServer
	Websocket
}

type TCPServer struct {
	Port string
}

type Websocket struct {
	Port string
}
