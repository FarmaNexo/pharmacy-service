// internal/application/handlers/update_inventory_item_handler.go
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

type UpdateInventoryItemHandler struct {
	pharmacyRepo   repositories.PharmacyRepository
	inventoryRepo  repositories.InventoryRepository
	eventPublisher services.EventPublisher
	cacheService   services.CacheService
	logger         *zap.Logger
}

func NewUpdateInventoryItemHandler(
	pharmacyRepo repositories.PharmacyRepository,
	inventoryRepo repositories.InventoryRepository,
	eventPublisher services.EventPublisher,
	cacheService services.CacheService,
	logger *zap.Logger,
) *UpdateInventoryItemHandler {
	return &UpdateInventoryItemHandler{
		pharmacyRepo:   pharmacyRepo,
		inventoryRepo:  inventoryRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
		logger:         logger,
	}
}

func (h *UpdateInventoryItemHandler) Handle(ctx context.Context, cmd commands.UpdateInventoryItemCommand) (*common.ApiResponse[responses.InventoryItemResponse], error) {
	if _, owns := assertPharmacyOwnership(ctx, h.pharmacyRepo, cmd.PharmacyID); owns != OwnershipOK {
		if owns == OwnershipNotFound {
			return common.NotFoundResponse[responses.InventoryItemResponse]("Farmacia no encontrada"), nil
		}
		return common.ForbiddenResponse[responses.InventoryItemResponse]("No tienes permiso para gestionar el inventario de esta farmacia"), nil
	}

	item, err := h.inventoryRepo.FindByPharmacyAndProduct(ctx, cmd.PharmacyID, cmd.ProductID)
	if err != nil || item == nil {
		return common.NotFoundResponse[responses.InventoryItemResponse]("Producto no encontrado en inventario"), nil
	}

	item.Stock = cmd.Stock
	item.Price = cmd.Price
	item.IsAvailable = cmd.IsAvailable

	if err := h.inventoryRepo.Update(ctx, item); err != nil {
		h.logger.Error("Error actualizando inventario", zap.Error(err))
		return common.InternalServerErrorResponse[responses.InventoryItemResponse]("Error actualizando inventario"), nil
	}

	go func() {
		_ = h.cacheService.Delete(context.Background(), "cache:pharmacy:inventory:"+cmd.PharmacyID)
	}()

	go func() {
		event := events.NewPharmacyEvent(events.EventInventoryUpdated, cmd.PharmacyID, "")
		event.ProductID = cmd.ProductID
		if err := h.eventPublisher.Publish(context.Background(), event); err != nil {
			h.logger.Error("Error publicando evento INVENTORY_UPDATED", zap.Error(err))
		}
	}()

	h.logger.Info("Inventario actualizado",
		zap.String("pharmacy_id", cmd.PharmacyID),
		zap.String("product_id", cmd.ProductID),
	)

	return common.OkResponse(responses.ToInventoryItemResponse(item)), nil
}

var _ mediator.RequestHandler[commands.UpdateInventoryItemCommand, responses.InventoryItemResponse] = (*UpdateInventoryItemHandler)(nil)
