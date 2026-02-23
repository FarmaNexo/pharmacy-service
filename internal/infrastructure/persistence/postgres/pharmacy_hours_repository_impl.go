// internal/infrastructure/persistence/postgres/pharmacy_hours_repository_impl.go
package postgres

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PharmacyHoursRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewPharmacyHoursRepository(db *gorm.DB, logger *zap.Logger) repositories.PharmacyHoursRepository {
	return &PharmacyHoursRepositoryImpl{db: db, logger: logger}
}

func (r *PharmacyHoursRepositoryImpl) FindByPharmacyID(ctx context.Context, pharmacyID string) ([]entities.PharmacyHours, error) {
	var hours []entities.PharmacyHours
	result := r.db.WithContext(ctx).
		Where("pharmacy_id = ?", pharmacyID).
		Order("day_of_week ASC").
		Find(&hours)
	if result.Error != nil {
		return nil, result.Error
	}
	return hours, nil
}

func (r *PharmacyHoursRepositoryImpl) UpsertHours(ctx context.Context, pharmacyID string, hours []entities.PharmacyHours) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing hours
		if err := tx.Where("pharmacy_id = ?", pharmacyID).Delete(&entities.PharmacyHours{}).Error; err != nil {
			return err
		}
		// Insert new hours
		for i := range hours {
			hours[i].PharmacyID = pharmacyID
			if err := tx.Create(&hours[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *PharmacyHoursRepositoryImpl) DeleteByPharmacyID(ctx context.Context, pharmacyID string) error {
	return r.db.WithContext(ctx).Where("pharmacy_id = ?", pharmacyID).Delete(&entities.PharmacyHours{}).Error
}

var _ repositories.PharmacyHoursRepository = (*PharmacyHoursRepositoryImpl)(nil)
