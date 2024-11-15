package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jmoiron/sqlx"
)

type Link struct {
	Original string `json:"original" db:"original"`
	Key      string `json:"short" db:"short"`
}

type LinkReq struct {
	Original string `json:"original" db:"original"`
}

func main() {
	db, _ := sqlx.Connect("mysql", "root:root@tcp(127.0.0.1:9001)/shortener")

	app := fiber.New()

	app.Use(logger.New())

	app.Get("/", func(c *fiber.Ctx) error {
		links := []Link{}
		db.Select(&links, "SELECT original, short FROM map")

		return c.Status(http.StatusOK).JSON(links)
	})

	app.Get("/:key", func(c *fiber.Ctx) error {
		key := c.Params("key")

		var link Link
		db.Get(&link, "SELECT original, short FROM map WHERE short = ?", key)

		return c.Redirect(link.Original, http.StatusTemporaryRedirect)
	})

	app.Post("/", func(c *fiber.Ctx) error {
		var req LinkReq
		c.BodyParser(&req)

		randomString := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

		rand.Seed(time.Now().UnixNano())
		key := make([]byte, 6)
		for i := range key {
			key[i] = randomString[rand.Intn(len(randomString))]
		}
		shortKey := string(key)

		db.Query("INSERT INTO map (original, short) VALUES (?, ?)", req.Original, shortKey)

		return c.Status(http.StatusCreated).JSON(Link{
			Original: req.Original,
			Key:      shortKey,
		})
	})

	log.Println("Running on port 9000")
	app.Listen(":9000")
}
