package routes

import (
	"main/src/handlers"

	"github.com/gofiber/fiber/v2"
)

func RegisterAuthRoutes(router fiber.Router) {
	router.Post("/register", handlers.RegisterUser)
	router.Post("/login", handlers.LoginUser)
}
