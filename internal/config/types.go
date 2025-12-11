package config

// a file that defines the configuration variables

type Config struct {
	TCP
	Websocket
	HttpServer
}

type TCP struct {
	TcpPort string
}

type Websocket struct {
	WsPort string
}

type HttpServer struct {
	HttpPort string
}
