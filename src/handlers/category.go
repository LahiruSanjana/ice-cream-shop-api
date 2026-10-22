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

func GetCategoryById(c *fiber.Ctx) error {
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

func UpdateCategory(c *fiber.Ctx) error {
	var category models.Category
	id := c.Params("id")

	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
	}

	type CategoryRequest struct {
		Name *string `json:"name"`
	}

	var body CategoryRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Name != nil {
		if *body.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name cannot be empty"})
		}
		category.Name = *body.Name
	}

	if err := database.DB.Save(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update category"})
	}

	return c.JSON(fiber.Map{
		"message": "Category updated successfully",
		"category": category,
	})
}

func DeleteCategory(c *fiber.Ctx) error {
	var category models.Category
	id := c.Params("id")

	if err := database.DB.First(&category, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Category not found"})
	}

	if err := database.DB.Delete(&category).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete category"})
	}

	return c.JSON(fiber.Map{
		"message": "Category deleted successfully",
		"category": category,
	})
}