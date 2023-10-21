package main

import (
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Создаем новый экземпляр Fiber
	app := fiber.New()

	// Обработчик GET-запросов на корневой путь "/"
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Запускаем сервер на порту 3000
	err := app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}
