package handlers

import (
	"github.com/gofiber/fiber/v2"
	"main/src/database"
	"main/src/models"
)

func CreateItem(c *fiber.Ctx) error {
	type CreateItemRequest struct {
		Name        string  `json:"name"`
		Description string  `json:"description"`
		Price       float64 `json:"price"`
		Quantity    int     `json:"quantity"`
		CategoryID  uint    `json:"category_id"`
	}

	var body CreateItemRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Name == "" || body.Price <= 0 || body.Quantity < 0 || body.CategoryID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "name, price, quantity, and category_id are required and must be valid"})
	}

	item := models.Item{
		Name:        body.Name,
		Description: body.Description,
		Price:       body.Price,
		Quantity:    body.Quantity,
		CategoryID:  body.CategoryID,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create item"})
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}