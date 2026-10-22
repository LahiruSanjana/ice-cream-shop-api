package routes

import (
	"main/src/handlers"
	"main/src/middleware"
	"github.com/gofiber/fiber/v2"
)

func CustomerRoutes(router fiber.Router) {
	customerRouter := router.Group("/customers")
	customerRouter.Use(middleware.AuthMiddleware())
	customerRouter.Get("/", handlers.GetAllCustomers)
	customerRouter.Get("/:id", handlers.GetCustomerById)
	customerRouter.Post("/", handlers.CreateCustomer)
	customerRouter.Put("/:id", handlers.UpdateCustomer)
	customerRouter.Delete("/:id",middleware.RoleMiddleware("owner"), handlers.DeleteCustomer) // Apply the RoleMiddleware to all customer routes, allowing only owner
}