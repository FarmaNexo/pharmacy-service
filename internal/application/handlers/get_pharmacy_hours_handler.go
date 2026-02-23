// internal/application/handlers/get_pharmacy_hours_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/application/queries"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"go.uber.org/zap"
)

type GetPharmacyHoursHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	hoursRepo    repositories.PharmacyHoursRepository
	logger       *zap.Logger
}

func NewGetPharmacyHoursHandler(
	pharmacyRepo repositories.PharmacyRepository,
	hoursRepo repositories.PharmacyHoursRepository,
	logger *zap.Logger,
) *GetPharmacyHoursHandler {
	return &GetPharmacyHoursHandler{
		pharmacyRepo: pharmacyRepo,
		hoursRepo:    hoursRepo,
		logger:       logger,
	}
}

func (h *GetPharmacyHoursHandler) Handle(ctx context.Context, query queries.GetPharmacyHoursQuery) (*common.ApiResponse[responses.HoursListResponse], error) {
	pharmacy, err := h.pharmacyRepo.FindByID(ctx, query.PharmacyID)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.HoursListResponse]("Farmacia no encontrada"), nil
	}

	hours, err := h.hoursRepo.FindByPharmacyID(ctx, query.PharmacyID)
	if err != nil {
		h.logger.Error("Error obteniendo horarios", zap.Error(err))
		return common.InternalServerErrorResponse[responses.HoursListResponse]("Error obteniendo horarios"), nil
	}

	hoursResponses := make([]responses.HoursResponse, len(hours))
	for i, h := range hours {
		hoursResponses[i] = responses.ToHoursResponse(&h)
	}

	return common.OkResponse(responses.HoursListResponse{Hours: hoursResponses}), nil
}

var _ mediator.RequestHandler[queries.GetPharmacyHoursQuery, responses.HoursListResponse] = (*GetPharmacyHoursHandler)(nil)
