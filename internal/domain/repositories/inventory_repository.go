// internal/domain/repositories/inventory_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
)

type InventoryRepository interface {
	Create(ctx context.Context, item *entities.PharmacyInventory) error
	FindByPharmacyID(ctx context.Context, pharmacyID string) ([]entities.PharmacyInventory, error)
	FindByPharmacyAndProduct(ctx context.Context, pharmacyID, productID string) (*entities.PharmacyInventory, error)
	Update(ctx context.Context, item *entities.PharmacyInventory) error
	Delete(ctx context.Context, pharmacyID, productID string) error
}
