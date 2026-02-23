// internal/application/postprocessors/log_audit_postprocessor.go
package postprocessors

import (
	"context"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/application/queries"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
	"go.uber.org/zap"
)

type LogAuditPostProcessor struct {
	logger *zap.Logger
}

func NewLogAuditPostProcessor(logger *zap.Logger) *LogAuditPostProcessor {
	return &LogAuditPostProcessor{logger: logger}
}

func (p *LogAuditPostProcessor) Process(ctx context.Context, request interface{}, response interface{}) error {
	userID := p.getUserIDFromContext(ctx)
	correlationID := mediator.GetCorrelationID(ctx)
	isSuccess := p.checkSuccess(response)

	switch request.(type) {
	case commands.CreatePharmacyCommand, *commands.CreatePharmacyCommand:
		p.logAudit("PHARMACY_REGISTERED", userID, correlationID, isSuccess)
	case commands.UpdatePharmacyCommand, *commands.UpdatePharmacyCommand:
		p.logAudit("PHARMACY_UPDATED", userID, correlationID, isSuccess)
	case commands.DeletePharmacyCommand, *commands.DeletePharmacyCommand:
		p.logAudit("PHARMACY_DELETED", userID, correlationID, isSuccess)
	case commands.VerifyPharmacyCommand, *commands.VerifyPharmacyCommand:
		p.logAudit("PHARMACY_VERIFIED", userID, correlationID, isSuccess)
	case commands.AddInventoryItemCommand, *commands.AddInventoryItemCommand:
		p.logAudit("INVENTORY_ITEM_ADDED", userID, correlationID, isSuccess)
	case commands.UpdateInventoryItemCommand, *commands.UpdateInventoryItemCommand:
		p.logAudit("INVENTORY_ITEM_UPDATED", userID, correlationID, isSuccess)
	case commands.RemoveInventoryItemCommand, *commands.RemoveInventoryItemCommand:
		p.logAudit("INVENTORY_ITEM_REMOVED", userID, correlationID, isSuccess)
	case commands.UpdatePharmacyHoursCommand, *commands.UpdatePharmacyHoursCommand:
		p.logAudit("PHARMACY_HOURS_UPDATED", userID, correlationID, isSuccess)
	case queries.SearchNearbyPharmaciesQuery, *queries.SearchNearbyPharmaciesQuery:
		p.logAudit("NEARBY_PHARMACIES_SEARCHED", userID, correlationID, isSuccess)
	default:
		p.logger.Debug("Post-processor: comando sin auditoría configurada")
	}

	return nil
}

func (p *LogAuditPostProcessor) logAudit(eventType, userID, correlationID string, success bool) {
	p.logger.Info("AUDIT",
		zap.String("event_type", eventType),
		zap.Bool("success", success),
		zap.String("correlation_id", correlationID),
		zap.String("user_id", userID),
		zap.Time("timestamp", time.Now()),
	)
}

func (p *LogAuditPostProcessor) checkSuccess(response interface{}) bool {
	if resp, ok := response.(interface{ IsValid() bool }); ok {
		return resp.IsValid()
	}
	return false
}

func (p *LogAuditPostProcessor) getUserIDFromContext(ctx context.Context) string {
	userID, _ := mediator.GetUserID(ctx)
	if userID == "" {
		return "ANONYMOUS"
	}
	return userID
}
