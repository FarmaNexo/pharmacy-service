// internal/domain/repositories/pharmacy_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
)

type PaginatedResult struct {
	Total int64
	Items []*entities.Pharmacy
}

// PharmacyUpsertParams — payload para UpsertBySource. Refleja los campos del
// PHARMACY_DISCOVERED del scraper. Clave UPSERT: source_pharmacy_code
// (unique parcial WHERE NOT NULL, definida en migration 000004).
type PharmacyUpsertParams struct {
	SourcePharmacyCode string
	CanonicalName      string
	RUC                string
	FullAddress        string
	Departamento       string
	Provincia          string
	Distrito           string
	Phone              string
	Email              string
	Hours              string
	TechnicalDirector  string
	ChainID            string
	ChainName          string
	Latitude           *float64
	Longitude          *float64
}

type PharmacyRepository interface {
	Create(ctx context.Context, pharmacy *entities.Pharmacy) error
	FindByID(ctx context.Context, id string) (*entities.Pharmacy, error)
	FindByIDWithDeleted(ctx context.Context, id string) (*entities.Pharmacy, error)
	FindBySlug(ctx context.Context, slug string) (*entities.Pharmacy, error)
	// FindBySourceCode busca una farmacia por su clave natural DIGEMID.
	// Lo usa el consumer SQS de INVENTORY_DISCOVERED para resolver
	// source_pharmacy_code → pharmacy_id local. Retorna nil sin error si
	// no existe (caller hace fail-soft para redrive).
	FindBySourceCode(ctx context.Context, sourcePharmacyCode string) (*entities.Pharmacy, error)
	FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*entities.Pharmacy, error)
	FindAll(ctx context.Context, page, limit int) (*PaginatedResult, error)
	FindNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*entities.Pharmacy, error)
	FindByChainID(ctx context.Context, chainID string, page, limit int) (*PaginatedResult, error)
	Update(ctx context.Context, pharmacy *entities.Pharmacy) error
	// UpsertBySource hace INSERT ... ON CONFLICT (source_pharmacy_code)
	// DO UPDATE. Idempotente. Retorna el id (UUID) de la farmacia resultante.
	// Usado por el SQS consumer de PHARMACY_DISCOVERED.
	UpsertBySource(ctx context.Context, params PharmacyUpsertParams) (string, error)
	SoftDelete(ctx context.Context, id string) error
}
