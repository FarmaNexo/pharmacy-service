// internal/application/preprocessors/sanitize_input_preprocessor.go
package preprocessors

import (
	"context"
	"strings"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"go.uber.org/zap"
)

type SanitizeInputPreProcessor struct {
	logger *zap.Logger
}

func NewSanitizeInputPreProcessor(logger *zap.Logger) *SanitizeInputPreProcessor {
	return &SanitizeInputPreProcessor{logger: logger}
}

func (p *SanitizeInputPreProcessor) Process(ctx context.Context, request interface{}) error {
	switch cmd := request.(type) {
	case *commands.CreatePharmacyCommand:
		cmd.Name = strings.TrimSpace(cmd.Name)
		cmd.Description = strings.TrimSpace(cmd.Description)
		cmd.Phone = strings.TrimSpace(cmd.Phone)
		cmd.Email = strings.TrimSpace(cmd.Email)
		cmd.Website = strings.TrimSpace(cmd.Website)
		cmd.Street = strings.TrimSpace(cmd.Street)
		cmd.City = strings.TrimSpace(cmd.City)
		cmd.State = strings.TrimSpace(cmd.State)
		cmd.PostalCode = strings.TrimSpace(cmd.PostalCode)
		p.logger.Debug("Input sanitizado", zap.String("command", "CreatePharmacyCommand"))

	case *commands.UpdatePharmacyCommand:
		cmd.Name = strings.TrimSpace(cmd.Name)
		cmd.Description = strings.TrimSpace(cmd.Description)
		cmd.Phone = strings.TrimSpace(cmd.Phone)
		cmd.Email = strings.TrimSpace(cmd.Email)
		cmd.Website = strings.TrimSpace(cmd.Website)
		cmd.Street = strings.TrimSpace(cmd.Street)
		cmd.City = strings.TrimSpace(cmd.City)
		cmd.State = strings.TrimSpace(cmd.State)
		cmd.PostalCode = strings.TrimSpace(cmd.PostalCode)
		p.logger.Debug("Input sanitizado", zap.String("command", "UpdatePharmacyCommand"))
	}

	return nil
}
