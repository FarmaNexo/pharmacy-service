// internal/application/handlers/list_pharmacy_inventory_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/application/queries"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"go.uber.org/zap"
)

type ListPharmacyInventoryHandler struct {
	pharmacyRepo  repositories.PharmacyRepository
	inventoryRepo repositories.InventoryRepository
	cacheService  services.CacheService
	logger        *zap.Logger
}

func NewListPharmacyInventoryHandler(
	pharmacyRepo repositories.PharmacyRepository,
	inventoryRepo repositories.InventoryRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *ListPharmacyInventoryHandler {
	return &ListPharmacyInventoryHandler{
		pharmacyRepo:  pharmacyRepo,
		inventoryRepo: inventoryRepo,
		cacheService:  cacheService,
		logger:        logger,
	}
}

func (h *ListPharmacyInventoryHandler) Handle(ctx context.Context, query queries.ListPharmacyInventoryQuery) (*common.ApiResponse[responses.InventoryListResponse], error) {
	pharmacy, err := h.pharmacyRepo.FindByID(ctx, query.PharmacyID)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.InventoryListResponse]("Farmacia no encontrada"), nil
	}

	cacheKey := "cache:pharmacy:inventory:" + query.PharmacyID
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var inventoryResp responses.InventoryListResponse
		if err := json.Unmarshal([]byte(cached), &inventoryResp); err == nil {
			return common.OkResponse(inventoryResp), nil
		}
	}

	items, err := h.inventoryRepo.FindByPharmacyID(ctx, query.PharmacyID)
	if err != nil {
		h.logger.Error("Error listando inventario", zap.Error(err))
		return common.InternalServerErrorResponse[responses.InventoryListResponse]("Error listando inventario"), nil
	}

	itemResponses := make([]responses.InventoryItemResponse, len(items))
	for i, item := range items {
		itemResponses[i] = responses.ToInventoryItemResponse(&item)
	}

	response := responses.InventoryListResponse{
		Items: itemResponses,
		Total: len(itemResponses),
	}

	if data, err := json.Marshal(response); err == nil {
		go func() {
			_ = h.cacheService.Set(context.Background(), cacheKey, string(data), 30*time.Minute)
		}()
	}

	return common.OkResponse(response), nil
}

var _ mediator.RequestHandler[queries.ListPharmacyInventoryQuery, responses.InventoryListResponse] = (*ListPharmacyInventoryHandler)(nil)
