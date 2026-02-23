// internal/infrastructure/persistence/postgres/inventory_repository_impl.go
package postgres

import (
	"context"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InventoryRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewInventoryRepository(db *gorm.DB, logger *zap.Logger) repositories.InventoryRepository {
	return &InventoryRepositoryImpl{db: db, logger: logger}
}

func (r *InventoryRepositoryImpl) Create(ctx context.Context, item *entities.PharmacyInventory) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *InventoryRepositoryImpl) FindByPharmacyID(ctx context.Context, pharmacyID string) ([]entities.PharmacyInventory, error) {
	var items []entities.PharmacyInventory
	result := r.db.WithContext(ctx).
		Where("pharmacy_id = ?", pharmacyID).
		Order("created_at DESC").
		Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (r *InventoryRepositoryImpl) FindByPharmacyAndProduct(ctx context.Context, pharmacyID, productID string) (*entities.PharmacyInventory, error) {
	var item entities.PharmacyInventory
	result := r.db.WithContext(ctx).
		Where("pharmacy_id = ? AND product_id = ?", pharmacyID, productID).
		First(&item)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &item, nil
}

func (r *InventoryRepositoryImpl) Update(ctx context.Context, item *entities.PharmacyInventory) error {
	item.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *InventoryRepositoryImpl) Delete(ctx context.Context, pharmacyID, productID string) error {
	result := r.db.WithContext(ctx).
		Where("pharmacy_id = ? AND product_id = ?", pharmacyID, productID).
		Delete(&entities.PharmacyInventory{})
	return result.Error
}

var _ repositories.InventoryRepository = (*InventoryRepositoryImpl)(nil)
