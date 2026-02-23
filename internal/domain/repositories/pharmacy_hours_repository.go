// internal/domain/repositories/pharmacy_hours_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
)

type PharmacyHoursRepository interface {
	FindByPharmacyID(ctx context.Context, pharmacyID string) ([]entities.PharmacyHours, error)
	UpsertHours(ctx context.Context, pharmacyID string, hours []entities.PharmacyHours) error
	DeleteByPharmacyID(ctx context.Context, pharmacyID string) error
}
