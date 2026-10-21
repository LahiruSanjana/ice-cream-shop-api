package routes

import (
	"main/src/handlers"
	"main/src/middleware"
	"github.com/gofiber/fiber/v2"
)

func ItemRoutes(router fiber.Router) {
	itemRouter := router.Group("/items")
	itemRouter.Use(middleware.AuthMiddleware()) // Apply the AuthMiddleware to all item routes
	itemRouter.Post("/",middleware.RoleMiddleware("owner","manager"), handlers.CreateItem) // Apply the RoleMiddleware to all item routes, allowing only admin users
}