package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"github.com/kaliadmen/url_shortener/routes"
	"log"
	"os"
)

func setupRouter(app *fiber.App) {
	app.Post("/api/v1", routes.ShortenURL)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	app := fiber.New()

	app.Use(logger.New())

	setupRouter(app)

	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}
