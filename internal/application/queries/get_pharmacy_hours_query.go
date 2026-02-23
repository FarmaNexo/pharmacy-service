// internal/application/queries/get_pharmacy_hours_query.go
package queries

type GetPharmacyHoursQuery struct {
	PharmacyID string `json:"pharmacy_id"`
}

func (q GetPharmacyHoursQuery) GetName() string {
	return "GetPharmacyHoursQuery"
}
