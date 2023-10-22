package main

import (
	"database/sql"
	"log"
	"testtasktambov/models"
	"testtasktambov/newsdb"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/mysql"

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

	logger.LogLine()

	db, err := sql.Open("mysql", "myuser:mypassword@tcp(db:3306)/mydb")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	session := reform.NewDB(db, mysql.Dialect, reform.NewPrintfLogger(log.Printf))

	// Define a news item and its categories
	newsData := models.NewsData{
		Title:   "Sample Title",
		Content: "Sample Content",
	}

	// Create the news item and its categories
	if err := newsdb.CreateNewsWithCategories(session, newsData); err != nil {
		log.Fatal(err)
	}

	logger.CleanLogCountLines(100)
}
