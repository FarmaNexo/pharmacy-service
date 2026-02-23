// internal/application/queries/get_pharmacy_query.go
package queries

type GetPharmacyQuery struct {
	ID string `json:"id"`
}

func (q GetPharmacyQuery) GetName() string {
	return "GetPharmacyQuery"
}
