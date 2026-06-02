// internal/application/handlers/add_inventory_item_handler.go
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

type AddInventoryItemHandler struct {
	pharmacyRepo   repositories.PharmacyRepository
	inventoryRepo  repositories.InventoryRepository
	eventPublisher services.EventPublisher
	cacheService   services.CacheService
	logger         *zap.Logger
}

func NewAddInventoryItemHandler(
	pharmacyRepo repositories.PharmacyRepository,
	inventoryRepo repositories.InventoryRepository,
	eventPublisher services.EventPublisher,
	cacheService services.CacheService,
	logger *zap.Logger,
) *AddInventoryItemHandler {
	return &AddInventoryItemHandler{
		pharmacyRepo:   pharmacyRepo,
		inventoryRepo:  inventoryRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
		logger:         logger,
	}
}

func (h *AddInventoryItemHandler) Handle(ctx context.Context, cmd commands.AddInventoryItemCommand) (*common.ApiResponse[responses.InventoryItemResponse], error) {
	if _, owns := assertPharmacyOwnership(ctx, h.pharmacyRepo, cmd.PharmacyID); owns != OwnershipOK {
		if owns == OwnershipNotFound {
			return common.NotFoundResponse[responses.InventoryItemResponse]("Farmacia no encontrada"), nil
		}
		return common.ForbiddenResponse[responses.InventoryItemResponse]("No tienes permiso para gestionar el inventario de esta farmacia"), nil
	}

	existing, _ := h.inventoryRepo.FindByPharmacyAndProduct(ctx, cmd.PharmacyID, cmd.ProductID)
	if existing != nil {
		return common.ConflictResponse[responses.InventoryItemResponse]("BUS_005", "El producto ya existe en el inventario de esta farmacia"), nil
	}

	item := &entities.PharmacyInventory{
		ID:          uuid.New().String(),
		PharmacyID:  cmd.PharmacyID,
		ProductID:   cmd.ProductID,
		Stock:       cmd.Stock,
		Price:       cmd.Price,
		IsAvailable: cmd.IsAvailable,
	}

	if err := h.inventoryRepo.Create(ctx, item); err != nil {
		h.logger.Error("Error agregando al inventario", zap.Error(err))
		return common.InternalServerErrorResponse[responses.InventoryItemResponse]("Error agregando al inventario"), nil
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

	h.logger.Info("Producto agregado al inventario",
		zap.String("pharmacy_id", cmd.PharmacyID),
		zap.String("product_id", cmd.ProductID),
	)

	return common.CreatedResponse(responses.ToInventoryItemResponse(item)), nil
}

var _ mediator.RequestHandler[commands.AddInventoryItemCommand, responses.InventoryItemResponse] = (*AddInventoryItemHandler)(nil)
