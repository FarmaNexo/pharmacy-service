// internal/application/handlers/get_inventory_item_handler.go
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

// GetInventoryItemHandler resuelve un item de inventario por
// (pharmacy_id, product_id). Lo consume order-service para validar
// stock antes de cart-add y checkout. Devuelve 404 si la farmacia no
// carga ese producto — caso esperado, no es error.
type GetInventoryItemHandler struct {
	pharmacyRepo  repositories.PharmacyRepository
	inventoryRepo repositories.InventoryRepository
	logger        *zap.Logger
}

func NewGetInventoryItemHandler(
	pharmacyRepo repositories.PharmacyRepository,
	inventoryRepo repositories.InventoryRepository,
	logger *zap.Logger,
) *GetInventoryItemHandler {
	return &GetInventoryItemHandler{
		pharmacyRepo:  pharmacyRepo,
		inventoryRepo: inventoryRepo,
		logger:        logger,
	}
}

func (h *GetInventoryItemHandler) Handle(ctx context.Context, query queries.GetInventoryItemQuery) (*common.ApiResponse[responses.InventoryItemResponse], error) {
	item, err := h.inventoryRepo.FindByPharmacyAndProduct(ctx, query.PharmacyID, query.ProductID)
	if err != nil {
		h.logger.Error("Error consultando inventario",
			zap.String("pharmacy_id", query.PharmacyID),
			zap.String("product_id", query.ProductID),
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.InventoryItemResponse]("Error consultando inventario"), nil
	}
	if item == nil {
		return common.NotFoundResponse[responses.InventoryItemResponse]("Inventario no encontrado para esa farmacia y producto"), nil
	}

	pharmacy, err := h.pharmacyRepo.FindByID(ctx, query.PharmacyID)
	if err != nil || pharmacy == nil {
		// Caso anómalo: el inventory tiene FK lógica a una farmacia que no
		// existe. Devolvemos el item igual con datos de pharmacy vacíos para
		// no romper al consumidor (order-service), pero loggeamos como warn.
		h.logger.Warn("Inventory huérfano: pharmacy no encontrada",
			zap.String("pharmacy_id", query.PharmacyID),
			zap.String("inventory_id", item.ID),
		)
		return common.OkResponse(responses.ToInventoryItemResponse(item)), nil
	}

	return common.OkResponse(responses.ToInventoryItemResponseWithPharmacy(item, pharmacy)), nil
}

var _ mediator.RequestHandler[queries.GetInventoryItemQuery, responses.InventoryItemResponse] = (*GetInventoryItemHandler)(nil)
