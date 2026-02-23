// internal/application/queries/list_pharmacies_by_chain_query.go
package queries

type ListPharmaciesByChainQuery struct {
	ChainID string `json:"chain_id"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
}

func (q ListPharmaciesByChainQuery) GetName() string {
	return "ListPharmaciesByChainQuery"
}
