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

	// Устанавливаем маршруты для API /list
	fiberapi.SetupApiRouteGetList(app, rDB)

	// Add news handler /edit/:id
	fiberapi.SetupApiRouteEdit(app, rDB)

	// Show jwt token
	jwtToken, err := fiberapi.GenerateJWTToken()
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(jwtToken)

	// Запускаем сервер на порту 3000
	err = app.Listen(":3000")
	if err != nil {
		panic(err)
	}
}

// curl -X POST -H "Content-Type: application/json" -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE2OTkxNjEyNDksIm5hbWUiOiJKb2huIERvZSIsInN1YiI6IjEyMzQ1Njc4OTAifQ.YjFcGpSub2fROKCuMlkIKcmr-YwJaMV59p1ARM4dcMOs2rwcDN2emv0SvNcvxMP4jzgoCq_h8udYsjZrs7PQXHYjJbnDaWkWbalklcOsrEUfd3gmzXe-tM9usUK7mOvGYR9rK64Gn21yy220SskqfXIEdVK200ynmdgsddFuiNDUM92SiwxDfisOkf51TmvfiXgy2f2mZ8V0esbn5sK9-ho9oP-YveiQSfT3KCW-2kEtKAxVnQW77t3fn81OpWtydgYtbgPhvCEcq9Dk-WTu7Xv-vcOU8i8zcczitZsOe780DxkJj47YWaHFFbOOA9OSPUc313K5GlDTnzlgzu-t9g" -d '{"title":"new title", "content":"new content777"}' http://127.0.0.1:3000/edit/2
