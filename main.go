package main

import (
	"os"
	"path"
	"regexp"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/keyauth"
)

func main() {
	app := fiber.New()
	root := os.Getenv("ROOT")
	if root == "" {
		root = "."
	}

	corsOrigins := os.Getenv("CORS_ORIGINS")

	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigins,
		AllowHeaders:     "Authorization, If-Modified-Since",
		AllowCredentials: false,
	}))

	accessToken := os.Getenv("ACCESS_TOKEN")

	if accessToken != "" {
		app.Use(keyauth.New(keyauth.Config{
			AuthScheme: "Token",
			Validator: func(ctx *fiber.Ctx, token string) (bool, error) {
				return token == accessToken, nil
			},
		}))
	}

	// sadly we cant use app.Static() because we use if-modified-since
	app.Get("/*", func(c *fiber.Ctx) error {
		// this may be vulnerable to path traversal attacks
		// but we're behind a reverse proxy which should take care of those
		file := path.Join(root, c.Path())
		stat, err := os.Stat(file)
		if err != nil { // file not found
			return c.SendStatus(fiber.StatusNotFound)
		}

		// we only support if-modified-since because thats the only one the client application
		// actually sends
		ifModifiedSince := c.Get("If-Modified-Since")
		if ifModifiedSince != "" {
			ifModifiedSinceTime, err := time.Parse(time.RFC1123, ifModifiedSince)
			if err == nil && stat.ModTime().Before(ifModifiedSinceTime) {
				return c.SendStatus(fiber.StatusNotModified)
			}
		}

		// we dont use compression because the files are encrypted
		return c.SendFile(file)
	})

	// we only allow writing to these files:
	allowed := regexp.MustCompile(`^\/[a-z\d_-]+-(starstore|readprogress)\.json$`)
	// [a-z\d_-]+ is the username
	// starstore.json is the encrypted file containing the user's starred items
	// readprogress.json is the encrypted file containing the user's read progress
	// we dont allow writing to any other files because that would be a security risk
	// the data is end to end encrypted, so we dont need to worry about the server being compromised
	app.Put("/*", func(c *fiber.Ctx) error {
		if !allowed.MatchString(c.Path()) {
			return c.SendStatus(fiber.StatusForbidden)
		}
		file := path.Join(root, c.Path())
		err := os.WriteFile(file, c.Body(), 0644)
		if err != nil {
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.SendStatus(fiber.StatusNoContent)
	})

	app.Listen(":3000")
}
