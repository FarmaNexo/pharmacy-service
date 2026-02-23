// internal/application/commands/remove_inventory_item_command.go
package commands

type RemoveInventoryItemCommand struct {
	PharmacyID string `json:"pharmacy_id"`
	ProductID  string `json:"product_id"`
}

func (c RemoveInventoryItemCommand) GetName() string {
	return "RemoveInventoryItemCommand"
}
