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

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	return config
}
