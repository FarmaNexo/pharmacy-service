// internal/application/handlers/remove_inventory_item_handler.go
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

type RemoveInventoryItemHandler struct {
	inventoryRepo repositories.InventoryRepository
	cacheService  services.CacheService
	logger        *zap.Logger
}

func NewRemoveInventoryItemHandler(
	inventoryRepo repositories.InventoryRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *RemoveInventoryItemHandler {
	return &RemoveInventoryItemHandler{
		inventoryRepo: inventoryRepo,
		cacheService:  cacheService,
		logger:        logger,
	}
}

func (h *RemoveInventoryItemHandler) Handle(ctx context.Context, cmd commands.RemoveInventoryItemCommand) (*common.ApiResponse[responses.EmptyResponse], error) {
	item, err := h.inventoryRepo.FindByPharmacyAndProduct(ctx, cmd.PharmacyID, cmd.ProductID)
	if err != nil || item == nil {
		return common.NotFoundResponse[responses.EmptyResponse]("Producto no encontrado en inventario"), nil
	}

	if err := h.inventoryRepo.Delete(ctx, cmd.PharmacyID, cmd.ProductID); err != nil {
		h.logger.Error("Error removiendo del inventario", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse]("Error removiendo del inventario"), nil
	}

	go func() {
		_ = h.cacheService.Delete(context.Background(), "cache:pharmacy:inventory:"+cmd.PharmacyID)
	}()

	h.logger.Info("Producto removido del inventario",
		zap.String("pharmacy_id", cmd.PharmacyID),
		zap.String("product_id", cmd.ProductID),
	)

	resp := common.NewApiResponse[responses.EmptyResponse]()
	resp.SetHttpStatus(constants.StatusOK.Int())
	resp.AddMessageWithType(constants.CodeInventoryRemoved, constants.GetDescription(constants.CodeInventoryRemoved), constants.MessageTypeSuccess)
	return resp, nil
}

var _ mediator.RequestHandler[commands.RemoveInventoryItemCommand, responses.EmptyResponse] = (*RemoveInventoryItemHandler)(nil)
