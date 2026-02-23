// internal/application/validators/add_inventory_item_validator.go
package validators

import (
	"context"
	"fmt"
	"strings"

	"github.com/farmanexo/pharmacy-service/internal/application/commands"
	"github.com/farmanexo/pharmacy-service/internal/presentation/dto/responses"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
)

type AddInventoryItemValidator struct{}

func NewAddInventoryItemValidator() *AddInventoryItemValidator {
	return &AddInventoryItemValidator{}
}

func (v *AddInventoryItemValidator) Validate(ctx context.Context, cmd commands.AddInventoryItemCommand) error {
	var errors []string

	if strings.TrimSpace(cmd.ProductID) == "" {
		errors = append(errors, "El product_id es requerido")
	}
	if cmd.Price <= 0 {
		errors = append(errors, "El precio debe ser mayor a 0")
	}
	if cmd.Stock < 0 {
		errors = append(errors, "El stock debe ser mayor o igual a 0")
	}

	if len(errors) > 0 {
		return fmt.Errorf("errores de validación: %s", strings.Join(errors, "; "))
	}

	return nil
}

var _ mediator.Validator[commands.AddInventoryItemCommand, responses.InventoryItemResponse] = (*AddInventoryItemValidator)(nil)
