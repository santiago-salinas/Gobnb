// config/config.go
package config

import (
	"log"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config/")

	viper.AutomaticEnv() // Automatically override values from environment variables

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}

// GetConfig is a convenience function to access the viper instance
func GetConfig() *viper.Viper {
	return viper.GetViper()
}
