package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"main/src/models"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func TestRegisterUser(t *testing.T) {
	testDB := setupTestDB(t)

	app := fiber.New()
	app.Post("/register", RegisterUser)

	req := httptest.NewRequest(
		http.MethodPost,
		"/register",
		bytes.NewBufferString(`{"name":"Jane Doe","email":"jane@example.com","password":"secret123"}`),
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

	var user models.User
	if err := testDB.Where("email = ?", "jane@example.com").First(&user).Error; err != nil {
		t.Fatalf("failed to find registered user: %v", err)
	}

	if user.Role != models.RoleCashier {
		t.Errorf("expected default role %q, got %q", models.RoleCashier, user.Role)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("secret123")); err != nil {
		t.Errorf("expected password to be hashed correctly: %v", err)
	}
}

func TestLoginUser(t *testing.T) {
	testDB := setupTestDB(t)
	t.Setenv("JWT_SECRET", "test-secret")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	user := models.User{
		Name:     "Jane Doe",
		Email:    "jane@example.com",
		Password: string(passwordHash),
		Role:     models.RoleManager,
	}
	if err := testDB.Create(&user).Error; err != nil {
		t.Fatalf("failed to seed test user: %v", err)
	}

	app := fiber.New()
	app.Post("/login", LoginUser)

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBufferString(`{"email":"jane@example.com","password":"secret123"}`),
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

	var response struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if response.Message != "Login successful" {
		t.Errorf("expected successful login message, got %q", response.Message)
	}

	cookie := resp.Header.Get("Set-Cookie")
	if !strings.Contains(cookie, "jwt=") {
		t.Fatalf("expected jwt cookie, got %q", cookie)
	}
	if !strings.Contains(cookie, "HttpOnly") {
		t.Errorf("expected HttpOnly cookie, got %q", cookie)
	}

	tokenString := strings.Split(strings.Split(cookie, "jwt=")[1], ";")[0]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret"), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		t.Fatalf("expected a valid JWT cookie: %v", err)
	}
}

func TestLoginUserWithInvalidPassword(t *testing.T) {
	testDB := setupTestDB(t)
	t.Setenv("JWT_SECRET", "test-secret")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash test password: %v", err)
	}

	if err := testDB.Create(&models.User{
		Name:     "Jane Doe",
		Email:    "jane@example.com",
		Password: string(passwordHash),
		Role:     models.RoleCashier,
	}).Error; err != nil {
		t.Fatalf("failed to seed test user: %v", err)
	}

	app := fiber.New()
	app.Post("/login", LoginUser)

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBufferString(`{"email":"jane@example.com","password":"wrong-password"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, resp.StatusCode)
	}
}

func TestLoginUserInvalidEmailAndValidation(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Post("/login", LoginUser)

	tests := []struct {
		body           string
		expectedStatus int
	}{
		{body: `{`, expectedStatus: http.StatusBadRequest},
		{body: `{}`, expectedStatus: http.StatusBadRequest},
		{
			body:           `{"email":"missing@example.com","password":"secret"}`,
			expectedStatus: http.StatusUnauthorized,
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(test.body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != test.expectedStatus {
			t.Fatalf("body %q: expected status code %d, got %d", test.body, test.expectedStatus, resp.StatusCode)
		}
	}
}

func TestLogoutUser(t *testing.T) {
	app := fiber.New()
	app.Post("/logout", LogoutUser)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Cookie", "jwt=test-token")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("failed to test app: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	cookie := resp.Header.Get("Set-Cookie")
	if !strings.Contains(cookie, "jwt=") || !strings.Contains(cookie, "expires=") {
		t.Errorf("expected jwt cookie to be cleared, got %q", cookie)
	}
}

func TestRegisterUserValidation(t *testing.T) {
	setupTestDB(t)
	app := fiber.New()
	app.Post("/register", RegisterUser)

	for _, body := range []string{`{`, `{}`, `{"name":"Jane","email":"jane@example.com","password":"x","role":"invalid"}`} {
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
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

func TestRegisterUserDuplicateEmail(t *testing.T) {
	testDB := setupTestDB(t)
	if err := testDB.Create(&models.User{Name: "Existing", Email: "existing@example.com", Password: "hash", Role: models.RoleCashier}).Error; err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	app := fiber.New()
	app.Post("/register", RegisterUser)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(
		`{"name":"Jane","email":"existing@example.com","password":"secret"}`,
	))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}
