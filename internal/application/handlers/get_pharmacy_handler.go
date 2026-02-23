// internal/application/handlers/get_pharmacy_handler.go
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

type GetPharmacyHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	hoursRepo    repositories.PharmacyHoursRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewGetPharmacyHandler(
	pharmacyRepo repositories.PharmacyRepository,
	hoursRepo repositories.PharmacyHoursRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *GetPharmacyHandler {
	return &GetPharmacyHandler{
		pharmacyRepo: pharmacyRepo,
		hoursRepo:    hoursRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *GetPharmacyHandler) Handle(ctx context.Context, query queries.GetPharmacyQuery) (*common.ApiResponse[responses.PharmacyResponse], error) {
	cacheKey := "cache:pharmacy:pharmacy:" + query.ID
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var pharmacyResp responses.PharmacyResponse
		if err := json.Unmarshal([]byte(cached), &pharmacyResp); err == nil {
			h.logger.Debug("Farmacia obtenida de caché", zap.String("pharmacy_id", query.ID))
			return common.OkResponse(pharmacyResp), nil
		}
	}

	pharmacy, err := h.pharmacyRepo.FindByID(ctx, query.ID)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.PharmacyResponse]("Farmacia no encontrada"), nil
	}

	hours, _ := h.hoursRepo.FindByPharmacyID(ctx, query.ID)
	pharmacy.Hours = hours

	pharmacyResp := responses.ToPharmacyResponse(pharmacy)

	if data, err := json.Marshal(pharmacyResp); err == nil {
		go func() {
			_ = h.cacheService.Set(context.Background(), cacheKey, string(data), 1*time.Hour)
		}()
	}

	return common.OkResponse(pharmacyResp), nil
}

var _ mediator.RequestHandler[queries.GetPharmacyQuery, responses.PharmacyResponse] = (*GetPharmacyHandler)(nil)
