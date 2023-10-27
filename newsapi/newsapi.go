package newsapi

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"testtasktambov/models"

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
	ID      int64  `json:"Id"`
	Title   string `json:"Title"`
	Content string `json:"Content"`
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
		newsList := make([]*models.NewsData, len(records))
		for i, record := range records {
			newsList[i] = record.(*models.NewsData)
		}

		// Отправить ответ клиенту
		return c.JSON(newsList)
	})
}

func SyncNewsWithFileDel(filePath string, db *reform.DB) error {
	// Чтение новостей из файла
	file, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var newsJsonArr []NewsJson
	err = json.Unmarshal(file, &newsJsonArr)
	if err != nil {
		return err
	}

	// Создание мапы новостей для быстрого поиска
	newsMap := make(map[int64]NewsJson)
	for _, news := range newsJsonArr {
		newsMap[news.ID] = news
	}

	return db.InTransaction(func(tx *reform.TX) error {
		// Загрузка всех новостей из базы данных
		records, err := tx.SelectAllFrom(models.NewsTable, "")
		if err != nil {
			return err
		}

		var multiErr error
		for _, record := range records {
			news := record.(*models.NewsData)

			// Проверка: есть ли новость из базы данных в загруженных из файла
			if _, exists := newsMap[news.ID]; !exists {
				err = tx.Delete(news)
				if err != nil {
					multiErr = multierr.Combine(multiErr, err)
					log.Printf("Failed to delete news with ID %v. Error: %v\n", news.ID, err)
				} else {
					log.Printf("Successfully deleted news with ID %v.\n", news.ID)
				}
			}

		}
		return multiErr
	})
}
