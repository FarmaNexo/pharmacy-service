// internal/application/queries/search_nearby_pharmacies_query.go
package queries

type SearchNearbyPharmaciesQuery struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	RadiusKm  float64 `json:"radius_km"`
	Limit     int     `json:"limit"`
}

func (q SearchNearbyPharmaciesQuery) GetName() string {
	return "SearchNearbyPharmaciesQuery"
}
