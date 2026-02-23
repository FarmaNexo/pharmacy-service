// internal/application/handlers/list_pharmacies_by_chain_handler.go
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

type ListPharmaciesByChainHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewListPharmaciesByChainHandler(
	pharmacyRepo repositories.PharmacyRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *ListPharmaciesByChainHandler {
	return &ListPharmaciesByChainHandler{
		pharmacyRepo: pharmacyRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *ListPharmaciesByChainHandler) Handle(ctx context.Context, query queries.ListPharmaciesByChainQuery) (*common.ApiResponse[responses.PharmacyListResponse], error) {
	if query.ChainID == "" {
		return common.BadRequestResponse[responses.PharmacyListResponse]("VAL_001", "Chain ID es requerido"), nil
	}

	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Try cache first
	cacheKey := fmt.Sprintf("cache:pharmacy:chain:%s:page:%d:limit:%d", query.ChainID, page, limit)
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var listResp responses.PharmacyListResponse
		if err := json.Unmarshal([]byte(cached), &listResp); err == nil {
			h.logger.Debug("Farmacias por cadena obtenidas de caché", zap.String("chain_id", query.ChainID))
			return common.OkResponse(listResp), nil
		}
	}

	result, err := h.pharmacyRepo.FindByChainID(ctx, query.ChainID, page, limit)
	if err != nil {
		h.logger.Error("Error listando farmacias por cadena", zap.Error(err), zap.String("chain_id", query.ChainID))
		return common.InternalServerErrorResponse[responses.PharmacyListResponse]("Error listando farmacias por cadena"), nil
	}

	pharmacyResponses := make([]responses.PharmacyResponse, len(result.Items))
	for i, p := range result.Items {
		pharmacyResponses[i] = responses.ToPharmacyResponse(p)
	}

	totalPages := int(result.Total) / limit
	if int(result.Total)%limit > 0 {
		totalPages++
	}

	response := responses.PharmacyListResponse{
		Pharmacies: pharmacyResponses,
		Total:      result.Total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	// Cache for 1 hour
	if data, err := json.Marshal(response); err == nil {
		go func() {
			_ = h.cacheService.Set(context.Background(), cacheKey, string(data), 1*time.Hour)
		}()
	}

	return common.OkResponse(response), nil
}

var _ mediator.RequestHandler[queries.ListPharmaciesByChainQuery, responses.PharmacyListResponse] = (*ListPharmaciesByChainHandler)(nil)
