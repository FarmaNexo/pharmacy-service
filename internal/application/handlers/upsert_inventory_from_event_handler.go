// internal/application/handlers/upsert_inventory_from_event_handler.go
package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/farmanexo/pharmacy-service/internal/domain/events"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"github.com/farmanexo/pharmacy-service/internal/infrastructure/clients"
	"go.uber.org/zap"
)

// ErrFailSoft indica al consumer SQS que NO debe borrar el mensaje:
// SQS lo redrive tras visibility timeout (3 reintentos antes de DLQ).
// Se usa cuando faltan dependencias (pharmacy o product no UPSERTeados aún).
var ErrFailSoft = errors.New("dependencia no resuelta — fail soft para redrive SQS")

// UpsertInventoryFromEventHandler procesa INVENTORY_DISCOVERED.
// Flujo:
//  1. Lookup pharmacy local por source_pharmacy_code → pharmacy_id.
//  2. Lookup catalog HTTP por (source_product_code, concentration) → product_id.
//  3. UPSERT por (pharmacy_id, product_id).
//
// Si 1 o 2 fallan por "no encontrado", retorna ErrFailSoft (orden de eventos:
// el INVENTORY puede haber llegado antes que su PHARMACY o PRODUCT).
type UpsertInventoryFromEventHandler struct {
	pharmacyRepo  repositories.PharmacyRepository
	inventoryRepo repositories.InventoryRepository
	catalogClient clients.CatalogClient
	logger        *zap.Logger
}

func NewUpsertInventoryFromEventHandler(
	pharmacyRepo repositories.PharmacyRepository,
	inventoryRepo repositories.InventoryRepository,
	catalogClient clients.CatalogClient,
	logger *zap.Logger,
) *UpsertInventoryFromEventHandler {
	return &UpsertInventoryFromEventHandler{
		pharmacyRepo:  pharmacyRepo,
		inventoryRepo: inventoryRepo,
		catalogClient: catalogClient,
		logger:        logger,
	}
}

func (h *UpsertInventoryFromEventHandler) Handle(ctx context.Context, data events.InventoryDiscoveredData) error {
	// 1. Lookup pharmacy local
	pharmacy, err := h.pharmacyRepo.FindBySourceCode(ctx, data.SourcePharmacyCode)
	if err != nil {
		return fmt.Errorf("lookup pharmacy: %w", err)
	}
	if pharmacy == nil {
		h.logger.Debug("Inventory recibido antes de su pharmacy — fail soft",
			zap.String("source_pharmacy_code", data.SourcePharmacyCode),
			zap.Int("source_product_code", data.SourceProductCode),
		)
		return ErrFailSoft
	}

	// 2. Lookup product via catalog HTTP
	productID, err := h.catalogClient.FindProductIDBySource(ctx, data.SourceProductCode, data.Concentration)
	if err != nil {
		if errors.Is(err, clients.ErrProductNotFound) {
			h.logger.Debug("Inventory recibido antes de su product — fail soft",
				zap.Int("source_product_code", data.SourceProductCode),
				zap.String("concentration", data.Concentration),
			)
			return ErrFailSoft
		}
		return fmt.Errorf("lookup catalog product: %w", err)
	}

	// 3. UPSERT inventory
	stock := data.Stock
	if stock <= 0 {
		stock = 1 // Convención: DIGEMID no expone stock real.
	}
	if err := h.inventoryRepo.UpsertByPharmacyAndProduct(ctx, repositories.InventoryUpsertParams{
		PharmacyID:  pharmacy.ID,
		ProductID:   productID,
		Stock:       stock,
		Price:       data.Price,
		IsAvailable: data.IsAvailable,
	}); err != nil {
		return fmt.Errorf("upsert inventory: %w", err)
	}

	h.logger.Info("Inventory upserteado desde scraper",
		zap.String("pharmacy_id", pharmacy.ID),
		zap.String("product_id", productID),
		zap.Float64("price", data.Price),
	)
	return nil
}
