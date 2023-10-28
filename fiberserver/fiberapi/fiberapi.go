package fiberapi

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/MajotraderLucky/TambovRepo/models.go"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/multierr"
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

		// Удалить старые связи
		_, err = db.DeleteFrom(models.NewsCategoryTable, "news_id", id)
		if err != nil {
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

		// Возврат значения
		return nil
	})

	app.Get("/list", func(c *fiber.Ctx) error {
		// Получить все записи из базы данных
		records, err := db.SelectAllFrom(models.NewsTable, "")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		// Обход всех записей и преобразование их в нужный тип
		newsList := make([]*models.NewsData, len(records))
		for i, record := range records {
			newsList[i] = record.(*models.NewsData)
		}

		// Отправить ответ клиенту
		return c.JSON(newsList)
	})
}

type NewsJson struct {
	ID         int64   `json:"Id"`
	Title      string  `json:"Title"`
	Content    string  `json:"Content"`
	Categories []int64 `json:"Categories"`
}

type ResponseJson struct {
	Success bool       `json:"Success"`
	News    []NewsJson `json:"News"`
}

func LoadNewsFromFile(filePath string, db *reform.DB) error {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var newsJsonArr []NewsJson
	err = json.Unmarshal(file, &newsJsonArr)
	if err != nil {
		return err
	}

	return db.InTransaction(func(tx *reform.TX) error {
		var multiErr error
		for _, n := range newsJsonArr {
			_, err := tx.Exec(`INSERT INTO News (id, title, content) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE title = ?, content = ?`,
				n.ID, n.Title, n.Content, n.Title, n.Content)
			if err != nil {
				multiErr = multierr.Combine(multiErr, err)
				log.Printf("Failed to save or update newsData with ID %v. Error: %v\n", n.ID, err)
			} else {
				log.Printf("Successfully saved or updated newsData with ID %v.\n", n.ID)
			}
		}
		return multiErr
	})
}

func SetupApiRouteGetList(app *fiber.App, db *reform.DB) {
	app.Get("/list", func(c *fiber.Ctx) error {
		// Получить все записи из базы данных
		records, err := db.SelectAllFrom(models.NewsTable, "")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}

		// Обход всех записей и преобразование их в нужный тип
		newsList := make([]NewsJson, len(records))
		for i, record := range records {
			news := record.(*models.NewsData)
			apiNews := NewsJson{ID: news.ID, Title: news.Title, Content: news.Content}

			// Запрос категорий для каждой новости
			categories, err := db.FindAllFrom(models.NewsCategoryTable, "NewsId", news.ID)
			if err != nil {
				log.Printf("Failed to load categories for news with ID %v. Error: %v\n", news.ID, err)
				continue
			}

			for _, cat := range categories {
				category := cat.(*models.NewsCategory)
				apiNews.Categories = append(apiNews.Categories, category.CategoryID)
			}

			newsList[i] = apiNews
		}

		// Создать и заполнить структуру ResponseJson
		response := ResponseJson{
			Success: len(records) != 0, // Успешно, если есть хотя бы одна новость.
			News:    newsList,
		}

		// Отправить ответ клиенту.
		return c.JSON(response)
	})
}
