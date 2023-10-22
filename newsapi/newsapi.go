package newsapi

import (
	"strconv"
	"testtasktambov/models"

	"github.com/gofiber/fiber"
	"gopkg.in/reform.v1"
)

func SetupApiRoutes(app *fiber.App, db *reform.DB) {
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
