package handlers

import (
	"github.com/gofiber/fiber/v2"
	"main/src/database"
	"main/src/models"
)

func GetAllCustomers(c *fiber.Ctx) error {
	var customers []models.Customer
	if err := database.DB.Find(&customers).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve customers"})
	}
	return c.JSON(fiber.Map{"customers": customers})
}

func GetCustomerById(c *fiber.Ctx) error {
	var customer models.Customer
	id := c.Params("id")
	if err := database.DB.First(&customer, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer not found"})
	}
	return c.JSON(fiber.Map{"customer": customer})
}

func CreateCustomer(c *fiber.Ctx) error {
	type CustomerRequest struct {
		Name    string `json:"name"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Address string `json:"address"`
	}

	var body CustomerRequest

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Name == "" || body.Email == "" || body.Phone == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name, email, and phone are required"})
	}

	customer := models.Customer{
		Name:    body.Name,
		Email:   body.Email,
		Phone:   body.Phone,
		Address: body.Address,
	}

	if err := database.DB.Create(&customer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create customer"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Customer created successfully", "customer": customer})
}

func UpdateCustomer(c *fiber.Ctx) error {
	var customer models.Customer
	id := c.Params("id")

	if err := database.DB.First(&customer, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer not found"})
	}

	type CustomerUpdateRequest struct {
		Name    *string `json:"name"`
		Email   *string `json:"email"`
		Phone   *string `json:"phone"`
		Address *string `json:"address"`
	}

	var body CustomerUpdateRequest
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if body.Name == nil && body.Email == nil && body.Phone == nil && body.Address == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "At least one field must be provided for update"})
	}

	if body.Name != nil {
		if *body.Name == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name cannot be empty"})
		}
		customer.Name = *body.Name
	}

	if body.Email != nil {
		if *body.Email == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Email cannot be empty"})
		}
		customer.Email = *body.Email
	}

	if body.Phone != nil {
		if *body.Phone == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Phone cannot be empty"})
		}
		customer.Phone = *body.Phone
	}

	if body.Address != nil {
		customer.Address = *body.Address
	}

	if err := database.DB.Save(&customer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update customer"})
	}

	return c.JSON(fiber.Map{
		"message": "Customer updated successfully", 
		"customer": customer,
	})
}

func DeleteCustomer(c *fiber.Ctx) error {
	var customer models.Customer
	id := c.Params("id")

	if err := database.DB.First(&customer, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer not found"})
	}

	if err := database.DB.Delete(&customer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete customer"})
	}

	return c.JSON(fiber.Map{
		"message": "Customer deleted successfully",
	})
}