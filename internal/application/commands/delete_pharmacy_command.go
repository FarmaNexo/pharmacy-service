// internal/application/commands/delete_pharmacy_command.go
package commands

type DeletePharmacyCommand struct {
	ID string `json:"id"`
}

func (c DeletePharmacyCommand) GetName() string {
	return "DeletePharmacyCommand"
}
