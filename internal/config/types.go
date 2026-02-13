package config

// a file that defines the configuration variables

type Config struct {
	TCP
	HttpServer
	GoogleOAuth
	PostgreSQL
	JWT
	CORS
}

type TCP struct {
	TcpPort string `env:"TCP_SERVER_PORT"`
}

type HttpServer struct {
	HttpPort string `env:"HTTP_PORT"`
}

type GoogleOAuth struct {
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
	RedirectURL        string `env:"REDIRECT_URL"`
}

type PostgreSQL struct {
	DBHost     string `env:"POSTGRES_HOST"`
	DBPort     int    `env:"POSTGRES_PORT"`
	DBUser     string `env:"POSTGRES_USER"`
	DBPassword string `env:"POSTGRES_PASSWORD"`
	DBName     string `env:"POSTGRES_DATABASE"`

	DatabaseURL string `env:"DATABASE_URL"`
	TestDBURL   string `env:"TEST_DB"`
}

type JWT struct {
	JwtSecretKey string `env:"JWT_SECRET_KEY"`
	JwtIssuer    string `env:"JWT_ISSUER"`
}

type CORS struct {
	FrontEndOrigin string `env:"FRONTEND_ORIGIN"`
}
