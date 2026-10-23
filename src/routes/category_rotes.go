package routes

import (
	"github.com/gofiber/fiber/v2"
	"main/src/handlers"
	"main/src/middleware"
)

func CategoryRoutes(router fiber.Router) {
	categoryRouter := router.Group("/categories")
	categoryRouter.Use(middleware.AuthMiddleware())
	categoryRouter.Post("/", middleware.RoleMiddleware("owner", "manager"), handlers.CreateCategory)
	categoryRouter.Get("/", handlers.GetCategories)
	categoryRouter.Get("/:id", handlers.GetCategoryById)
	categoryRouter.Put("/:id", middleware.RoleMiddleware("owner", "manager"), handlers.UpdateCategory)
	categoryRouter.Delete("/:id", middleware.RoleMiddleware("owner"), handlers.DeleteCategory)
}
