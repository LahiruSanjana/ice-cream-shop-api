package models

type Customer struct {
	ID      uint   `gorm:"primaryKey;autoIncrement"`
	Name    string `gorm:"not null"`
	Email   string `gorm:"unique;not null"`
	Phone   string `gorm:"not null"`
	Address string

	CreatedAt int64 `gorm:"autoCreateTime"`
	UpdatedAt int64 `gorm:"autoUpdateTime"`
}
