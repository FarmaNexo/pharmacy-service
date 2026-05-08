// internal/application/handlers/upsert_pharmacy_from_event_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/domain/events"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"go.uber.org/zap"
)

// UpsertPharmacyFromEventHandler procesa PHARMACY_DISCOVERED del scraper.
// NO usa mediator: se invoca directamente desde el SQS consumer.
type UpsertPharmacyFromEventHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	logger       *zap.Logger
}

func NewUpsertPharmacyFromEventHandler(pharmacyRepo repositories.PharmacyRepository, logger *zap.Logger) *UpsertPharmacyFromEventHandler {
	return &UpsertPharmacyFromEventHandler{pharmacyRepo: pharmacyRepo, logger: logger}
}

// Handle hace UPSERT por source_pharmacy_code. Idempotente.
func (h *UpsertPharmacyFromEventHandler) Handle(ctx context.Context, data events.PharmacyDiscoveredData) (string, error) {
	id, err := h.pharmacyRepo.UpsertBySource(ctx, repositories.PharmacyUpsertParams{
		SourcePharmacyCode: data.SourcePharmacyCode,
		CanonicalName:      data.CanonicalName,
		RUC:                data.RUC,
		FullAddress:        data.FullAddress,
		Departamento:       data.Departamento,
		Provincia:          data.Provincia,
		Distrito:           data.Distrito,
		Phone:              data.Phone,
		Email:              data.Email,
		Hours:              data.Hours,
		TechnicalDirector:  data.TechnicalDirector,
		ChainID:            data.ChainID,
		ChainName:          data.ChainName,
		Latitude:           data.Latitude,
		Longitude:          data.Longitude,
	})
	if err != nil {
		h.logger.Warn("UPSERT farmacia desde evento falló",
			zap.String("source_pharmacy_code", data.SourcePharmacyCode),
			zap.Error(err),
		)
		return "", err
	}
	h.logger.Info("Farmacia upserteada desde scraper",
		zap.String("id", id),
		zap.String("source_pharmacy_code", data.SourcePharmacyCode),
		zap.String("name", data.CanonicalName),
	)
	return id, nil
}
