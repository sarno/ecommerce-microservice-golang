package config

import "github.com/spf13/viper"

type App struct {
	AppPort string `json:"app_port"`
	AppEnv  string `json:"app_env"`

	JwtSecretKey string `json:"jwt_secret_key"`
}

type Database struct {
	Host               string `json:"host"`
	Port               string    `json:"port"`
	User               string `json:"user"`
	Password           string `json:"password"`
	Name               string `json:"name"`
	MaxOpenConnections int    `json:"max_open_connections"`
	MaxIdleConnections int    `json:"max_idle_connections"`
}

type Redis struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

type RabbitMQ struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type Supabase struct {
	Url    string `json:"url"`
	Key    string `json:"key"`
	Bucket string `json:"bucket"`
}

type EmailConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Sending  string `json:"sending"`
	IsTLS    bool   `json:"is_tls"`
}

type Config struct {
	App      App      `json:"app"`
	Database Database `json:"database"`
	Redis    Redis    `json:"redis"`
	RabbitMQ RabbitMQ `json:"rabbitmq"`
	Storage  Supabase `json:"storage"`
	EmailConfig EmailConfig `json:"email_config"`
}


func NewConfig() *Config {
	return &Config{
		App: App{
			AppPort: viper.GetString("APP_PORT"),
			AppEnv:  viper.GetString("APP_ENV"),

			JwtSecretKey: viper.GetString("JWT_SECRET_KEY"),
		},
		Database: Database{
			Host:               viper.GetString("DATABASE_HOST"),
			Port:               viper.GetString("DATABASE_PORT"),
			User:               viper.GetString("DATABASE_USER"),
			Password:           viper.GetString("DATABASE_PASSWORD"),
			Name:               viper.GetString("DATABASE_NAME"),
			MaxOpenConnections: viper.GetInt("DAATABASE_MAX_OPEN_CONNECTIONS"),
			MaxIdleConnections: viper.GetInt("DAATABASE_MAX_IDLE_CONNECTIONS"),
		},
		Redis: Redis{
			Host: viper.GetString("REDIS_HOST"),
			Port: viper.GetString("REDIS_PORT"),
		},
		RabbitMQ: RabbitMQ{
			Host:     viper.GetString("RABBITMQ_HOST"),
			Port:     viper.GetString("RABBITMQ_PORT"),
			User:     viper.GetString("RABBITMQ_USER"),
			Password: viper.GetString("RABBITMQ_PASSWORD"),
		},
		Storage: Supabase{
			Url:    viper.GetString("SUPABASE_STORAGE_URL"),
			Key:    viper.GetString("SUPABASE_STORAGE_KEY"),
			Bucket: viper.GetString("SUPABASE_STORAGE_BUCKET"),
		},
		EmailConfig: EmailConfig{
			Host:     viper.GetString("EMAIL_HOST"),
			Port:     viper.GetInt("EMAIL_PORT"),
			Username: viper.GetString("EMAIL_USERNAME"),
			Password: viper.GetString("EMAIL_PASSWORD"),
			Sending:  viper.GetString("EMAIL_SENDING"),
			IsTLS:    viper.GetBool("EMAIL_IS_TLS"),
		},
	}
}
