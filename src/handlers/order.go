package handlers

import (
	"errors"
	"fmt"
	"main/src/database"
	"main/src/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var (
	errInvalidOrder      = errors.New("invalid order")
	errInsufficientStock = errors.New("insufficient stock")
	errCustomerNotFound  = errors.New("customer not found")
	errItemNotFound      = errors.New("item not found")
)

func CreateOrder(c *fiber.Ctx) error {
	type OrderItemRequest struct {
		ItemID   uint `json:"item_id"`
		Quantity int  `json:"quantity"`
	}

	type CreateOrderRequest struct {
		CustomerID uint               `json:"customer_id"`
		Items      []OrderItemRequest `json:"items"`
	}

	var body CreateOrderRequest

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if len(body.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Order must contain at least one item"})
	}

	if body.CustomerID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CustomerID is required"})
	}

	var calculatedTotalPrice float64
	var orderItemsToCreate []models.OrderItem

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		if err := tx.First(&customer, body.CustomerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %d", errCustomerNotFound, body.CustomerID)
			}
			return err
		}

		for _, orderItem := range body.Items {
			if orderItem.ItemID == 0 || orderItem.Quantity <= 0 {
				return fmt.Errorf("%w: each item must have a valid item_id and a positive quantity", errInvalidOrder)
			}

			var item models.Item
			if err := tx.First(&item, orderItem.ItemID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return fmt.Errorf("%w: %d", errItemNotFound, orderItem.ItemID)
				}
				return err
			}

			result := tx.Model(&models.Item{}).
				Where("id = ? AND quantity >= ?", orderItem.ItemID, orderItem.Quantity).
				Update("quantity", gorm.Expr("quantity - ?", orderItem.Quantity))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("%w for item: %s", errInsufficientStock, item.Name)
			}

			calculatedTotalPrice += float64(orderItem.Quantity) * item.Price

			orderItemsToCreate = append(orderItemsToCreate, models.OrderItem{
				ItemID:   item.ID,
				Quantity: orderItem.Quantity,
				Price:    item.Price,
			})
		}

		order := models.Order{
			CustomerID: body.CustomerID,
			TotalPrice: calculatedTotalPrice,
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		for i := range orderItemsToCreate {
			orderItemsToCreate[i].OrderID = order.ID
			if err := tx.Create(&orderItemsToCreate[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		switch {
		case errors.Is(err, errCustomerNotFound), errors.Is(err, errItemNotFound):
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		case errors.Is(err, errInvalidOrder):
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		case errors.Is(err, errInsufficientStock):
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not create order"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Order created successfully",
		"total_price": calculatedTotalPrice,
	})
}

func GetOrdersLatest(c *fiber.Ctx) error {
	currentTime := time.Now()
	last24Hours := currentTime.Add(-24 * time.Hour)

	var orders []models.Order
	if err := database.DB.Preload("Items").Where("created_at >= ?", last24Hours.Unix()).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve orders"})
	}

	return c.JSON(orders)
}

func GetOrdersByDate(c *fiber.Ctx) error {
	var orders []models.Order
	date := c.Query("date")
	if date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Date is required"})
	}

	location, err := time.LoadLocation("Asia/Colombo")
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not load location"})
	}

	parsedDate, err := time.ParseInLocation("2006-01-02", date, location)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid date format. Use YYYY-MM-DD"})
	}

	startOfDay := parsedDate.Unix()
	endOfDay := parsedDate.Add(24 * time.Hour).Unix()

	if err := database.DB.Preload("Items").Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve orders"})
	}

	return c.JSON(orders)
}

func GetOrdersByCustomer(c *fiber.Ctx) error {
	customerID := c.Params("customer_id")
	if customerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CustomerID is required"})
	}

	customerIDUint, err := strconv.ParseUint(customerID, 10, 64)
	if err != nil || customerIDUint == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid CustomerID format"})
	}

	var orders []models.Order
	if err := database.DB.Preload("Items").Where("customer_id = ?", customerIDUint).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve orders"})
	}

	return c.JSON(orders)
}

func GetOrdersPaginated(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)

	if page < 1 || limit < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "page and limit must be positive integers"})
	}

	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	var orders []models.Order
	if err := database.DB.Preload("Items").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not retrieve orders"})
	}

	return c.JSON(orders)
}
