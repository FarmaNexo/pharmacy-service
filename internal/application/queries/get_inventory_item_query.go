// internal/application/queries/get_inventory_item_query.go
package queries

// GetInventoryItemQuery — lookup de un item de inventario por
// (pharmacy_id, product_id). Lo consume order-service al validar
// stock antes de agregar al carrito o hacer checkout.
type GetInventoryItemQuery struct {
	PharmacyID string `json:"pharmacy_id"`
	ProductID  string `json:"product_id"`
}

func (q GetInventoryItemQuery) GetName() string {
	return "GetInventoryItemQuery"
}
