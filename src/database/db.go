package database

import (
	"os"

	"github.com/glebarez/sqlite"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
	"main/src/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		panic("failed to load environment variables: " + err.Error())
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		panic("DB_NAME environment variable is required")
	}

	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	if err := db.AutoMigrate(&models.Customer{}, &models.Item{}, &models.Category{}, &models.User{}); err != nil {
		panic("failed to migrate database: " + err.Error())
	}

	DB = db
}
