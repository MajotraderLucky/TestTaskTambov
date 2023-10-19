package main

import (
	"database/sql"
	"log"
	"strconv"
	"testtasktambov/models"

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

	log.Println("Hello, Tambov!")

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
	if err := createNewsWithCategories(session, newsData); err != nil {
		log.Fatal(err)
	}

	logger.CleanLogCountLines(100)
}

// Создание новости и связывание ее с категориями
func createNewsWithCategories(session *reform.DB, newsData models.NewsData) error {
	tx, err := session.Begin()
	if err != nil {
		return err
	}

	// Создать новую новость
	news := &models.NewsData{
		Title:   newsData.Title,
		Content: newsData.Content,
	}

	if err := tx.Insert(news); err != nil {
		tx.Rollback()
		return err
	}

	// Пройти по списку категорий и создать связи для новости
	for _, category := range newsData.Categories {
		intCat, err := strconv.ParseInt(category, 10, 64)
		if err != nil {
			tx.Rollback()
			return err
		}

		newsCategory := &models.NewsCategory{
			NewsID:     news.ID,
			CategoryID: intCat,
		}

		if err := tx.Insert(newsCategory); err != nil {
			tx.Rollback()
			return err
		}
	}

	// Зафиксировать транзакцию при успешном создании новости и связей
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
