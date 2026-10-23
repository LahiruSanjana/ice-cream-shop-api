package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"main/src/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func seedOrderData(t *testing.T) (*gorm.DB, models.Customer, models.Item) {
	t.Helper()

	testDB := setupTestDB(t)
	customer := models.Customer{
		Name:    "Order Customer",
		Email:   "order.customer@example.com",
		Phone:   "1234567890",
		Address: "Main Street",
	}
	if err := testDB.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed test customer: %v", err)
	}

	category := models.Category{Name: "Ice Cream"}
	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed test category: %v", err)
	}

	item := models.Item{
		Name:       "Chocolate Ice Cream",
		Price:      5.50,
		Quantity:   10,
		CategoryID: category.ID,
	}
	if err := testDB.Create(&item).Error; err != nil {
		t.Fatalf("failed to seed test item: %v", err)
	}

	return testDB, customer, item
}

func TestCreateOrder(t *testing.T) {
	testDB, customer, item := seedOrderData(t)

	app := fiber.New()
	app.Post("/orders", CreateOrder)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(fmt.Sprintf(
			`{"customer_id":%d,"items":[{"item_id":%d,"quantity":2}]}`,
			customer.ID,
			item.ID,
		)),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var response struct {
		TotalPrice float64 `json:"total_price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.TotalPrice != 11.00 {
		t.Errorf("expected total price %.2f, got %.2f", 11.00, response.TotalPrice)
	}

	var updatedItem models.Item
	if err := testDB.First(&updatedItem, item.ID).Error; err != nil {
		t.Fatalf("failed to find updated item: %v", err)
	}
	if updatedItem.Quantity != 8 {
		t.Errorf("expected stock quantity %d, got %d", 8, updatedItem.Quantity)
	}

	var order models.Order
	if err := testDB.Preload("Items").First(&order).Error; err != nil {
		t.Fatalf("failed to find created order: %v", err)
	}
	if order.CustomerID != customer.ID {
		t.Errorf("expected customer ID %d, got %d", customer.ID, order.CustomerID)
	}
	if len(order.Items) != 1 || order.Items[0].Quantity != 2 {
		t.Fatalf("expected one order item with quantity 2, got %+v", order.Items)
	}
}

func TestCreateOrderInsufficientStockRollsBack(t *testing.T) {
	testDB, customer, item := seedOrderData(t)
	app := fiber.New()
	app.Post("/orders", CreateOrder)

	req := httptest.NewRequest(
		http.MethodPost,
		"/orders",
		bytes.NewBufferString(fmt.Sprintf(
			`{"customer_id":%d,"items":[{"item_id":%d,"quantity":11}]}`,
			customer.ID,
			item.ID,
		)),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected status code %d, got %d", http.StatusConflict, resp.StatusCode)
	}

	var updatedItem models.Item
	if err := testDB.First(&updatedItem, item.ID).Error; err != nil {
		t.Fatalf("failed to find item: %v", err)
	}
	if updatedItem.Quantity != item.Quantity {
		t.Errorf("expected stock to remain %d, got %d", item.Quantity, updatedItem.Quantity)
	}

	var orderCount int64
	if err := testDB.Model(&models.Order{}).Count(&orderCount).Error; err != nil {
		t.Fatalf("failed to count orders: %v", err)
	}
	if orderCount != 0 {
		t.Errorf("expected no order to be created, got %d", orderCount)
	}
}

func TestGetOrdersByCustomer(t *testing.T) {
	testDB, customer, item := seedOrderData(t)

	order := models.Order{
		CustomerID: customer.ID,
		TotalPrice: item.Price * 2,
		Items: []models.OrderItem{
			{ItemID: item.ID, Quantity: 2, Price: item.Price},
		},
	}

	if err := testDB.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed test order: %v", err)
	}

	app := fiber.New()
	app.Get("/orders/customer/:customer_id", GetOrdersByCustomer)

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/orders/customer/%d", customer.ID),
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var orders []models.Order
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(orders))
	}
	if orders[0].CustomerID != customer.ID {
		t.Errorf("expected customer ID %d, got %d", customer.ID, orders[0].CustomerID)
	}
	if len(orders[0].Items) != 1 {
		t.Errorf("expected 1 preloaded order item, got %d", len(orders[0].Items))
	}
}

func TestCreateOrderValidation(t *testing.T) {
	_, customer, item := seedOrderData(t)
	app := fiber.New()
	app.Post("/orders", CreateOrder)

	tests := []string{
		`{`,
		`{"customer_id":1,"items":[]}`,
		fmt.Sprintf(`{"items":[{"item_id":%d,"quantity":1}]}`, item.ID),
		fmt.Sprintf(`{"customer_id":%d,"items":[{"item_id":0,"quantity":1}]}`, customer.ID),
		fmt.Sprintf(`{"customer_id":%d,"items":[{"item_id":%d,"quantity":0}]}`, customer.ID, item.ID),
	}
	for _, body := range tests {
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("body %q: expected status code %d, got %d", body, http.StatusBadRequest, resp.StatusCode)
		}
	}
}

func TestCreateOrderNotFound(t *testing.T) {
	_, _, item := seedOrderData(t)
	app := fiber.New()
	app.Post("/orders", CreateOrder)

	req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBufferString(
		fmt.Sprintf(`{"customer_id":9999,"items":[{"item_id":%d,"quantity":1}]}`, item.ID),
	))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestGetOrdersByCustomerInvalidID(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Get("/orders/customer/:customer_id", GetOrdersByCustomer)

	for _, id := range []string{"abc", "0"} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/orders/customer/"+id, nil))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("id %q: expected status code %d, got %d", id, http.StatusBadRequest, resp.StatusCode)
		}
	}
}

func TestGetOrdersByDateValidation(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Get("/orders/date", GetOrdersByDate)

	for _, path := range []string{"/orders/date", "/orders/date?date=10-10-2026"} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("path %q: expected status code %d, got %d", path, http.StatusBadRequest, resp.StatusCode)
		}
	}
}

func TestGetOrdersPaginatedValidation(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Get("/orders", GetOrdersPaginated)

	for _, path := range []string{"/orders?page=0", "/orders?limit=0", "/orders?page=-1", "/orders?limit=-1"} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("path %q: expected status code %d, got %d", path, http.StatusBadRequest, resp.StatusCode)
		}
	}
}
