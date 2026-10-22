package handlers

import (
	"fmt"
	"testing"

	"main/src/database"
	"main/src/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	testDB, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	if err:= testDB.AutoMigrate(
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
