package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"main/src/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func seedTestCategory(t *testing.T, testDB *gorm.DB) models.Category {
	t.Helper()

	category := models.Category{Name: "Ice Cream"}
	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed test category: %v", err)
	}

	return category
}

func TestGetItems(t *testing.T) {
	testDB := setupTestDB(t)
	category := seedTestCategory(t, testDB)

	item := models.Item{
		Name:       "Chocolate Ice Cream",
		Price:      6.99,
		Quantity:   10,
		CategoryID: category.ID,
	}

	if err := testDB.Create(&item).Error; err != nil {
		t.Fatalf("failed to seed test item: %v", err)
	}

	app := fiber.New()
	app.Get("/items", GetItems)

	req := httptest.NewRequest(http.MethodGet, "/items", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var items []models.Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}

	if items[0].Name != item.Name {
		t.Errorf("expected item name %q, got %q", item.Name, items[0].Name)
	}
}

func TestGetItemById(t *testing.T) {
	testDB := setupTestDB(t)
	category := seedTestCategory(t, testDB)

	item := models.Item{
		Name:       "Chocolate Ice Cream",
		Price:      6.99,
		Quantity:   10,
		CategoryID: category.ID,
	}

	if err := testDB.Create(&item).Error; err != nil {
		t.Fatalf("failed to seed test item: %v", err)
	}

	app := fiber.New()
	app.Get("/items/:id", GetItemById)

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/items/%d", item.ID),
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

	var responseItem models.Item
	if err := json.NewDecoder(resp.Body).Decode(&responseItem); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseItem.ID != item.ID {
		t.Errorf("expected item ID %d, got %d", item.ID, responseItem.ID)
	}

	if responseItem.Name != item.Name {
		t.Errorf("expected item name %q, got %q", item.Name, responseItem.Name)
	}
}

func TestCreateItem(t *testing.T) {
	testDB := setupTestDB(t)
	category := seedTestCategory(t, testDB)

	app := fiber.New()
	app.Post("/items", CreateItem)

	req := httptest.NewRequest(
		http.MethodPost,
		"/items",
		bytes.NewBufferString(fmt.Sprintf(
			`{"name":"Strawberry Ice Cream","description":"Fresh strawberry flavor","price":7.99,"quantity":20,"category_id":%d}`,
			category.ID,
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

	var item models.Item
	if err := testDB.First(&item, "name = ?", "Strawberry Ice Cream").Error; err != nil {
		t.Fatalf("failed to find created item: %v", err)
	}

	if item.Price != 7.99 {
		t.Errorf("expected price %.2f, got %.2f", 7.99, item.Price)
	}

	if item.Quantity != 20 {
		t.Errorf("expected quantity %d, got %d", 20, item.Quantity)
	}

	if item.CategoryID != category.ID {
		t.Errorf("expected category ID %d, got %d", category.ID, item.CategoryID)
	}
}

func TestUpdateItem(t *testing.T) {
	testDB := setupTestDB(t)
	category := seedTestCategory(t, testDB)

	item := models.Item{
		Name:       "Chocolate Ice Cream",
		Price:      6.99,
		Quantity:   10,
		CategoryID: category.ID,
	}

	if err := testDB.Create(&item).Error; err != nil {
		t.Fatalf("failed to seed test item: %v", err)
	}

	app := fiber.New()
	app.Put("/items/:id", UpdateItem)

	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/items/%d", item.ID),
		bytes.NewBufferString(`{"name":"Mint Chocolate Chip Ice Cream","price":6.99,"quantity":18}`),
	)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var updatedItem models.Item
	if err := testDB.First(&updatedItem, item.ID).Error; err != nil {
		t.Fatalf("failed to find updated item: %v", err)
	}

	if updatedItem.Name != "Mint Chocolate Chip Ice Cream" {
		t.Errorf("expected name %q, got %q", "Mint Chocolate Chip Ice Cream", updatedItem.Name)
	}

	if updatedItem.Price != 6.99 {
		t.Errorf("expected price %.2f, got %.2f", 6.99, updatedItem.Price)
	}

	if updatedItem.Quantity != 18 {
		t.Errorf("expected quantity %d, got %d", 18, updatedItem.Quantity)
	}
}

func TestDeleteItem(t *testing.T) {
	testDB := setupTestDB(t)
	category := seedTestCategory(t, testDB)

	item := models.Item{
		Name:       "Chocolate Ice Cream",
		Price:      6.99,
		Quantity:   10,
		CategoryID: category.ID,
	}

	if err := testDB.Create(&item).Error; err != nil {
		t.Fatalf("failed to seed test item: %v", err)
	}

	app := fiber.New()
	app.Delete("/items/:id", DeleteItem)

	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/items/%d", item.ID),
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

	var deletedItem models.Item
	err = testDB.First(&deletedItem, item.ID).Error
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected item to be deleted, but it still exists")
	}
}

func TestGetItemByIdNotFound(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Get("/items/:id", GetItemById)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/items/9999", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestCreateItemValidation(t *testing.T) {
	testDB := setupTestDB(t)
	category := seedTestCategory(t, testDB)
	app := fiber.New()
	app.Post("/items", CreateItem)

	tests := []string{
		`{`,
		`{"name":"Ice Cream"}`,
		fmt.Sprintf(`{"name":"Ice Cream","price":0,"quantity":1,"category_id":%d}`, category.ID),
		fmt.Sprintf(`{"name":"Ice Cream","price":1,"quantity":-1,"category_id":%d}`, category.ID),
	}
	for _, body := range tests {
		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewBufferString(body))
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

func TestUpdateItemValidation(t *testing.T) {
	testDB := setupTestDB(t)
	category := seedTestCategory(t, testDB)
	item := models.Item{Name: "Vanilla", Price: 5, Quantity: 10, CategoryID: category.ID}
	if err := testDB.Create(&item).Error; err != nil {
		t.Fatalf("failed to seed item: %v", err)
	}

	app := fiber.New()
	app.Put("/items/:id", UpdateItem)
	tests := []string{`{`, `{}`, `{"name":""}`, `{"price":-1}`, `{"quantity":-1}`}
	for _, body := range tests {
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/items/%d", item.ID), bytes.NewBufferString(body))
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

func TestUpdateItemNotFound(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Put("/items/:id", UpdateItem)

	req := httptest.NewRequest(http.MethodPut, "/items/9999", bytes.NewBufferString(`{"name":"Vanilla"}`))
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

func TestDeleteItemNotFound(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Delete("/items/:id", DeleteItem)

	resp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/items/9999", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}
