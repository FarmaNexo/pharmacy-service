// internal/domain/events/pharmacy_events.go
package events

import "time"

const (
	EventPharmacyRegistered     = "PHARMACY_REGISTERED"
	EventPharmacyVerified       = "PHARMACY_VERIFIED"
	EventInventoryUpdated       = "INVENTORY_UPDATED"
	EventAuthorizationUploaded  = "AUTHORIZATION_UPLOADED"
)

type PharmacyEvent struct {
	EventType  string                 `json:"event_type"`
	PharmacyID string                 `json:"pharmacy_id"`
	UserID     string                 `json:"user_id,omitempty"`
	ProductID  string                 `json:"product_id,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

func NewPharmacyEvent(eventType, pharmacyID, userID string) PharmacyEvent {
	return PharmacyEvent{
		EventType:  eventType,
		PharmacyID: pharmacyID,
		UserID:     userID,
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"source":  "pharmacy-service",
			"version": "1.0",
		},
	}
}
