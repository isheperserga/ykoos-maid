package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DiscordBotToken string
	DB              DBConfig
	Redis           RedisConfig
	HdevApiKey      string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

type Option func(*Config)

func WithDiscordBotToken(token string) Option {
	return func(c *Config) {
		c.DiscordBotToken = token
	}
}

func WithDBConfig(dbConfig DBConfig) Option {
	return func(c *Config) {
		c.DB = dbConfig
	}
}

func NewConfig(opts ...Option) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{
		DiscordBotToken: v.GetString("DISCORD_BOT_TOKEN"),
		DB: DBConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetString("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			Name:     v.GetString("DB_NAME"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("REDIS_HOST"),
			Port:     v.GetString("REDIS_PORT"),
			Password: v.GetString("REDIS_PASSWORD"),
		},
		HdevApiKey: v.GetString("HENRIKDEV_API_KEY"),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg, nil
}
