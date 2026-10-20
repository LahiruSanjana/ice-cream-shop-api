package database

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"main/src/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	db, err := gorm.Open(sqlite.Open("icecream.db"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database"+ err.Error())
	}

	if err := db.AutoMigrate(&models.Customer{}, &models.Item{}, &models.Category{}, &models.User{}); err != nil {
		panic("Failed to migrate database")
	}

	DB = db
}
