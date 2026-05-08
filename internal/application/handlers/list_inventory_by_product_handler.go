// internal/application/handlers/list_inventory_by_product_handler.go
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/application/queries"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"go.uber.org/zap"
)

type ListInventoryByProductHandler struct {
	inventoryRepo repositories.InventoryRepository
	cacheService  services.CacheService
	logger        *zap.Logger
}

func NewListInventoryByProductHandler(
	inventoryRepo repositories.InventoryRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *ListInventoryByProductHandler {
	return &ListInventoryByProductHandler{
		inventoryRepo: inventoryRepo,
		cacheService:  cacheService,
		logger:        logger,
	}
}

func (h *ListInventoryByProductHandler) Handle(ctx context.Context, query queries.ListInventoryByProductQuery) (*common.ApiResponse[responses.InventoryListResponse], error) {
	if query.ProductID == "" {
		return common.BadRequestResponse[responses.InventoryListResponse]("VAL_001", "product_id es requerido"), nil
	}

	geo := repositories.GeoFilter{Lat: query.Latitude, Lng: query.Longitude, RadiusKm: query.RadiusKm}
	cacheKey := buildInventoryCacheKey(query.ProductID, geo)
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var inventoryResp responses.InventoryListResponse
		if err := json.Unmarshal([]byte(cached), &inventoryResp); err == nil {
			return common.OkResponse(inventoryResp), nil
		}
	}

	items, err := h.inventoryRepo.FindByProductID(ctx, query.ProductID, geo)
	if err != nil {
		h.logger.Error("Error listando inventario por producto", zap.Error(err))
		return common.InternalServerErrorResponse[responses.InventoryListResponse]("Error listando inventario"), nil
	}

	itemResponses := make([]responses.InventoryItemResponse, len(items))
	for i, item := range items {
		// HU-016 — la regla de dominio decide si "está caro" basándose en
		// los datos brutos del repository (avg + count). El SQL no aplica
		// el threshold; eso vive en `services.IsOverpriced`.
		overpriced := services.IsOverpriced(item.Price, item.DistrictAvgPrice, item.DistrictPharmacyCount)
		var overpricePct *float64
		if overpriced {
			overpricePct = services.OverpricePercentage(item.Price, item.DistrictAvgPrice)
		}

		itemResponses[i] = responses.InventoryItemResponse{
			ID:               item.ID,
			PharmacyID:       item.PharmacyID,
			PharmacySlug:     item.PharmacySlug,
			PharmacyName:     item.PharmacyName,
			PharmacyDistrict: item.District,
			PharmacyAddress:  item.Address,
			ProductID:        item.ProductID,
			Stock:            item.Stock,
			Price:            item.Price,
			IsAvailable:      item.IsAvailable,
			DistanceKm:       item.DistanceKm,
			DistrictAvgPrice: item.DistrictAvgPrice,
			IsOverpriced:     overpriced,
			OverpricePct:     overpricePct,
			CreatedAt:        item.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:        item.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	response := responses.InventoryListResponse{
		Items: itemResponses,
		Total: len(itemResponses),
	}

	if data, err := json.Marshal(response); err == nil {
		go func() {
			_ = h.cacheService.Set(context.Background(), cacheKey, string(data), 5*time.Minute)
		}()
	}

	return common.OkResponse(response), nil
}

// buildInventoryCacheKey genera la clave Redis con bucket geo a 3 decimales
// (~110m de precisión) para evitar explosión de claves cuando el browser
// reporta coordenadas con jitter mínimo. Sin geo activa → clave legacy.
func buildInventoryCacheKey(productID string, geo repositories.GeoFilter) string {
	if !geo.IsActive() {
		return "cache:pharmacy:inventory:product:" + productID
	}
	return fmt.Sprintf("cache:pharmacy:inventory:product:%s:geo:%.3f:%.3f:%.1f",
		productID, geo.Lat, geo.Lng, geo.RadiusKm)
}

var _ mediator.RequestHandler[queries.ListInventoryByProductQuery, responses.InventoryListResponse] = (*ListInventoryByProductHandler)(nil)
