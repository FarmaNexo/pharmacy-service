// internal/application/commands/update_pharmacy_hours_command.go
package commands

type HoursEntry struct {
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	IsClosed  bool   `json:"is_closed"`
}

type UpdatePharmacyHoursCommand struct {
	PharmacyID string       `json:"pharmacy_id"`
	Hours      []HoursEntry `json:"hours"`
}

func (c UpdatePharmacyHoursCommand) GetName() string {
	return "UpdatePharmacyHoursCommand"
}
