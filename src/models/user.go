package models

import (
	_ "gorm.io/gorm"
)

var userRole = map[string]string{
	"owner":   "owner",
	"manager": "manager",
	"cashier": "cashier",
}

const (
	RoleOwner   = "owner"
	RoleManager = "manager"
	RoleCashier = "cashier"
)

type User struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null" json:"-"`
	Role      string `gorm:"not null"`
	CreatedAt int64  `gorm:"autoCreateTime"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}
