package routes

import (
	"main/src/handlers"
	"main/src/middleware"

	"github.com/gofiber/fiber/v2"
)

func OrderRoutes(router fiber.Router) {
	orderRouter := router.Group("/orders")
	orderRouter.Use(middleware.AuthMiddleware())
	orderRouter.Post("/", handlers.CreateOrder)
	orderRouter.Get("/latest", handlers.GetOrdersLatest)
	orderRouter.Get("/date", handlers.GetOrdersByDate)
	orderRouter.Get("/customer/:customer_id", handlers.GetOrdersByCustomer)
	orderRouter.Get("/", handlers.GetOrdersPaginated)
}
