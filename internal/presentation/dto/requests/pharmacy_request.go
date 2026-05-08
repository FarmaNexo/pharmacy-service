// internal/presentation/dto/requests/pharmacy_request.go
package requests

type CreatePharmacyRequest struct {
	Name              string  `json:"name"`
	Slug              string  `json:"slug,omitempty"`
	Description       string  `json:"description,omitempty"`
	Phone             string  `json:"phone"`
	Email             string  `json:"email,omitempty"`
	Website           string  `json:"website,omitempty"`
	Street            string  `json:"street"`
	City              string  `json:"city"`
	State             string  `json:"state,omitempty"`
	PostalCode        string  `json:"postal_code,omitempty"`
	Country           string  `json:"country,omitempty"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Is24h             bool    `json:"is_24h"`
	ChainID           string  `json:"chain_id,omitempty"`
	ChainName         string  `json:"chain_name,omitempty"`
	RUC               string  `json:"ruc,omitempty"`
	TechnicalDirector string  `json:"technical_director,omitempty"`
	HoursRaw          string  `json:"hours_raw,omitempty"`
}

type UpdatePharmacyRequest struct {
	Name              string  `json:"name,omitempty"`
	Slug              string  `json:"slug,omitempty"`
	Description       string  `json:"description,omitempty"`
	Phone             string  `json:"phone,omitempty"`
	Email             string  `json:"email,omitempty"`
	Website           string  `json:"website,omitempty"`
	Street            string  `json:"street,omitempty"`
	City              string  `json:"city,omitempty"`
	State             string  `json:"state,omitempty"`
	PostalCode        string  `json:"postal_code,omitempty"`
	Country           string  `json:"country,omitempty"`
	Latitude          float64 `json:"latitude"`
	Longitude         float64 `json:"longitude"`
	Is24h             bool    `json:"is_24h"`
	IsActive          bool    `json:"is_active"`
	ChainID           string  `json:"chain_id,omitempty"`
	ChainName         string  `json:"chain_name,omitempty"`
	RUC               string  `json:"ruc,omitempty"`
	TechnicalDirector string  `json:"technical_director,omitempty"`
	HoursRaw          string  `json:"hours_raw,omitempty"`
}

type NearbyPharmaciesRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	RadiusKm  float64 `json:"radius_km"`
	Limit     int     `json:"limit,omitempty"`
}

type AddInventoryItemRequest struct {
	ProductID   string  `json:"product_id"`
	Stock       int     `json:"stock"`
	Price       float64 `json:"price"`
	IsAvailable bool    `json:"is_available"`
}

type UpdateInventoryItemRequest struct {
	Stock       int     `json:"stock"`
	Price       float64 `json:"price"`
	IsAvailable bool    `json:"is_available"`
}

type HoursEntryRequest struct {
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	IsClosed  bool   `json:"is_closed"`
}

type UpdatePharmacyHoursRequest struct {
	Hours []HoursEntryRequest `json:"hours"`
}
