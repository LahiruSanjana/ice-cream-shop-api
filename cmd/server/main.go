package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"main/src/database"
	"main/src/routes"
)

func main() {
	database.ConnectDatabase()
	database.SeedDatabase(database.DB)

	app := fiber.New()
	app.Use(logger.New())
	api := app.Group("/api")
	routes.RegisterAuthRoutes(api.Group("/auth"))
	routes.ItemRoutes(api)
	routes.CategoryRoutes(api)
	routes.OrderRoutes(api)

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("App is running healthy.")
	})

	log.Fatal(app.Listen(":8000"))
}
