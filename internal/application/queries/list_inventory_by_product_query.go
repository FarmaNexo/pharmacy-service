// internal/application/queries/list_inventory_by_product_query.go
package queries

// ListInventoryByProductQuery devuelve todas las farmacias que venden un
// producto (comparador de precios).
//
// Geolocalización (HU-014, opcional):
//   - Latitude/Longitude: si ambos son != 0, la query devuelve distance_km
//     y ordena por cercanía en lugar de precio.
//   - RadiusKm: si > 0 y hay lat/lng, filtra farmacias dentro de ese radio.
//     Default a 0 = sin filtro de radio (solo cálculo + orden).
type ListInventoryByProductQuery struct {
	ProductID string  `json:"product_id"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	RadiusKm  float64 `json:"radius_km,omitempty"`
}

func (q ListInventoryByProductQuery) GetName() string {
	return "ListInventoryByProductQuery"
}
