// internal/application/queries/get_pharmacy_by_slug_query.go
package queries

// GetPharmacyBySlugQuery consulta para obtener una farmacia por slug (SEO-friendly URL).
type GetPharmacyBySlugQuery struct {
	Slug string `json:"slug"`
}

func (q GetPharmacyBySlugQuery) GetName() string {
	return "GetPharmacyBySlugQuery"
}
