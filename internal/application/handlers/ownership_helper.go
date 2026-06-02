package handlers

import (
	"context"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/pkg/mediator"
)

// OwnershipResult discrimina los 3 outcomes de la verificación de ownership.
type OwnershipResult int

const (
	OwnershipOK OwnershipResult = iota
	OwnershipNotFound
	OwnershipForbidden
)

// assertPharmacyOwnership verifica que el usuario en el contexto sea el dueño
// de la farmacia identificada por pharmacyID.
//
// Devuelve la entidad pharmacy + un OwnershipResult. Si OK, el handler puede
// reusar la pharmacy devuelta para evitar un segundo FindByID.
//
// Casos:
//   - userID ausente en ctx → Forbidden (defensa en profundidad por si el middleware falla)
//   - pharmacy no existe / soft-deleted → NotFound
//   - pharmacy.OwnerUserID nil (scrapeada sin reclamar) → Forbidden
//   - pharmacy.OwnerUserID != userID → Forbidden
//
// El gateway/middleware ya hace el primer gate de "role==pharmacy_owner".
// Este helper agrega el gate "y además, esta farmacia te pertenece".
func assertPharmacyOwnership(
	ctx context.Context,
	repo repositories.PharmacyRepository,
	pharmacyID string,
) (*entities.Pharmacy, OwnershipResult) {
	userID, ok := mediator.GetUserID(ctx)
	if !ok || userID == "" {
		return nil, OwnershipForbidden
	}
	pharmacy, err := repo.FindByID(ctx, pharmacyID)
	if err != nil || pharmacy == nil {
		return nil, OwnershipNotFound
	}
	if pharmacy.OwnerUserID == nil || *pharmacy.OwnerUserID != userID {
		return pharmacy, OwnershipForbidden
	}
	return pharmacy, OwnershipOK
}
