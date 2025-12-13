package config

import "github.com/spf13/viper"

type App struct {
	AppPort string `json:"app_port"`
	AppEnv string `json:"app_env"`

	JwtSecretKey string `json:"jwt_secret"`
	JwtExpire int `json:"jwt_expire"`
	JwtIssuer string `json:"jwt_issuer"`

	UrlFrontFE string `json:"url_front_fe"`
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

type Redis struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type RabbitMQ struct {
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
	Password string `json:"password"`
}

type Supabase struct {
	Url string `json:"url"`
	Key string `json:"key"`
	Bucket string `json:"bucket"`
}

type Config struct {
	App App `json:"app"`
	Database Database `json:"database"`
	Redis Redis `json:"redis"`
	RabbitMQ RabbitMQ `json:"rabbitmq"`
	Storage Supabase `json:"storage"`
}

func NewConfig() *Config {
	return &Config{
		App: App{
			AppPort: viper.GetString("APP_PORT"),
			AppEnv: viper.GetString("APP_ENV"),
			JwtSecretKey: viper.GetString("JWT_SECRET"),
			JwtExpire: viper.GetInt("JWT_EXPIRATION"),
			JwtIssuer: viper.GetString("JWT_ISSUER"),
			UrlFrontFE: viper.GetString("URL_FRONT_FE"),
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
		Redis: Redis{
			Host: viper.GetString("REDIS_HOST"),
			Port: viper.GetString("REDIS_PORT"),
		},
		RabbitMQ: RabbitMQ{
			Host: viper.GetString("RABBITMQ_HOST"),
			Port: viper.GetString("RABBITMQ_PORT"),
			User: viper.GetString("RABBITMQ_USER"),
			Password: viper.GetString("RABBITMQ_PASSWORD"),
		},
		Storage: Supabase{
			Url: viper.GetString("SUPABASE_STORAGE_URL"),
			Key: viper.GetString("SUPABASE_STORAGE_KEY"),
			Bucket: viper.GetString("SUPABASE_STORAGE_BUCKET"),
		},
	}
}