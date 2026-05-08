// internal/application/handlers/get_pharmacy_by_slug_handler.go
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

type GetPharmacyBySlugHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	hoursRepo    repositories.PharmacyHoursRepository
	cacheService services.CacheService
	logger       *zap.Logger
}

func NewGetPharmacyBySlugHandler(
	pharmacyRepo repositories.PharmacyRepository,
	hoursRepo repositories.PharmacyHoursRepository,
	cacheService services.CacheService,
	logger *zap.Logger,
) *GetPharmacyBySlugHandler {
	return &GetPharmacyBySlugHandler{
		pharmacyRepo: pharmacyRepo,
		hoursRepo:    hoursRepo,
		cacheService: cacheService,
		logger:       logger,
	}
}

func (h *GetPharmacyBySlugHandler) Handle(ctx context.Context, query queries.GetPharmacyBySlugQuery) (*common.ApiResponse[responses.PharmacyResponse], error) {
	cacheKey := "cache:pharmacy:slug:" + query.Slug
	if cached, err := h.cacheService.Get(ctx, cacheKey); err == nil && cached != "" {
		var pharmacyResp responses.PharmacyResponse
		if err := json.Unmarshal([]byte(cached), &pharmacyResp); err == nil {
			h.logger.Debug("Farmacia obtenida de caché por slug", zap.String("slug", query.Slug))
			return common.OkResponse(pharmacyResp), nil
		}
	}

	pharmacy, err := h.pharmacyRepo.FindBySlug(ctx, query.Slug)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.PharmacyResponse]("Farmacia no encontrada"), nil
	}

	hours, _ := h.hoursRepo.FindByPharmacyID(ctx, pharmacy.ID)
	pharmacy.Hours = hours

	pharmacyResp := responses.ToPharmacyResponse(pharmacy)

	if data, err := json.Marshal(pharmacyResp); err == nil {
		go func() {
			_ = h.cacheService.Set(context.Background(), cacheKey, string(data), 1*time.Hour)
		}()
	}

	return common.OkResponse(pharmacyResp), nil
}

var _ mediator.RequestHandler[queries.GetPharmacyBySlugQuery, responses.PharmacyResponse] = (*GetPharmacyBySlugHandler)(nil)
