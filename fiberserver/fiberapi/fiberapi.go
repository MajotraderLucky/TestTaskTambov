package fiberapi

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/MajotraderLucky/TambovRepo/models.go"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/multierr"
	"gopkg.in/reform.v1"
)

var jwtKey = []byte("my_secret_key")

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

func SetupApiRouteEdit(app *fiber.App, db *reform.DB) {
	app.Post("/edit/:id", IsAuthorized, func(c *fiber.Ctx) error {
		id, err := strconv.ParseInt(c.Params("id"), 10, 64)
		if err != nil || id <= 0 {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid ID")
		}

		var news NewsJson
		err = c.BodyParser(&news)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Failed to parse body")
		}

		// Get existing news record
		record, err := db.FindByPrimaryKeyFrom(models.NewsTable, id)
		if err != nil {
			return c.Status(fiber.StatusNotFound).SendString("News with specified ID not found")
		}
		existingNews := record.(*models.NewsData)

		// Update existingNews with new values from news
		existingNews.Title = news.Title
		existingNews.Content = news.Content
		// Note: Update all the other fields as necessary

		// Update news record in the database
		err = db.Update(existingNews)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("Failed to update news")
		}

		// Return updated news
		return c.JSON(existingNews)
	})
}

func IsAuthorized(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})

	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	return c.Next()
}

type MyCustomClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewCreateJWT() (string, error) {
	expirationTime := time.Now().Add(60 * time.Minute)

	claims := &MyCustomClaims{
		Username: "exampleUser",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)

	if err != nil {
		return "", err
	}

	fmt.Println(tokenString)
	return tokenString, nil
}
