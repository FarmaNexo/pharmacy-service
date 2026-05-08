// internal/domain/entities/pharmacy.go
package entities

import "time"

type Pharmacy struct {
	ID          string     `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	// OwnerUserID es NULL para farmacias scrapeadas (aún no reclamadas por un dueño).
	OwnerUserID *string    `gorm:"column:owner_user_id;type:uuid" json:"owner_user_id,omitempty"`
	Name        string     `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Slug        string     `gorm:"column:slug;type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description string     `gorm:"column:description;type:text" json:"description"`
	Phone       string     `gorm:"column:phone;type:varchar(20)" json:"phone"`
	Email       string     `gorm:"column:email;type:varchar(255)" json:"email"`
	Website     string     `gorm:"column:website;type:varchar(500)" json:"website"`
	LogoURL                 string `gorm:"column:logo_url;type:varchar(500)" json:"logo_url"`
	AuthorizationDocumentURL string `gorm:"column:authorization_document_url;type:varchar(500)" json:"authorization_document_url,omitempty"`

	// Fuente externa (DIGEMID scraper). NULL para farmacias registradas manualmente.
	SourcePharmacyCode *string `gorm:"column:source_pharmacy_code;type:varchar(20)" json:"source_pharmacy_code,omitempty"`

	// Datos regulatorios peruanos
	RUC               string `gorm:"column:ruc;type:varchar(20)" json:"ruc,omitempty"`
	TechnicalDirector string `gorm:"column:technical_director;type:varchar(255)" json:"technical_director,omitempty"`
	HoursRaw          string `gorm:"column:hours_raw;type:text" json:"hours_raw,omitempty"`

	// Chain
	ChainID   string `gorm:"column:chain_id;type:varchar(100)" json:"chain_id,omitempty"`
	ChainName string `gorm:"column:chain_name;type:varchar(255)" json:"chain_name,omitempty"`

	// Address — street y city ahora nullable (relaxed en migración 000004).
	Street     string `gorm:"column:street;type:varchar(255)" json:"street"`
	City       string `gorm:"column:city;type:varchar(100)" json:"city"`
	State      string `gorm:"column:state;type:varchar(100)" json:"state"`
	PostalCode string `gorm:"column:postal_code;type:varchar(20)" json:"postal_code"`
	Country    string `gorm:"column:country;type:varchar(100);not null;default:'Perú'" json:"country"`

	// Geolocation - stored as PostGIS GEOGRAPHY. SELECTs hacen ST_Y/ST_X
	// con alias `latitude`/`longitude` para que GORM los mapee aquí via column tag.
	// INSERT/UPDATE usan raw SQL con ST_MakePoint; GORM nunca intenta usar
	// estas columnas en queries auto-generadas porque todas las operaciones
	// pasan por Raw()/Exec().
	Latitude  float64 `gorm:"column:latitude" json:"latitude"`
	Longitude float64 `gorm:"column:longitude" json:"longitude"`

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
