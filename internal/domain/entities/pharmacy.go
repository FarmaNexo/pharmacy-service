// internal/domain/entities/pharmacy.go
package entities

import "time"

type Pharmacy struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	OwnerUserID string     `gorm:"column:owner_user_id;type:uuid;not null" json:"owner_user_id"`
	Name        string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Slug        string     `gorm:"column:slug;type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description string     `gorm:"column:description;type:text" json:"description"`
	Phone       string     `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Email       string     `gorm:"column:email;type:varchar(255)" json:"email"`
	Website     string     `gorm:"column:website;type:varchar(500)" json:"website"`
	LogoURL                 string `gorm:"column:logo_url;type:varchar(500)" json:"logo_url"`
	AuthorizationDocumentURL string `gorm:"column:authorization_document_url;type:varchar(500)" json:"authorization_document_url,omitempty"`

	// Chain
	ChainID   string `gorm:"column:chain_id;type:varchar(100)" json:"chain_id,omitempty"`
	ChainName string `gorm:"column:chain_name;type:varchar(255)" json:"chain_name,omitempty"`

	// Address
	Street     string `gorm:"column:street;type:varchar(255);not null" json:"street"`
	City       string `gorm:"column:city;type:varchar(100);not null" json:"city"`
	State      string `gorm:"column:state;type:varchar(100)" json:"state"`
	PostalCode string `gorm:"column:postal_code;type:varchar(20)" json:"postal_code"`
	Country    string `gorm:"column:country;type:varchar(100);not null;default:'Perú'" json:"country"`

	// Geolocation - stored as PostGIS GEOGRAPHY but we use lat/lng floats in Go
	Latitude  float64 `gorm:"-" json:"latitude"`
	Longitude float64 `gorm:"-" json:"longitude"`

	// State
	IsVerified bool `gorm:"column:is_verified;not null;default:false" json:"is_verified"`
	IsActive   bool `gorm:"column:is_active;not null;default:true" json:"is_active"`
	Is24h      bool `gorm:"column:is_24h;not null;default:false" json:"is_24h"`

	CreatedAt time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deleted_at,omitempty"`

	// Distance (only in nearby queries, not stored)
	DistanceKm *float64 `gorm:"-" json:"distance_km,omitempty"`

	// Relations
	Hours     []PharmacyHours     `gorm:"foreignKey:PharmacyID" json:"hours,omitempty"`
	Inventory []PharmacyInventory `gorm:"foreignKey:PharmacyID" json:"inventory,omitempty"`
}

func (Pharmacy) TableName() string {
	return "pharmacy.pharmacies"
}
