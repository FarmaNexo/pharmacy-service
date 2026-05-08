// internal/application/handlers/verify_pharmacy_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/domain/events"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"go.uber.org/zap"
)

type VerifyPharmacyHandler struct {
	pharmacyRepo   repositories.PharmacyRepository
	eventPublisher services.EventPublisher
	cacheService   services.CacheService
	logger         *zap.Logger
}

func NewVerifyPharmacyHandler(
	pharmacyRepo repositories.PharmacyRepository,
	eventPublisher services.EventPublisher,
	cacheService services.CacheService,
	logger *zap.Logger,
) *VerifyPharmacyHandler {
	return &VerifyPharmacyHandler{
		pharmacyRepo:   pharmacyRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
		logger:         logger,
	}
}

func (h *VerifyPharmacyHandler) Handle(ctx context.Context, cmd commands.VerifyPharmacyCommand) (*common.ApiResponse[responses.PharmacyResponse], error) {
	pharmacy, err := h.pharmacyRepo.FindByID(ctx, cmd.ID)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.PharmacyResponse]("Farmacia no encontrada"), nil
	}

	if pharmacy.IsVerified {
		return common.OkResponse(responses.ToPharmacyResponse(pharmacy)), nil
	}

	pharmacy.IsVerified = true
	if err := h.pharmacyRepo.Update(ctx, pharmacy); err != nil {
		h.logger.Error("Error verificando farmacia", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PharmacyResponse]("Error verificando farmacia"), nil
	}

	go func() {
		_ = h.cacheService.Delete(context.Background(), "cache:pharmacy:pharmacy:"+cmd.ID)
		_ = h.cacheService.DeleteByPattern(context.Background(), "cache:pharmacy:nearby:*")
	}()

	go func() {
		ownerID := ""
		if pharmacy.OwnerUserID != nil {
			ownerID = *pharmacy.OwnerUserID
		}
		event := events.NewPharmacyEvent(events.EventPharmacyVerified, pharmacy.ID, ownerID)
		if err := h.eventPublisher.Publish(context.Background(), event); err != nil {
			h.logger.Error("Error publicando evento PHARMACY_VERIFIED", zap.Error(err))
		}
	}()

	h.logger.Info("Farmacia verificada", zap.String("pharmacy_id", cmd.ID))

	updated, _ := h.pharmacyRepo.FindByID(ctx, cmd.ID)
	if updated != nil {
		return common.OkResponse(responses.ToPharmacyResponse(updated)), nil
	}
	return common.OkResponse(responses.ToPharmacyResponse(pharmacy)), nil
}

var _ mediator.RequestHandler[commands.VerifyPharmacyCommand, responses.PharmacyResponse] = (*VerifyPharmacyHandler)(nil)
