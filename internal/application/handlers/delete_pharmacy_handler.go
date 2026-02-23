// internal/application/handlers/delete_pharmacy_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/internal/shared/constants"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"go.uber.org/zap"
)

type DeletePharmacyHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewDeletePharmacyHandler(
	pharmacyRepo repositories.PharmacyRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *DeletePharmacyHandler {
	return &DeletePharmacyHandler{
		pharmacyRepo: pharmacyRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *DeletePharmacyHandler) Handle(ctx context.Context, cmd commands.DeletePharmacyCommand) (*common.ApiResponse[responses.EmptyResponse], error) {
	pharmacy, err := h.pharmacyRepo.FindByID(ctx, cmd.ID)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.EmptyResponse]("Farmacia no encontrada"), nil
	}

	if err := h.pharmacyRepo.SoftDelete(ctx, cmd.ID); err != nil {
		h.logger.Error("Error eliminando farmacia", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse]("Error eliminando farmacia"), nil
	}

	go func() {
		_ = h.cacheService.Delete(context.Background(), "cache:pharmacy:pharmacy:"+cmd.ID)
		_ = h.cacheService.DeleteByPattern(context.Background(), "cache:pharmacy:nearby:*")
	}()

	h.logger.Info("Farmacia eliminada (soft delete)", zap.String("pharmacy_id", cmd.ID))

	resp := common.NewApiResponse[responses.EmptyResponse]()
	resp.SetHttpStatus(constants.StatusOK.Int())
	resp.AddMessageWithType(constants.CodeDeletedSuccess, constants.GetDescription(constants.CodeDeletedSuccess), constants.MessageTypeSuccess)
	return resp, nil
}

var _ mediator.RequestHandler[commands.DeletePharmacyCommand, responses.EmptyResponse] = (*DeletePharmacyHandler)(nil)
