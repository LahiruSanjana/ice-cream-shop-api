package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"main/src/database"
	"main/src/models"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func TestGetAllCustomers(t *testing.T) {
	testDB := setupTestDB(t)

	customer := models.Customer{
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Phone:   "1234567890",
		Address: "123 Main St",
	}

	if err := testDB.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed test customer: %v", err)
	}

	app := fiber.New()
	app.Get("/customers", GetAllCustomers)

	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var response struct {
		Customers []models.Customer `json:"customers"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(response.Customers) != 1 {
		t.Fatalf("expected 1 customer, got %d", len(response.Customers))
	}

	if response.Customers[0].Name != customer.Name {
		t.Errorf("expected customer name %q, got %q", customer.Name, response.Customers[0].Name)
	}

	if response.Customers[0].Email != customer.Email {
		t.Errorf("expected customer email %q, got %q", customer.Email, response.Customers[0].Email)
	}

	if response.Customers[0].Phone != customer.Phone {
		t.Errorf("expected customer phone %q, got %q", customer.Phone, response.Customers[0].Phone)
	}

	if response.Customers[0].Address != customer.Address {
		t.Errorf("expected customer address %q, got %q", customer.Address, response.Customers[0].Address)
	}
}

func TestGetCustomerById(t *testing.T) {
	testDB := setupTestDB(t)

	customer := models.Customer{
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Phone:   "1234567890",
		Address: "123 Main St",
	}

	if err := testDB.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed test customer: %v", err)
	}

	app := fiber.New()
	app.Get("/customers/:id", GetCustomerById)

	req := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("/customers/%d", customer.ID),
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

	var response struct {
		Customer models.Customer `json:"customer"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Customer.ID != customer.ID {
		t.Errorf("expected customer ID %d, got %d", customer.ID, response.Customer.ID)
	}

	if response.Customer.Name != customer.Name {
		t.Errorf("expected customer name %q, got %q", customer.Name, response.Customer.Name)
	}

	if response.Customer.Email != customer.Email {
		t.Errorf("expected customer email %q, got %q", customer.Email, response.Customer.Email)
	}

	if response.Customer.Phone != customer.Phone {
		t.Errorf("expected customer phone %q, got %q", customer.Phone, response.Customer.Phone)
	}

	if response.Customer.Address != customer.Address {
		t.Errorf("expected customer address %q, got %q", customer.Address, response.Customer.Address)
	}
}

func TestCreateCustomer(t *testing.T) {
	testDB, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := testDB.AutoMigrate(&models.Customer{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	previousDB := database.DB
	database.DB = testDB
	t.Cleanup(func() {
		database.DB = previousDB
	})

	app := fiber.New()
	app.Post("/customers", CreateCustomer)

	req := httptest.NewRequest(
		http.MethodPost,
		"/customers",
		bytes.NewBufferString(
			`{"name":"Jane Doe","email":"jane.doe@example.com","phone":"0987654321","address":"456 Oak Ave"}`,
		),
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

	var customer models.Customer
	if err := testDB.First(&customer, "name = ?", "Jane Doe").Error; err != nil {
		t.Fatalf("failed to find created customer: %v", err)
	}

	if customer.Name != "Jane Doe" {
		t.Errorf("expected customer name %q, got %q", "Jane Doe", customer.Name)
	}
}

func TestUpdateCustomer(t *testing.T) {
	testDB := setupTestDB(t)

	customer := models.Customer{
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Phone:   "1234567890",
		Address: "123 Main St",
	}

	if err := testDB.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed test customer: %v", err)
	}

	app := fiber.New()
	app.Put("/customers/:id", UpdateCustomer)

	req := httptest.NewRequest(
		http.MethodPut,
		fmt.Sprintf("/customers/%d", customer.ID),
		bytes.NewBufferString(`{"name":"Jane Doe","email":"jane.doe@example.com","phone":"0987654321","address":"456 Oak Ave"}`),
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

	var updatedCustomer models.Customer
	if err := testDB.First(&updatedCustomer, customer.ID).Error; err != nil {
		t.Fatalf("failed to find updated customer: %v", err)
	}

	if updatedCustomer.Name != "Jane Doe" {
		t.Errorf("expected name %q, got %q", "Jane Doe", updatedCustomer.Name)
	}

	if updatedCustomer.Email != "jane.doe@example.com" {
		t.Errorf("expected email %q, got %q", "jane.doe@example.com", updatedCustomer.Email)
	}

	if updatedCustomer.Phone != "0987654321" {
		t.Errorf("expected phone %q, got %q", "0987654321", updatedCustomer.Phone)
	}

	if updatedCustomer.Address != "456 Oak Ave" {
		t.Errorf("expected address %q, got %q", "456 Oak Ave", updatedCustomer.Address)
	}
}

func TestDeleteCustomer(t *testing.T) {
	testDB := setupTestDB(t)

	customer := models.Customer{
		Name:    "John Doe",
		Email:   "john.doe@example.com",
		Phone:   "1234567890",
		Address: "123 Main St",
	}

	if err := testDB.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed test customer: %v", err)
	}

	app := fiber.New()
	app.Delete("/customers/:id", DeleteCustomer)

	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/customers/%d", customer.ID),
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

	var deletedCustomer models.Customer
	err = testDB.First(&deletedCustomer, customer.ID).Error
	if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected customer to be deleted, but it still exists")
	}
}

func TestGetCustomerByIdNotFound(t *testing.T) {
	setupTestDB(t)

	app := fiber.New()
	app.Get("/customers/:id", GetCustomerById)

	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/customers/9999", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

func TestCreateCustomerInvalidRequest(t *testing.T) {
	setupTestDB(t)

	app := fiber.New()
	app.Post("/customers", CreateCustomer)

	tests := []string{`{"name":`, `{"name":"Jane"}`}
	for _, body := range tests {
		resp, err := app.Test(func() *http.Request {
			req := httptest.NewRequest(http.MethodPost, "/customers", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			return req
		}())
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	}
}

func TestUpdateCustomerValidation(t *testing.T) {
	testDB := setupTestDB(t)
	customer := models.Customer{Name: "John Doe", Email: "john@example.com", Phone: "1234567890"}
	if err := testDB.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed customer: %v", err)
	}

	app := fiber.New()
	app.Put("/customers/:id", UpdateCustomer)

	tests := []string{`{`, `{}`, `{"name":""}`, `{"email":""}`, `{"phone":""}`}
	for _, body := range tests {
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/customers/%d", customer.ID), bytes.NewBufferString(body))
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

func TestUpdateCustomerNotFound(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Put("/customers/:id", UpdateCustomer)

	req := httptest.NewRequest(http.MethodPut, "/customers/9999", bytes.NewBufferString(`{"name":"Jane"}`))
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

func TestDeleteCustomerNotFound(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Delete("/customers/:id", DeleteCustomer)

	resp, err := app.Test(httptest.NewRequest(http.MethodDelete, "/customers/9999", nil))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected status code %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}
