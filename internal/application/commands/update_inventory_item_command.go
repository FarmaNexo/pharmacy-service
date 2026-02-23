// internal/application/commands/update_inventory_item_command.go
package commands

type UpdateInventoryItemCommand struct {
	PharmacyID  string  `json:"pharmacy_id"`
	ProductID   string  `json:"product_id"`
	Stock       int     `json:"stock"`
	Price       float64 `json:"price"`
	IsAvailable bool    `json:"is_available"`
}

func (c UpdateInventoryItemCommand) GetName() string {
	return "UpdateInventoryItemCommand"
}
