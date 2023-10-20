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
	"github.com/gofiber/fiber/v2"
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

	for _, intCat := range newsData.Categories {
		// создание связей с категорией
		newsCategory := &models.NewsCategory{
			NewsID:     int64(news.ID),
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

func setupApiRoutes(app *fiber.App, db *reform.DB) {
	// Обновить новость по id
	app.Post("/edit/:id", func(c *fiber.Ctx) error {
		// Принимаем входные данные
		var input models.NewsData

		if err := c.BodyParser(&input); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		// Преобразование строкового ID в int64
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		// Находим существующую запись
		existingEntity, err := db.FindOneFrom(models.NewsTable, "id", id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("News not found")
		}

		news := existingEntity.(*models.NewsData)

		// Updating news
		if input.Title != "" {
			news.Title = input.Title
		}

		if input.Content != "" {
			news.Content = input.Content
		}

		// Обновление категорий
		for _, categoryID := range input.Categories {
			// Создать новую связь
			newsCategory := &models.NewsCategory{
				NewsID:     id,
				CategoryID: categoryID,
			}

			// Добавить связь в базу данных
			err := db.Insert(newsCategory)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
			}
		}

		if err := db.Update(news); err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to update news")
		}

		return c.JSON(news)
	})

	// Обновление категорий новости
	app.Get("/categories/:id", func(c *fiber.Ctx) error {
		// Получить ID новости
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil {
			return err
		}

		// Получить входные данные
		var input []int
		if err := c.BodyParser(&input); err != nil {
			return err
		}

		// Обновление категорий новости
		for _, categoryID := range input {
			// Создать новую связь
			var newCategory = &models.NewsCategory{
				NewsID:     id,
				CategoryID: int64(categoryID),
			}

			// Добавить связь в базу данных
			err := db.Insert(newCategory)
			if err != nil {
				return err
			}
		}

		// Удалить старые связи
		_, err = db.DeleteFrom(models.NewsCategoryTable, "news_id", id)
		if err != nil {
			return err
		}

		// Обработка ошибок
		if err != nil {
			return err
		}

		// Возврат значения
		return nil
	})
}
