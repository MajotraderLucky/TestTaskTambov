package main

import (
	"database/sql"
	"fiberserver/fiberapi"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"gopkg.in/reform.v1"
	"gopkg.in/reform.v1/dialects/mysql"
)

// func main() {
// 	// Создаем новый экземпляр Fiber
// 	app := fiber.New()

// 	// Обработчик GET-запросов на корневой путь "/"
// 	app.Get("/", func(c *fiber.Ctx) error {
// 		return c.SendString("Hello, World!")
// 	})

// 	// Запускаем сервер на порту 3000
// 	err := app.Listen(":3000")
// 	if err != nil {
// 		panic(err)
// 	}
// }

func main() {
	// Создаем новый экземпляр Fiber
	app := fiber.New()

	// Создаем подключение к базе данных
	db, err := sql.Open("mysql", "myuser:mypassword@tcp(db:3306)/mydb")
	if err != nil {
		log.Fatal(err)
	}

	// Set the maximum number of concurrently open connections (e.g. 10)
	db.SetMaxOpenConns(10)

	// Set the maximum number of idle connections in the connection pool (e.g. 5)
	db.SetMaxIdleConns(5)

	// Set the maximum amount of time a connection may be reused (e.g. 30 minutes)
	db.SetConnMaxLifetime(time.Duration(30) * time.Minute)

	// Создаем экземпляр reform.DB, используя нашу настройку соединения
	rDB := reform.NewDB(db, mysql.Dialect, nil)

	// Закрываем базу данных, когда завершим
	defer db.Close()

	// Устанавливаем маршруты для API
	fiberapi.SetupApiRouteGetList(app, rDB)

	// Запускаем сервер на порту 3000
	err = app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
