// internal/presentation/dto/responses/pharmacy_response.go
package responses

import (
	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
)

type PharmacyResponse struct {
	ID                       string  `json:"id"`
	OwnerUserID              *string `json:"owner_user_id,omitempty"`
	Name                     string  `json:"name"`
	Slug                     string  `json:"slug"`
	Description              string  `json:"description"`
	Phone                    string  `json:"phone"`
	Email                    string  `json:"email"`
	Website                  string  `json:"website"`
	LogoURL                  string  `json:"logo_url,omitempty"`
	AuthorizationDocumentURL string  `json:"authorization_document_url,omitempty"`
	ChainID                  string  `json:"chain_id,omitempty"`
	ChainName                string  `json:"chain_name,omitempty"`
	SourcePharmacyCode       *string `json:"source_pharmacy_code,omitempty"`
	RUC                      string  `json:"ruc,omitempty"`
	TechnicalDirector        string  `json:"technical_director,omitempty"`
	HoursRaw                 string  `json:"hours_raw,omitempty"`
	Street                   string  `json:"street"`
	City                     string  `json:"city"`
	State                    string  `json:"state"`
	PostalCode               string  `json:"postal_code"`
	Country                  string  `json:"country"`
	Latitude                 float64 `json:"latitude"`
	Longitude                float64 `json:"longitude"`
	IsVerified               bool    `json:"is_verified"`
	IsActive                 bool    `json:"is_active"`
	Is24h                    bool    `json:"is_24h"`
	DistanceKm               *float64        `json:"distance_km,omitempty"`
	Hours                    []HoursResponse `json:"hours,omitempty"`
	CreatedAt                string  `json:"created_at"`
	UpdatedAt                string  `json:"updated_at"`
}

type PharmacyListResponse struct {
	Pharmacies []PharmacyResponse `json:"pharmacies"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

type NearbyPharmaciesResponse struct {
	Pharmacies []PharmacyResponse `json:"pharmacies"`
	Total      int                `json:"total"`
	RadiusKm   float64            `json:"radius_km"`
}

type HoursResponse struct {
	ID        string `json:"id"`
	DayOfWeek int    `json:"day_of_week"`
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	IsClosed  bool   `json:"is_closed"`
}

type HoursListResponse struct {
	Hours []HoursResponse `json:"hours"`
}

type InventoryItemResponse struct {
	ID               string   `json:"id"`
	PharmacyID       string   `json:"pharmacy_id"`
	PharmacySlug     string   `json:"pharmacy_slug,omitempty"`
	PharmacyName     string   `json:"pharmacy_name,omitempty"`
	PharmacyDistrict string   `json:"pharmacy_district,omitempty"` // Distrito (DIGEMID Perú)
	PharmacyAddress  string   `json:"pharmacy_address,omitempty"`  // Dirección de calle
	ProductID        string   `json:"product_id"`
	Stock            int      `json:"stock"`
	Price            float64  `json:"price"`
	IsAvailable      bool     `json:"is_available"`
	DistanceKm       *float64 `json:"distance_km,omitempty"` // HU-014: solo presente si la query incluyó lat/lng
	// HU-016 — alerta de diferencia excesiva de precios:
	// `district_avg_price`: promedio del producto en el distrito (referencia educativa).
	// `is_overpriced`: true si price > avg * threshold y el distrito tiene suficientes farmacias.
	// `overprice_pct`: cuánto encima del promedio (ej. 0.45 = 45% más caro). Nil si no aplica.
	DistrictAvgPrice *float64 `json:"district_avg_price,omitempty"`
	IsOverpriced     bool     `json:"is_overpriced"`
	OverpricePct     *float64 `json:"overprice_pct,omitempty"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

type InventoryListResponse struct {
	Items []InventoryItemResponse `json:"items"`
	Total int                     `json:"total"`
}

type AuthorizationDocumentResponse struct {
	PharmacyID  string `json:"pharmacy_id"`
	DocumentURL string `json:"document_url"`
}

type EmptyResponse struct{}

func ToPharmacyResponse(p *entities.Pharmacy) PharmacyResponse {
	resp := PharmacyResponse{
		ID:                       p.ID,
		OwnerUserID:              p.OwnerUserID,
		Name:                     p.Name,
		Slug:                     p.Slug,
		Description:              p.Description,
		Phone:                    p.Phone,
		Email:                    p.Email,
		Website:                  p.Website,
		LogoURL:                  p.LogoURL,
		AuthorizationDocumentURL: p.AuthorizationDocumentURL,
		ChainID:                  p.ChainID,
		ChainName:                p.ChainName,
		SourcePharmacyCode:       p.SourcePharmacyCode,
		RUC:                      p.RUC,
		TechnicalDirector:        p.TechnicalDirector,
		HoursRaw:                 p.HoursRaw,
		Street:                   p.Street,
		City:                     p.City,
		State:                    p.State,
		PostalCode:               p.PostalCode,
		Country:                  p.Country,
		Latitude:                 p.Latitude,
		Longitude:                p.Longitude,
		IsVerified:               p.IsVerified,
		IsActive:                 p.IsActive,
		Is24h:                    p.Is24h,
		DistanceKm:               p.DistanceKm,
		CreatedAt:                p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:                p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if p.Hours != nil {
		resp.Hours = make([]HoursResponse, len(p.Hours))
		for i, h := range p.Hours {
			resp.Hours[i] = HoursResponse{
				ID:        h.ID,
				DayOfWeek: h.DayOfWeek,
				OpenTime:  h.OpenTime,
				CloseTime: h.CloseTime,
				IsClosed:  h.IsClosed,
			}
		}
	}

	return resp
}

func ToInventoryItemResponse(item *entities.PharmacyInventory) InventoryItemResponse {
	return InventoryItemResponse{
		ID:          item.ID,
		PharmacyID:  item.PharmacyID,
		ProductID:   item.ProductID,
		Stock:       item.Stock,
		Price:       item.Price,
		IsAvailable: item.IsAvailable,
		CreatedAt:   item.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func ToHoursResponse(h *entities.PharmacyHours) HoursResponse {
	return HoursResponse{
		ID:        h.ID,
		DayOfWeek: h.DayOfWeek,
		OpenTime:  h.OpenTime,
		CloseTime: h.CloseTime,
		IsClosed:  h.IsClosed,
	}
}
