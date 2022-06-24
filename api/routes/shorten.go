package routes

import (
	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/kaliadmen/url_shortener/database"
	"github.com/kaliadmen/url_shortener/helpers"
	"log"
	"os"
	"strconv"
	"time"
)

type request struct {
	URL         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	URL            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`
	XRateLimitRate time.Duration `json:"rate_limit_reset"`
}

func ShortenURL(c *fiber.Ctx) error {
	body := new(request)

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	//rate limiting
	r2 := database.CreateClient(1)

	defer func(r2 *redis.Client) {
		err := r2.Close()
		if err != nil {
			log.Println(err)
		}
	}(r2)

	val, err := r2.Get(database.Ctx, c.IP()).Result()
	if err != redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else {
		val, _ = r2.Get(database.Ctx, c.IP()).Result()
		intVal, _ := strconv.Atoi(val)

		if intVal <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Rate limit exceeded", "rate_limit_reset": limit / time.Nanosecond / time.Minute})
		}
	}

	//validate URL
	if !govalidator.IsURL(body.URL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid URL"})
	}

	//check if domain error
	if !helpers.RemoveDomainError(body.URL) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Invalid URL"})
	}

	//enforce SSL
	body.URL = helpers.EnforceHTTP(body.URL)

	r2.Decr(database.Ctx, c.IP())

	return nil
}
