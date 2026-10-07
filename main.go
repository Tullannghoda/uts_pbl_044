package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"uts-044/config"
	"uts-044/route"
	"uts-044/seeder"
)

func main() {
	config.InitJWT()
	pool := config.NewDB()
	defer pool.Close()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			seeder.Migrate(pool)
			return
		case "seed":
			seeder.Seed(pool)
			return
		}
	}

	app := fiber.New(fiber.Config{
		AppName: "SIAKAD Mini - UTS Backend",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())

	route.Setup(app, pool)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	fmt.Printf("Server berjalan di port %s\n", port)
	log.Fatal(app.Listen(":" + port))
}
