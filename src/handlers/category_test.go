package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"fmt"

	"main/src/database"
	"main/src/models"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func TestGetCategories(t *testing.T) {
	testDB := setupTestDB(t)

	category := models.Category{
		Name: "Vanilla",
	}

	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed test category: %v", err)
	}

	app := fiber.New()
	app.Get("/categories", GetCategories)

	req := httptest.NewRequest("GET", "/categories", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetCategoryById(t *testing.T) {
	testDB := setupTestDB(t)

	category := models.Category{
		Name: "Vanilla",
	}

	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed test category: %v", err)
	}

	app := fiber.New()
	app.Get("/categories/:id", GetCategoryById)

	req := httptest.NewRequest("GET", fmt.Sprintf("/categories/%d", category.ID), nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetCategoryByIdNotFound(t *testing.T) {
	setupTestDB(t)

	app := fiber.New()
	app.Get("/categories/:id", GetCategoryById)

	req := httptest.NewRequest(http.MethodGet, "/categories/9999", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, but got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestCreateCategory(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := testDB.AutoMigrate(&models.Category{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	previousDB := database.DB
	database.DB = testDB
	t.Cleanup(func() {
		database.DB = previousDB
	})

	app := fiber.New()
	app.Post("/categories", CreateCategory)

	jsonBody := []byte(`{"name": "Strawberry"}`)
	req := httptest.NewRequest("POST", "/categories", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status code %d, but got %d", http.StatusCreated, resp.StatusCode)
	}
}

func TestCreateCategoryInvalidJSON(t *testing.T) {
	setupTestDB(t)

	app := fiber.New()
	app.Post("/categories", CreateCategory)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{"name":`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code %d, but got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestCreateCategoryNameRequired(t *testing.T) {
	setupTestDB(t)

	app := fiber.New()
	app.Post("/categories", CreateCategory)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code %d, but got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestUpdateCategory(t *testing.T) {
	testDB := setupTestDB(t)

	category := models.Category{
		Name: "Vanilla",
	}

	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed test category: %v", err)
	}

	app := fiber.New()
	app.Put("/categories/:id", UpdateCategory)

	jsonBody := []byte(`{"name": "Peppermint"}`)
	req := httptest.NewRequest("PUT", fmt.Sprintf("/categories/%d", category.ID), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestUpdateCategoryNotFound(t *testing.T) {
	setupTestDB(t)

	app := fiber.New()
	app.Put("/categories/:id", UpdateCategory)

	req := httptest.NewRequest(http.MethodPut, "/categories/9999", bytes.NewBufferString(`{"name":"Peppermint"}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, but got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestUpdateCategoryInvalidJSON(t *testing.T) {
	testDB := setupTestDB(t)
	category := models.Category{Name: "Vanilla"}
	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed test category: %v", err)
	}

	app := fiber.New()
	app.Put("/categories/:id", UpdateCategory)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/categories/%d", category.ID), bytes.NewBufferString(`{"name":`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code %d, but got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestUpdateCategoryNameCannotBeEmpty(t *testing.T) {
	testDB := setupTestDB(t)
	category := models.Category{Name: "Vanilla"}
	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed test category: %v", err)
	}

	app := fiber.New()
	app.Put("/categories/:id", UpdateCategory)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/categories/%d", category.ID), bytes.NewBufferString(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code %d, but got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestDeleteCategory(t *testing.T) {
	testDB := setupTestDB(t)

	category := models.Category{
		Name: "Vanilla",
	}

	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed test category: %v", err)
	}

	app := fiber.New()
	app.Delete("/categories/:id", DeleteCategory)

	req := httptest.NewRequest("DELETE", fmt.Sprintf("/categories/%d", category.ID), nil)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDeleteCategoryNotFound(t *testing.T) {
	setupTestDB(t)

	app := fiber.New()
	app.Delete("/categories/:id", DeleteCategory)

	req := httptest.NewRequest(http.MethodDelete, "/categories/9999", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, but got %d", http.StatusNotFound, resp.StatusCode)
	}
}
