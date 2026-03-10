package config

type Config struct {
	TCP
	HTTPServer
	GoogleOAuth
	PostgreSQL
	JWT
	CORS
}

type TCP struct {
	TCPPort string `env:"TCP_SERVER_PORT,required"`
}

type HTTPServer struct {
	HTTPPort string `env:"HTTP_PORT,required"`
	Env      string `env:"APP_ENV,required"`
}

func (h HTTPServer) IsProduction() bool {
	return h.Env == "production"
}

// FrontEndOrigin returns the active frontend origin based on the current environment.
func (c *Config) FrontEndOrigin() string {
	if c.IsProduction() {
		return c.FrontEndOriginProd
	}
	return c.FrontEndOriginDev
}

type GoogleOAuth struct {
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	RedirectURLDev     string `env:"REDIRECT_URL_DEV"`
	RedirectURLProd    string `env:"REDIRECT_URL_PRODUCTION"`
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
	JwtSecretKey string `env:"JWT_SECRET_KEY,required"`
	JwtIssuer    string `env:"JWT_ISSUER,required"`
}

type CORS struct {
	FrontEndOriginDev  string `env:"FRONTEND_ORIGIN_DEV"`
	FrontEndOriginProd string `env:"FRONTEND_ORIGIN_PRODUCTION"`
}
