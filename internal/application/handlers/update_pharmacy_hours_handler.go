// internal/application/handlers/update_pharmacy_hours_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"go.uber.org/zap"
)

type UpdatePharmacyHoursHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	hoursRepo    repositories.PharmacyHoursRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewUpdatePharmacyHoursHandler(
	pharmacyRepo repositories.PharmacyRepository,
	hoursRepo repositories.PharmacyHoursRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *UpdatePharmacyHoursHandler {
	return &UpdatePharmacyHoursHandler{
		pharmacyRepo: pharmacyRepo,
		hoursRepo:    hoursRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *UpdatePharmacyHoursHandler) Handle(ctx context.Context, cmd commands.UpdatePharmacyHoursCommand) (*common.ApiResponse[responses.HoursListResponse], error) {
	pharmacy, err := h.pharmacyRepo.FindByID(ctx, cmd.PharmacyID)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.HoursListResponse]("Farmacia no encontrada"), nil
	}

	hoursEntities := make([]entities.PharmacyHours, len(cmd.Hours))
	for i, h := range cmd.Hours {
		hoursEntities[i] = entities.PharmacyHours{
			PharmacyID: cmd.PharmacyID,
			DayOfWeek:  h.DayOfWeek,
			OpenTime:   h.OpenTime,
			CloseTime:  h.CloseTime,
			IsClosed:   h.IsClosed,
		}
	}

	if err := h.hoursRepo.UpsertHours(ctx, cmd.PharmacyID, hoursEntities); err != nil {
		h.logger.Error("Error actualizando horarios", zap.Error(err))
		return common.InternalServerErrorResponse[responses.HoursListResponse]("Error actualizando horarios"), nil
	}

	go func() {
		_ = h.cacheService.Delete(context.Background(), "cache:pharmacy:pharmacy:"+cmd.PharmacyID)
	}()

	hours, _ := h.hoursRepo.FindByPharmacyID(ctx, cmd.PharmacyID)
	hoursResponses := make([]responses.HoursResponse, len(hours))
	for i, hr := range hours {
		hoursResponses[i] = responses.ToHoursResponse(&hr)
	}

	h.logger.Info("Horarios actualizados", zap.String("pharmacy_id", cmd.PharmacyID))
	return common.OkResponse(responses.HoursListResponse{Hours: hoursResponses}), nil
}

var _ mediator.RequestHandler[commands.UpdatePharmacyHoursCommand, responses.HoursListResponse] = (*UpdatePharmacyHoursHandler)(nil)
