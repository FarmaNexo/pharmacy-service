// Package events — eventos consumidos del scraper.
//
// scraper_events.go define el schema "consumer-side" de los 2 eventos
// que pharmacy-service procesa de la cola farmanexo-{env}-scraper-events:
//   - PHARMACY_DISCOVERED  → UPSERT en pharmacy.pharmacies con PostGIS
//   - INVENTORY_DISCOVERED → lookup farmacia local + lookup producto via
//     HTTP catalog-service + UPSERT en pharmacy.pharmacy_inventory
package events

import (
	"encoding/json"
	"time"
)

// Tipos de eventos del scraper. Solo PHARMACY/INVENTORY tienen handler en
// pharmacy-service. PRODUCT viaja por una cola separada (scraper-product-events)
// consumida por catalog-service.
const (
	ScraperEventPharmacyDiscovered  = "PHARMACY_DISCOVERED"
	ScraperEventInventoryDiscovered = "INVENTORY_DISCOVERED"
)

// ScraperEvent envelope común. Schema idéntico al publisher en scraper-service.
type ScraperEvent struct {
	EventType string          `json:"event_type"`
	SourceID  string          `json:"source_id"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
}

// PharmacyDiscoveredData — payload de PHARMACY_DISCOVERED.
//
// Clave UPSERT: source_pharmacy_code (unique parcial WHERE NOT NULL).
type PharmacyDiscoveredData struct {
	SourcePharmacyCode string   `json:"source_pharmacy_code"`
	CanonicalName      string   `json:"canonical_name"`
	RUC                string   `json:"ruc,omitempty"`
	FullAddress        string   `json:"full_address,omitempty"`
	Departamento       string   `json:"departamento,omitempty"`
	Provincia          string   `json:"provincia,omitempty"`
	Distrito           string   `json:"distrito,omitempty"`
	Phone              string   `json:"phone,omitempty"`
	Email              string   `json:"email,omitempty"`
	Hours              string   `json:"hours,omitempty"`
	TechnicalDirector  string   `json:"technical_director,omitempty"`
	PharmacyType       string   `json:"pharmacy_type,omitempty"`
	ChainID            string   `json:"chain_id,omitempty"`
	ChainName          string   `json:"chain_name,omitempty"`
	Latitude           *float64 `json:"latitude,omitempty"`
	Longitude          *float64 `json:"longitude,omitempty"`
}

// InventoryDiscoveredData — payload de INVENTORY_DISCOVERED.
//
// Solo trae claves naturales — el consumer resuelve los UUIDs internos
// (pharmacy local lookup + catalog HTTP lookup). Si alguno no existe aún,
// el handler hace "fail soft" para que SQS reintente tras visibility timeout.
type InventoryDiscoveredData struct {
	SourcePharmacyCode string    `json:"source_pharmacy_code"`
	SourceProductCode  int       `json:"source_product_code"`
	Concentration      string    `json:"concentration"`
	Price              float64   `json:"price"`
	Stock              int       `json:"stock"`
	IsAvailable        bool      `json:"is_available"`
	ObservedAt         time.Time `json:"observed_at"`
}
