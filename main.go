package main

import (
	"log"

	"github.com/MajotraderLucky/Utils/logger"
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

	logger.CleanLogCountLines(100)
}
