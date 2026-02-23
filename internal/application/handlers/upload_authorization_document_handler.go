// internal/application/handlers/upload_authorization_document_handler.go
package handlers

import (
	"context"
	"fmt"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/domain/events"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/domain/services"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/internal/shared/common"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxDocumentSize = 10 * 1024 * 1024 // 10MB
)

var allowedDocumentTypes = map[string]bool{
	"application/pdf": true,
}

type UploadAuthorizationDocumentHandler struct {
	pharmacyRepo   repositories.PharmacyRepository
	fileStorage    services.FileStorage
	eventPublisher services.EventPublisher
	cacheService   services.CacheService
	bucket         string
	logger         *zap.Logger
}

func NewUploadAuthorizationDocumentHandler(
	pharmacyRepo repositories.PharmacyRepository,
	fileStorage services.FileStorage,
	eventPublisher services.EventPublisher,
	cacheService services.CacheService,
	bucket string,
	logger *zap.Logger,
) *UploadAuthorizationDocumentHandler {
	return &UploadAuthorizationDocumentHandler{
		pharmacyRepo:   pharmacyRepo,
		fileStorage:    fileStorage,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
		bucket:         bucket,
		logger:         logger,
	}
}

func (h *UploadAuthorizationDocumentHandler) Handle(ctx context.Context, cmd commands.UploadAuthorizationDocumentCommand) (*common.ApiResponse[responses.AuthorizationDocumentResponse], error) {
	// Verify pharmacy exists
	pharmacy, err := h.pharmacyRepo.FindByID(ctx, cmd.PharmacyID)
	if err != nil || pharmacy == nil {
		return common.NotFoundResponse[responses.AuthorizationDocumentResponse]("Farmacia no encontrada"), nil
	}

	// Validate file type
	if !allowedDocumentTypes[cmd.ContentType] {
		return common.BadRequestResponse[responses.AuthorizationDocumentResponse]("VAL_004", "Tipo de archivo no permitido. Solo se acepta PDF"), nil
	}

	// Validate file size
	if cmd.Size > maxDocumentSize {
		return common.BadRequestResponse[responses.AuthorizationDocumentResponse]("VAL_005", "El archivo excede el tamaño máximo de 10MB"), nil
	}

	// Upload to S3
	key := fmt.Sprintf("pharmacies/%s/authorization/%s.pdf", cmd.PharmacyID, uuid.New().String())
	url, err := h.fileStorage.Upload(ctx, h.bucket, key, cmd.Reader, cmd.ContentType)
	if err != nil {
		h.logger.Error("Error subiendo documento a S3", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AuthorizationDocumentResponse]("Error subiendo documento"), nil
	}

	// Update pharmacy record
	pharmacy.AuthorizationDocumentURL = url
	if err := h.pharmacyRepo.Update(ctx, pharmacy); err != nil {
		h.logger.Error("Error actualizando farmacia con documento", zap.Error(err))
		return common.InternalServerErrorResponse[responses.AuthorizationDocumentResponse]("Error actualizando farmacia"), nil
	}

	// Invalidate cache
	go func() {
		_ = h.cacheService.Delete(context.Background(), "cache:pharmacy:pharmacy:"+cmd.PharmacyID)
	}()

	// Publish event
	go func() {
		event := events.NewPharmacyEvent(events.EventAuthorizationUploaded, cmd.PharmacyID, pharmacy.OwnerUserID)
		event.Metadata["document_url"] = url
		if err := h.eventPublisher.Publish(context.Background(), event); err != nil {
			h.logger.Error("Error publicando evento AUTHORIZATION_UPLOADED", zap.Error(err))
		}
	}()

	h.logger.Info("Documento de autorización subido exitosamente",
		zap.String("pharmacy_id", cmd.PharmacyID),
		zap.String("document_url", url),
	)

	return common.OkResponse(responses.AuthorizationDocumentResponse{
		PharmacyID:  cmd.PharmacyID,
		DocumentURL: url,
	}), nil
}

var _ mediator.RequestHandler[commands.UploadAuthorizationDocumentCommand, responses.AuthorizationDocumentResponse] = (*UploadAuthorizationDocumentHandler)(nil)
