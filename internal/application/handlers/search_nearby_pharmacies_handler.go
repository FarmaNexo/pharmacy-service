// internal/application/handlers/search_nearby_pharmacies_handler.go
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

type SearchNearbyPharmaciesHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewSearchNearbyPharmaciesHandler(
	pharmacyRepo repositories.PharmacyRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *SearchNearbyPharmaciesHandler {
	return &SearchNearbyPharmaciesHandler{
		pharmacyRepo: pharmacyRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *SearchNearbyPharmaciesHandler) Handle(ctx context.Context, query queries.SearchNearbyPharmaciesQuery) (*common.ApiResponse[responses.NearbyPharmaciesResponse], error) {
	if query.Latitude < -90 || query.Latitude > 90 || query.Longitude < -180 || query.Longitude > 180 {
		return common.BadRequestResponse[responses.NearbyPharmaciesResponse]("VAL_002", "Coordenadas inválidas"), nil
	}
	if query.RadiusKm <= 0 || query.RadiusKm > 50 {
		return common.BadRequestResponse[responses.NearbyPharmaciesResponse]("VAL_003", "Radio debe ser entre 0.1 y 50 km"), nil
	}

	limit := query.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	cacheKey := fmt.Sprintf("cache:pharmacy:nearby:%.4f:%.4f:%.1f:%d", query.Latitude, query.Longitude, query.RadiusKm, limit)
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var nearbyResp responses.NearbyPharmaciesResponse
		if err := json.Unmarshal([]byte(cached), &nearbyResp); err == nil {
			h.logger.Debug("Farmacias cercanas obtenidas de caché")
			return common.OkResponse(nearbyResp), nil
		}
	}

	pharmacies, err := h.pharmacyRepo.FindNearby(ctx, query.Latitude, query.Longitude, query.RadiusKm, limit)
	if err != nil {
		h.logger.Error("Error buscando farmacias cercanas", zap.Error(err))
		return common.InternalServerErrorResponse[responses.NearbyPharmaciesResponse]("Error buscando farmacias cercanas"), nil
	}

	pharmacyResponses := make([]responses.PharmacyResponse, len(pharmacies))
	for i, p := range pharmacies {
		pharmacyResponses[i] = responses.ToPharmacyResponse(p)
	}

	response := responses.NearbyPharmaciesResponse{
		Pharmacies: pharmacyResponses,
		Total:      len(pharmacyResponses),
		RadiusKm:   query.RadiusKm,
	}

	if data, err := json.Marshal(response); err == nil {
		go func() {
			_ = h.cacheService.Set(context.Background(), cacheKey, string(data), 15*time.Minute)
		}()
	}

	return common.OkResponse(response), nil
}

var _ mediator.RequestHandler[queries.SearchNearbyPharmaciesQuery, responses.NearbyPharmaciesResponse] = (*SearchNearbyPharmaciesHandler)(nil)
