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

type PharmacyRepository interface {
	Create(ctx context.Context, pharmacy *entities.Pharmacy) error
	FindByID(ctx context.Context, id string) (*entities.Pharmacy, error)
	FindByIDWithDeleted(ctx context.Context, id string) (*entities.Pharmacy, error)
	FindBySlug(ctx context.Context, slug string) (*entities.Pharmacy, error)
	FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*entities.Pharmacy, error)
	FindAll(ctx context.Context, page, limit int) (*PaginatedResult, error)
	FindNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*entities.Pharmacy, error)
	FindByChainID(ctx context.Context, chainID string, page, limit int) (*PaginatedResult, error)
	Update(ctx context.Context, pharmacy *entities.Pharmacy) error
	SoftDelete(ctx context.Context, id string) error
}
