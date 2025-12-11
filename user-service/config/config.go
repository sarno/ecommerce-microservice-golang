package config

import "github.com/spf13/viper"

type App struct {
	AppPort string `json:"app_port"`
	AppEnv string `json:"app_env"`

	JwtSecretKey string `json:"jwt_secret"`
	JwtExpire int `json:"jwt_expire"`
	JwtIssuer string `json:"jwt_issuer"`
}

type Database struct {
	Host string `json:"host"`
	Port int `json:"port"`
	User string `json:"user"`
	Password string `json:"password"`
	Name string `json:"name"`
	MaxOpenConnections int `json:"max_open_connections"`
	MaxIdleConnections int `json:"max_idle_connections"`
}

type Config struct {
	App App `json:"app"`
	Database Database `json:"database"`
}

func NewConfig() *Config {
	return &Config{
		App: App{
			AppPort: viper.GetString("APP_PORT"),
			AppEnv: viper.GetString("APP_ENV"),
			JwtSecretKey: viper.GetString("JWT_SECRET"),
			JwtExpire: viper.GetInt("JWT_EXPIRATION"),
			JwtIssuer: viper.GetString("JWT_ISSUER"),
		},
		Database: Database{
			Host: viper.GetString("DATABASE_HOST"),
			Port: viper.GetInt("DATABASE_PORT"),
			User: viper.GetString("DATABASE_USER"),
			Password: viper.GetString("DATABASE_PASSWORD"),
			Name: viper.GetString("DATABASE_NAME"),
			MaxOpenConnections: viper.GetInt("DAATABASE_MAX_OPEN_CONNECTIONS"),
			MaxIdleConnections: viper.GetInt("DAATABASE_MAX_IDLE_CONNECTIONS"),
		},
	}
}