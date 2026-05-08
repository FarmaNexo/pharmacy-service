// internal/application/commands/create_pharmacy_command.go
package commands

type CreatePharmacyCommand struct {
	OwnerUserID       string  `json:"owner_user_id"`
	Name              string  `json:"name"`
	Slug              string  `json:"slug"`
	Description       string  `json:"description"`
	Phone             string  `json:"phone"`
	Email             string  `json:"email"`
	Website           string  `json:"website"`
	Street            string  `json:"street"`
	City              string  `json:"city"`
	State             string  `json:"state"`
	PostalCode        string  `json:"postal_code"`
	Country           string  `json:"country"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Is24h             bool    `json:"is_24h"`
	ChainID           string  `json:"chain_id"`
	ChainName         string  `json:"chain_name"`
	RUC               string  `json:"ruc"`
	TechnicalDirector string  `json:"technical_director"`
	HoursRaw          string  `json:"hours_raw"`
}

func (c CreatePharmacyCommand) GetName() string {
	return "CreatePharmacyCommand"
}
