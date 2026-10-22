package models

type OrderItem struct{
	ID uint `gorm:"primaryKey;autoIncrement"`
	OrderID uint `gorm:"not null"`
	ItemID uint `gorm:"not null"`
	Quantity int `gorm:"not null"`
	Price float64 `gorm:"not null"`
}

type Order struct {
	ID         uint     `gorm:"primaryKey;autoIncrement"`
	CustomerID uint     `gorm:"not null"`
	TotalPrice float64  `gorm:"not null"`
	Items      []OrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt  int64    `gorm:"autoCreateTime"`
	UpdatedAt  int64    `gorm:"autoUpdateTime"`
}
