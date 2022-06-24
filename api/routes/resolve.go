package routes

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/kaliadmen/url_shortener/database"
	"log"
)

func ResolveURL(c *fiber.Ctx) error {
	url := c.Params("url")

	r := database.CreateClient(0)

	defer func(r *redis.Client) {
		err := r.Close()
		if err != nil {
			log.Println(err)
		}
	}(r)

	value, err := r.Get(database.Ctx, url).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "shorten url not found in database"})
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "cannot connect to database"})
	}

	r2 := database.CreateClient(1)

	defer func(r2 *redis.Client) {
		err := r2.Close()
		if err != nil {
			log.Println(err)
		}
	}(r2)

	_ = r2.Incr(database.Ctx, "counter")

	return c.Redirect(value, 301)

}
