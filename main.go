package main

import (
	"database/sql"
	"log"
	"testtasktambov/models"
	"testtasktambov/newsapi"
	"testtasktambov/newsdb"
	"time"

	"github.com/MajotraderLucky/Utils/logger"
	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/mysql"
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
	// Set the maximum number of concurrently open connections (e.g. 10).
	// If <= 0, then there is no limit on the number of open connections.
	db.SetMaxOpenConns(10)

	// Set the maximum number of idle connections in the connection pool (e.g. 5).
	// If <= 0, no idle connections are retained.
	db.SetMaxIdleConns(5)

	// Set the maximum amount of time a connection may be reused (e.g. 30 minutes).
	// Expired connections may be closed lazily before reuse.
	// If <= 0, connections are reused forever.
	db.SetConnMaxLifetime(time.Duration(30) * time.Minute)
	defer db.Close()

	session := reform.NewDB(db, mysql.Dialect, reform.NewPrintfLogger(log.Printf))

	err = newsapi.LoadNewsFromFileCat("news.json", session)
	if err != nil {
		log.Fatalf("Failed to load news from file: %v", err)
	}

	// Синхронизация новостей из файла с базой данных
	err = newsapi.SyncNewsWithFileDel("news.json", session)
	if err != nil {
		log.Fatalf("Failed to sync news with file: %v", err)
	}

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
