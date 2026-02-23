// internal/application/commands/verify_pharmacy_command.go
package commands

type VerifyPharmacyCommand struct {
	ID string `json:"id"`
}

func (c VerifyPharmacyCommand) GetName() string {
	return "VerifyPharmacyCommand"
}
