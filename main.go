package main

import (
	"log"

	"github.com/MajotraderLucky/Utils/logger"
	"github.com/spf13/viper"
)

func main() {
	logger := logger.Logger{}
	err := logger.CreateLogsDir()
	if err != nil {
		log.Fatal(err)
	}
	err = logger.OpenLogFile()
	if err != nil {
		log.Fatal(err)
	}
	logger.SetLogger()
	log.Println()
	logger.LogLine()

	log.Println("Hello, Tambov!")

	// Set the configuration file name and format
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set the path to search for the configuration file
	viper.AddConfigPath(".")
	// You can use other paths for searching if needed
	// viper.AddConfigPath("/etc/appname")

	// Read the configuration file
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read the configuration file: %s", err)
	}

	logger.CleanLogCountLines(100)
}
