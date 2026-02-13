package config

// a file that defines the configuration variables

type Config struct {
	TCP
	HttpServer
	GoogleOAuth
	PostgreSQL
	JWT
	CORS
	Scalingo
}

type TCP struct {
	TcpPort string `env:"TCP_SERVER_PORT,required"`
}

type HttpServer struct {
	HttpPort string `env:"HTTP_PORT,required"`
}

type GoogleOAuth struct {
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	RedirectURL        string `env:"REDIRECT_URL,required"`
}

type PostgreSQL struct {
	DBHost     string `env:"POSTGRES_HOST,required"`
	DBPort     int    `env:"POSTGRES_PORT,required"`
	DBUser     string `env:"POSTGRES_USER,required"`
	DBPassword string `env:"POSTGRES_PASSWORD,required"`
	DBName     string `env:"POSTGRES_DATABASE,required"`
}

type JWT struct {
	JwtSecretKey string `env:"JWT_SECRET_KEY,required"`
	JwtIssuer    string `env:"JWT_ISSUER,required"`
}

type CORS struct {
	FrontEndOrigin string `env:"FRONTEND_ORIGIN,required"`
}

type Scalingo struct {
	DatabaseURL string `env:"SCALINGO_POSTGRESQL_URL"`
}
