// internal/application/handlers/list_pharmacies_handler.go
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

type ListPharmaciesHandler struct {
	pharmacyRepo repositories.PharmacyRepository
	logger       *zap.Logger
}

func NewListPharmaciesHandler(pharmacyRepo repositories.PharmacyRepository, logger *zap.Logger) *ListPharmaciesHandler {
	return &ListPharmaciesHandler{pharmacyRepo: pharmacyRepo, logger: logger}
}

func (h *ListPharmaciesHandler) Handle(ctx context.Context, query queries.ListPharmaciesQuery) (*common.ApiResponse[responses.PharmacyListResponse], error) {
	page := query.Page
	if page < 1 {
		page = 1
	}
	limit := query.Limit
	if limit < 1 || limit > 100 {
		limit = 20
	}

	result, err := h.pharmacyRepo.FindAll(ctx, page, limit)
	if err != nil {
		h.logger.Error("Error listando farmacias", zap.Error(err))
		return common.InternalServerErrorResponse[responses.PharmacyListResponse]("Error listando farmacias"), nil
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

	return common.OkResponse(response), nil
}

var _ mediator.RequestHandler[queries.ListPharmaciesQuery, responses.PharmacyListResponse] = (*ListPharmaciesHandler)(nil)
