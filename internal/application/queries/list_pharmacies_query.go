// internal/application/queries/list_pharmacies_query.go
package queries

type ListPharmaciesQuery struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

func (q ListPharmaciesQuery) GetName() string {
	return "ListPharmaciesQuery"
}
