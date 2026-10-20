package database

import (
	"log"
	"main/src/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) {
	// 1. Check if users already exist (Idempotency check)
	var count int64
	if err := db.Model(&models.User{}).Count(&count).Error; err != nil {
		log.Printf("Failed to check existing users: %v\n", err)
		return
	}
	if count > 0 {
		log.Println("Database already seeded with users. Skipping seeding.")
		return
	}

	// 2. Helper function to hash passwords securely
	hashPassword := func(password string) string {
		bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Failed to hash password: ", err)
		}
		return string(bytes)
	}

	// 3. Define initial users according to the assessment roles
	users := []models.User{
		{
			Name:     "Shop Owner",
			Email:    "owner@icecream.com",
			Password: hashPassword("owner123"),
			Role:     models.RoleOwner,
		},
		{
			Name:     "Shop Manager",
			Email:    "manager@icecream.com",
			Password: hashPassword("manager123"),
			Role:     models.RoleManager,
		},
		{
			Name:     "Shop Cashier",
			Email:    "cashier@icecream.com",
			Password: hashPassword("cashier123"),
			Role:     models.RoleCashier,
		},
	}

	// 4. Insert users into the database
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to seed user %s: %v\n", user.Email, err)
		} else {
			log.Printf("Seeded user: %s (%s)\n", user.Email, user.Role)
		}
	}

	log.Println("Database seeding completed successfully!")
}
