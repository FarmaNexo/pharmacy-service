// internal/application/handlers/update_pharmacy_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type UpdatePharmacyHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewUpdatePharmacyHandler(
	pharmacyRepo repositories.PharmacyRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *UpdatePharmacyHandler {
	return &UpdatePharmacyHandler{
		pharmacyRepo: pharmacyRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *UpdatePharmacyHandler) Handle(ctx context.Context, cmd commands.UpdatePharmacyCommand) (*common.ApiResponse[responses.PharmacyResponse], error) {
	pharmacy, err := h.pharmacyRepo.FindByID(ctx, cmd.ID)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.PharmacyResponse]("Farmacia no encontrada"), nil
	}

	if cmd.Name != "" {
		pharmacy.Name = cmd.Name
	}
	if cmd.Slug != "" {
		existing, _ := h.pharmacyRepo.FindBySlug(ctx, cmd.Slug)
		if existing != nil && existing.ID != cmd.ID {
			cmd.Slug = cmd.Slug + "-" + uuid.New().String()[:8]
		}
		pharmacy.Slug = cmd.Slug
	}
	if cmd.Description != "" {
		pharmacy.Description = cmd.Description
	}
	if cmd.Phone != "" {
		pharmacy.Phone = cmd.Phone
	}
	if cmd.Email != "" {
		pharmacy.Email = cmd.Email
	}
	if cmd.Website != "" {
		pharmacy.Website = cmd.Website
	}
	if cmd.Street != "" {
		pharmacy.Street = cmd.Street
	}
	if cmd.City != "" {
		pharmacy.City = cmd.City
	}
	if cmd.State != "" {
		pharmacy.State = cmd.State
	}
	if cmd.PostalCode != "" {
		pharmacy.PostalCode = cmd.PostalCode
	}
	if cmd.Country != "" {
		pharmacy.Country = cmd.Country
	}
	if cmd.Latitude != 0 {
		pharmacy.Latitude = cmd.Latitude
	}
	if cmd.Longitude != 0 {
		pharmacy.Longitude = cmd.Longitude
	}
	pharmacy.Is24h = cmd.Is24h
	pharmacy.IsActive = cmd.IsActive
	if cmd.ChainID != "" {
		pharmacy.ChainID = cmd.ChainID
	}
	if cmd.ChainName != "" {
		pharmacy.ChainName = cmd.ChainName
	}
	if cmd.RUC != "" {
		pharmacy.RUC = cmd.RUC
	}
	if cmd.TechnicalDirector != "" {
		pharmacy.TechnicalDirector = cmd.TechnicalDirector
	}
	if cmd.HoursRaw != "" {
		pharmacy.HoursRaw = cmd.HoursRaw
	}

	if err := h.pharmacyRepo.Update(ctx, pharmacy); err != nil {
		h.logger.Error("Error actualizando farmacia", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PharmacyResponse]("Error actualizando farmacia"), nil
	}

	go func() {
		_ = h.cacheService.Delete(context.Background(), "cache:pharmacy:pharmacy:"+cmd.ID)
		_ = h.cacheService.DeleteByPattern(context.Background(), "cache:pharmacy:nearby:*")
	}()

	updated, err := h.pharmacyRepo.FindByID(ctx, cmd.ID)
	if err != nil {
		return common.OkResponse(responses.ToPharmacyResponse(pharmacy)), nil
	}

	h.logger.Info("Farmacia actualizada", zap.String("pharmacy_id", cmd.ID))
	return common.OkResponse(responses.ToPharmacyResponse(updated)), nil
}

var _ mediator.RequestHandler[commands.UpdatePharmacyCommand, responses.PharmacyResponse] = (*UpdatePharmacyHandler)(nil)
