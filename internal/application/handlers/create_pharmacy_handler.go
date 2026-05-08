// internal/application/handlers/create_pharmacy_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
	"github.com/farmanexo/pharmacy-service/internal/domain/events"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type CreatePharmacyHandler struct {
	pharmacyRepo   repositories.PharmacyRepository
	eventPublisher services.EventPublisher
	cacheService   services.CacheService
	logger         *zap.Logger
}

func NewCreatePharmacyHandler(
	pharmacyRepo repositories.PharmacyRepository,
	eventPublisher services.EventPublisher,
	cacheService services.CacheService,
	logger *zap.Logger,
) *CreatePharmacyHandler {
	return &CreatePharmacyHandler{
		pharmacyRepo:   pharmacyRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
		logger:         logger,
	}
}

func (h *CreatePharmacyHandler) Handle(ctx context.Context, cmd commands.CreatePharmacyCommand) (*common.ApiResponse[responses.PharmacyResponse], error) {
	slug := cmd.Slug
	if slug == "" {
		slug = GenerateSlug(cmd.Name)
	}

	existing, _ := h.pharmacyRepo.FindBySlug(ctx, slug)
	if existing != nil {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	country := cmd.Country
	if country == "" {
		country = "Perú"
	}

	pharmacy := &entities.Pharmacy{
		ID:                uuid.New().String(),
		Name:              cmd.Name,
		Slug:              slug,
		Description:       cmd.Description,
		Phone:             cmd.Phone,
		Email:             cmd.Email,
		Website:           cmd.Website,
		Street:            cmd.Street,
		City:              cmd.City,
		State:             cmd.State,
		PostalCode:        cmd.PostalCode,
		Country:           country,
		Latitude:          cmd.Latitude,
		Longitude:         cmd.Longitude,
		Is24h:             cmd.Is24h,
		ChainID:           cmd.ChainID,
		ChainName:         cmd.ChainName,
		RUC:               cmd.RUC,
		TechnicalDirector: cmd.TechnicalDirector,
		HoursRaw:          cmd.HoursRaw,
		IsActive:          true,
		IsVerified:        false,
	}
	// OwnerUserID puede venir vacío (flujo scraper). Si se proporciona, lo seteamos.
	if cmd.OwnerUserID != "" {
		owner := cmd.OwnerUserID
		pharmacy.OwnerUserID = &owner
	}

	if err := h.pharmacyRepo.Create(ctx, pharmacy); err != nil {
		h.logger.Error("Error creando farmacia", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PharmacyResponse]("Error registrando farmacia"), nil
	}

	go func() {
		_ = h.cacheService.DeleteByPattern(context.Background(), "cache:pharmacy:*")
	}()

	go func() {
		event := events.NewPharmacyEvent(events.EventPharmacyRegistered, pharmacy.ID, cmd.OwnerUserID)
		if err := h.eventPublisher.Publish(context.Background(), event); err != nil {
			h.logger.Error("Error publicando evento PHARMACY_REGISTERED", zap.Error(err))
		}
	}()

	created, err := h.pharmacyRepo.FindByID(ctx, pharmacy.ID)
	if err != nil {
		return common.CreatedResponse(responses.ToPharmacyResponse(pharmacy)), nil
	}

	h.logger.Info("Farmacia registrada exitosamente",
		zap.String("pharmacy_id", pharmacy.ID),
		zap.String("name", pharmacy.Name),
		zap.String("owner", cmd.OwnerUserID),
	)

	return common.CreatedResponse(responses.ToPharmacyResponse(created)), nil
}

var _ mediator.RequestHandler[commands.CreatePharmacyCommand, responses.PharmacyResponse] = (*CreatePharmacyHandler)(nil)
