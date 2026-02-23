// internal/application/queries/list_pharmacy_inventory_query.go
package queries

type ListPharmacyInventoryQuery struct {
	PharmacyID string `json:"pharmacy_id"`
}

func (q ListPharmacyInventoryQuery) GetName() string {
	return "ListPharmacyInventoryQuery"
}
