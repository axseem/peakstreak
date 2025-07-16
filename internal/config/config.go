package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBUrl        string        `mapstructure:"DB_URL"`
	ServerPort   string        `mapstructure:"SERVER_PORT"`
	JWTSecret    string        `mapstructure:"JWT_SECRET"`
	JWTExpiresIn time.Duration `mapstructure:"JWT_EXPIRES_IN"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("JWT_EXPIRES_IN", "24h")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return
		}
	}

	err = viper.Unmarshal(&config)
	return
}
