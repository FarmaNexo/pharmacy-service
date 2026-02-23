// internal/application/validators/create_pharmacy_validator.go
package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
)

type CreatePharmacyValidator struct{}

func NewCreatePharmacyValidator() *CreatePharmacyValidator {
	return &CreatePharmacyValidator{}
}

func (v *CreatePharmacyValidator) Validate(ctx context.Context, cmd commands.CreatePharmacyCommand) error {
	var errors []string

	if strings.TrimSpace(cmd.Name) == "" {
		errors = append(errors, "El nombre de la farmacia es requerido")
	}
	if len(cmd.Name) > 255 {
		errors = append(errors, "El nombre no puede exceder 255 caracteres")
	}
	if strings.TrimSpace(cmd.Street) == "" {
		errors = append(errors, "La dirección es requerida")
	}
	if strings.TrimSpace(cmd.City) == "" {
		errors = append(errors, "La ciudad es requerida")
	}
	if cmd.Latitude < -90 || cmd.Latitude > 90 {
		errors = append(errors, "Latitud inválida (debe estar entre -90 y 90)")
	}
	if cmd.Longitude < -180 || cmd.Longitude > 180 {
		errors = append(errors, "Longitud inválida (debe estar entre -180 y 180)")
	}
	if cmd.Latitude == 0 && cmd.Longitude == 0 {
		errors = append(errors, "Las coordenadas son requeridas")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}

	return nil
}

var _ mediator.Validator[commands.CreatePharmacyCommand, responses.PharmacyResponse] = (*CreatePharmacyValidator)(nil)
