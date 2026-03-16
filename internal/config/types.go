package config

type Config struct {
	TCP
	HTTPServer
	GoogleOAuth
	PostgreSQL
	JWT
	CORS
	EnvLoad
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

func (c *Config) LoadEnv() {
	if c.IsProduction() {
		c.DBName = c.ProdDBName
		c.Cors = c.FrontEndOriginProd
		c.RedirectURL = c.RedirectURLProd
	} else {
		c.DBName = c.DevDBName
		c.Cors = c.FrontEndOriginDev
		c.RedirectURL = c.RedirectURLDev
	}
}

type TCP struct {
	TCPPort string `env:"TCP_SERVER_PORT,required"`
}

type HTTPServer struct {
	HTTPPort string `env:"HTTP_PORT,required"`
}

type GoogleOAuth struct {
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID,required"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET,required"`
	RedirectURL        string `env:"REDIRECT_URL"`
}

type PostgreSQL struct {
	DBHost     string `env:"POSTGRES_HOST"`
	DBPort     int    `env:"POSTGRES_PORT"`
	DBUser     string `env:"POSTGRES_USER"`
	DBPassword string `env:"POSTGRES_PASSWORD"`
	DBName     string `env:"POSTGRES_DATABASE"`
}

type JWT struct {
	JwtSecretKey string `env:"JWT_SECRET_KEY,required"`
	JwtIssuer    string `env:"JWT_ISSUER,required"`
}

type CORS struct {
	Cors string `env:"CORS"`
}

type EnvLoad struct {
	Env                string `env:"APP_ENV,required"`
	ProdDBName         string `env:"PRODUCTION_DBNAME,required"`
	DevDBName          string `env:"DEV_DBNAME,required"`
	FrontEndOriginDev  string `env:"FRONTEND_ORIGIN_DEV"`
	FrontEndOriginProd string `env:"FRONTEND_ORIGIN_PRODUCTION"`
	RedirectURLDev     string `env:"REDIRECT_URL_DEV"`
	RedirectURLProd    string `env:"REDIRECT_URL_PRODUCTION"`
}
