// internal/domain/repositories/inventory_repository.go
package repositories

import (
	"context"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
)

// InventoryWithPharmacy es la vista cross-table que combina un item de
// inventario con datos de la farmacia que lo ofrece. La usa el endpoint
// de comparación de precios (GET /pharmacies/inventory/product/{id}).
//
// Nota sobre nomenclatura geográfica (Perú):
//   - District: distrito (entidad administrativa fina, ej. "MIRAFLORES").
//     Mapea físicamente al campo `city` de la tabla `pharmacies` porque la
//     data DIGEMID registra establecimientos por distrito y el scraper la
//     persiste en esa columna. Hasta que el negocio pida separar
//     departamento/provincia/distrito en columnas independientes, este es
//     el contrato de lectura.
//   - Address: dirección de calle (campo `street`).
type InventoryWithPharmacy struct {
	ID           string
	PharmacyID   string
	PharmacySlug string
	PharmacyName string
	District     string
	Address      string
	ProductID    string
	Stock        int
	Price        float64
	IsAvailable  bool
	// DistanceKm se llena solo cuando la query incluye lat/lng (HU-014).
	// Nil en respuestas sin geolocalización.
	DistanceKm *float64
	// DistrictAvgPrice (HU-016) — precio promedio del producto en el mismo
	// distrito que esta farmacia. Nil si la farmacia no tiene `District`
	// asignado o si SQL no encontró otras farmacias del mismo distrito con
	// el producto. Es DATO BRUTO; la decisión de "caro" la toma el dominio
	// con `services.IsOverpriced(...)`, no el SQL.
	DistrictAvgPrice *float64
	// DistrictPharmacyCount — cuántas farmacias del mismo distrito ofrecen
	// este producto (incluyendo la actual). Se usa para validar que el
	// promedio sea estadísticamente representativo antes de marcar como caro.
	DistrictPharmacyCount int
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// GeoFilter agrupa los parámetros opcionales de geolocalización para
// queries de inventario. Si Lat == 0 && Lng == 0 se considera "sin geo"
// y la query se ejecuta como antes (orden por precio, sin distancia).
// RadiusKm == 0 con lat/lng presentes significa "sin filtro de radio,
// pero calcular y ordenar por distancia".
type GeoFilter struct {
	Lat      float64
	Lng      float64
	RadiusKm float64
}

func (g GeoFilter) IsActive() bool { return g.Lat != 0 || g.Lng != 0 }

// InventoryUpsertParams — payload para UpsertByPharmacyAndProduct. Lo
// invoca el SQS consumer de INVENTORY_DISCOVERED tras resolver los UUIDs
// de farmacia y producto via lookups locales/HTTP.
type InventoryUpsertParams struct {
	PharmacyID  string
	ProductID   string
	Stock       int
	Price       float64
	IsAvailable bool
}

type InventoryRepository interface {
	Create(ctx context.Context, item *entities.PharmacyInventory) error
	FindByPharmacyID(ctx context.Context, pharmacyID string) ([]entities.PharmacyInventory, error)
	FindByPharmacyAndProduct(ctx context.Context, pharmacyID, productID string) (*entities.PharmacyInventory, error)
	FindByProductID(ctx context.Context, productID string, geo GeoFilter) ([]InventoryWithPharmacy, error)
	Update(ctx context.Context, item *entities.PharmacyInventory) error
	// UpsertByPharmacyAndProduct hace INSERT ... ON CONFLICT (pharmacy_id, product_id)
	// DO UPDATE. Idempotente. Usado por el SQS consumer de INVENTORY_DISCOVERED.
	UpsertByPharmacyAndProduct(ctx context.Context, params InventoryUpsertParams) error
	Delete(ctx context.Context, pharmacyID, productID string) error
}
