// internal/domain/entities/pharmacy_inventory.go
package entities

import "time"

type PharmacyInventory struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PharmacyID  string    `gorm:"column:pharmacy_id;type:uuid;not null" json:"pharmacy_id"`
	ProductID   string    `gorm:"column:product_id;type:uuid;not null" json:"product_id"`
	Stock       int       `gorm:"column:stock;not null;default:0" json:"stock"`
	Price       float64   `gorm:"column:price;type:decimal(10,2);not null" json:"price"`
	IsAvailable bool      `gorm:"column:is_available;not null;default:true" json:"is_available"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

func (PharmacyInventory) TableName() string {
	return "pharmacy.pharmacy_inventory"
}
