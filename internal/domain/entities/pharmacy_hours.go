// internal/domain/entities/pharmacy_hours.go
package entities

import "time"

type PharmacyHours struct {
	ID         string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PharmacyID string    `gorm:"column:pharmacy_id;type:uuid;not null" json:"pharmacy_id"`
	DayOfWeek  int       `gorm:"column:day_of_week;not null" json:"day_of_week"`
	OpenTime   string    `gorm:"column:open_time;type:time;not null" json:"open_time"`
	CloseTime  string    `gorm:"column:close_time;type:time;not null" json:"close_time"`
	IsClosed   bool      `gorm:"column:is_closed;not null;default:false" json:"is_closed"`
	CreatedAt  time.Time `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (PharmacyHours) TableName() string {
	return "pharmacy.pharmacy_hours"
}
