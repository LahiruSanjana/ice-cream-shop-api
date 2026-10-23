package routes

import (
	"bytes"
	"errors"
	"fmt"
	"main/src/database"
	"main/src/models"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const testJWTSecret = "test-secret"

func testJWT(t *testing.T, role string) string {
	t.Helper()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": uint(1),
		"role":    role,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}

	return tokenString
}

func setupRouteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	testDB, err := gorm.Open(
		sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())),
		&gorm.Config{},
	)
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err := testDB.AutoMigrate(
		&models.Customer{},
		&models.Item{},
		&models.Category{},
		&models.User{},
		&models.Order{},
		&models.OrderItem{},
	); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	previousDB := database.DB
	database.DB = testDB
	t.Cleanup(func() {
		database.DB = previousDB
	})

	return testDB
}

func TestProtectedItemRouteRequiresJWT(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	app := fiber.New()
	api := app.Group("/api")
	ItemRoutes(api)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/items",
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestCashierCannotCreateItem(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	app := fiber.New()
	api := app.Group("/api")
	ItemRoutes(api)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/items",
		bytes.NewBufferString(`{"name":"Chocolate","price":5.5,"quantity":10,"category_id":1}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: testJWT(t, "cashier"),
	})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status code %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
}

func TestManagerCannotDeleteItem(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	app := fiber.New()
	api := app.Group("/api")
	ItemRoutes(api)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/items/1",
		nil,
	)
	req.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: testJWT(t, "manager"),
	})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status code %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
}

func TestManagerCannotDeleteCustomer(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	app := fiber.New()
	api := app.Group("/api")
	CustomerRoutes(api)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/customers/1",
		nil,
	)
	req.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: testJWT(t, "manager"),
	})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status code %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
}

func TestCashierCannotDeleteCustomer(t *testing.T) {
	t.Setenv("JWT_SECRET", testJWTSecret)

	app := fiber.New()
	api := app.Group("/api")
	CustomerRoutes(api)

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/customers/1",
		nil,
	)
	req.AddCookie(&http.Cookie{
		Name:  "jwt",
		Value: testJWT(t, "cashier"),
	})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status code %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
}

func TestOwnerCanDeleteItem(t *testing.T) {
	testDB := setupRouteTestDB(t)
	category := models.Category{Name: "Ice Cream"}
	if err := testDB.Create(&category).Error; err != nil {
		t.Fatalf("failed to seed category: %v", err)
	}

	item := models.Item{
		Name:       "Chocolate Ice Cream",
		Price:      5.50,
		Quantity:   10,
		CategoryID: category.ID,
	}
	if err := testDB.Create(&item).Error; err != nil {
		t.Fatalf("failed to seed item: %v", err)
	}

	t.Setenv("JWT_SECRET", testJWTSecret)
	app := fiber.New()
	api := app.Group("/api")
	ItemRoutes(api)

	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/api/items/%d", item.ID),
		nil,
	)
	req.AddCookie(&http.Cookie{Name: "jwt", Value: testJWT(t, "owner")})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var deletedItem models.Item
	err = testDB.First(&deletedItem, item.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected item to be deleted, got error: %v", err)
	}
}

func TestOwnerCanDeleteCustomer(t *testing.T) {
	testDB := setupRouteTestDB(t)
	customer := models.Customer{
		Name:    "Jane Doe",
		Email:   "jane@example.com",
		Phone:   "0771234567",
		Address: "Main Street",
	}
	if err := testDB.Create(&customer).Error; err != nil {
		t.Fatalf("failed to seed customer: %v", err)
	}

	t.Setenv("JWT_SECRET", testJWTSecret)
	app := fiber.New()
	api := app.Group("/api")
	CustomerRoutes(api)

	req := httptest.NewRequest(
		http.MethodDelete,
		fmt.Sprintf("/api/customers/%d", customer.ID),
		nil,
	)
	req.AddCookie(&http.Cookie{Name: "jwt", Value: testJWT(t, "owner")})

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	var deletedCustomer models.Customer
	err = testDB.First(&deletedCustomer, customer.ID).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected customer to be deleted, got error: %v", err)
	}
}
