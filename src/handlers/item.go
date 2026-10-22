package handlers

import (
	"github.com/gofiber/fiber/v2"
	"main/src/database"
	"main/src/models"
)

func GetItems(c *fiber.Ctx) error {
	var items []models.Item
	if err := database.DB.Find(&items).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch items"})
	}
	return c.JSON(items)
}

func GetItemById(c *fiber.Ctx) error {
	var item models.Item
	id := c.Params("id")

	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
	}

	return c.JSON(item)
}

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

func UpdateItem(c *fiber.Ctx) error {
	var item models.Item
	id := c.Params("id")

	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
	}

	type UpdateItemRequest struct {
		Name        *string  `json:"name"`
		Description *string  `json:"description"`
		Price       *float64 `json:"price"`
		Quantity    *int     `json:"quantity"`
	}

	var body UpdateItemRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Name == nil && body.Description == nil && body.Price == nil && body.Quantity == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one field must be provided for update"})
	}

	if body.Name != nil {
		if *body.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name cannot be empty"})
		}
		item.Name = *body.Name
	}

	if body.Description != nil {
		item.Description = *body.Description
	}

	if body.Price != nil {
		if *body.Price <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Price must be a positive number"})
		}
		item.Price = *body.Price
	}

	if body.Quantity != nil {
		if *body.Quantity < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Quantity cannot be negative"})
		}
		item.Quantity = *body.Quantity
	}

	if err := database.DB.Save(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update item"})
	}

	return c.JSON(fiber.Map{
		"message": "Item updated successfully",
		"item":    item,
	})
}

func DeleteItem(c *fiber.Ctx) error {
	var item models.Item
	id := c.Params("id")

	if err := database.DB.First(&item, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Item not found"})
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete item"})
	}

	return c.JSON(fiber.Map{
		"message": "Item deleted successfully",
	})
}
