package handlers

import (
	"github.com/gofiber/fiber/v2"
	"main/src/database"
	"main/src/models"
)

func GetCategories(c *fiber.Ctx) error {
	var categories []models.Category
	if err := database.DB.Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not find categories"})
	}

	return c.JSON(categories)
}

func GetCategory(c *fiber.Ctx) error {
	var category models.Category
	id := c.Params("id")
	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
	}
	return c.JSON(category)
}

func CreateCategory(c *fiber.Ctx) error {
	type CategoryRequest struct {
		Name string `json:"name"`
	}

	var body CategoryRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Category name is required"})
	}

	category := models.Category{
		Name: body.Name,
	}

	if err := database.DB.Create(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create category"})
	}

	return c.Status(fiber.StatusCreated).JSON(category)
}