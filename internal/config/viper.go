package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// NewViper is a function to load config from .env file
func NewViper() *viper.Viper {
	config := viper.New()
	config.SetConfigFile(".env")
	config.SetConfigType("env")
	config.AutomaticEnv()
	config.SetDefault("APP_NAME", "Golang Clean Architecture")
	config.SetDefault("WEB_PORT", "3001")

	err := config.ReadInConfig()
	if err != nil {
		// Don't panic if .env file doesn't exist, just use defaults and env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			panic(fmt.Errorf("Fatal error config file: %w \n", err))
		}
	}

	return config
}
