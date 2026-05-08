// internal/infrastructure/persistence/postgres/inventory_repository_impl.go
package postgres

import (
	"context"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type InventoryRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewInventoryRepository(db *gorm.DB, logger *zap.Logger) repositories.InventoryRepository {
	return &InventoryRepositoryImpl{db: db, logger: logger}
}

func (r *InventoryRepositoryImpl) Create(ctx context.Context, item *entities.PharmacyInventory) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *InventoryRepositoryImpl) FindByPharmacyID(ctx context.Context, pharmacyID string) ([]entities.PharmacyInventory, error) {
	var items []entities.PharmacyInventory
	result := r.db.WithContext(ctx).
		Where("pharmacy_id = ?", pharmacyID).
		Order("created_at DESC").
		Find(&items)
	if result.Error != nil {
		return nil, result.Error
	}
	return items, nil
}

func (r *InventoryRepositoryImpl) FindByPharmacyAndProduct(ctx context.Context, pharmacyID, productID string) (*entities.PharmacyInventory, error) {
	var item entities.PharmacyInventory
	result := r.db.WithContext(ctx).
		Where("pharmacy_id = ? AND product_id = ?", pharmacyID, productID).
		First(&item)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &item, nil
}

func (r *InventoryRepositoryImpl) Update(ctx context.Context, item *entities.PharmacyInventory) error {
	item.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *InventoryRepositoryImpl) Delete(ctx context.Context, pharmacyID, productID string) error {
	result := r.db.WithContext(ctx).
		Where("pharmacy_id = ? AND product_id = ?", pharmacyID, productID).
		Delete(&entities.PharmacyInventory{})
	return result.Error
}

// FindByProductID retorna todo el inventario (de farmacias activas) que
// contiene un producto dado. Sin geo: ordenado por precio ascendente.
// Con geo (lat/lng != 0): incluye distance_km vía PostGIS, opcionalmente
// filtra por radio (ST_DWithin) y ordena por distancia ascendente.
func (r *InventoryRepositoryImpl) FindByProductID(ctx context.Context, productID string, geo repositories.GeoFilter) ([]repositories.InventoryWithPharmacy, error) {
	if geo.IsActive() {
		return r.findByProductIDWithGeo(ctx, productID, geo)
	}
	return r.findByProductIDPriceOrdered(ctx, productID)
}

func (r *InventoryRepositoryImpl) findByProductIDPriceOrdered(ctx context.Context, productID string) ([]repositories.InventoryWithPharmacy, error) {
	type row struct {
		ID                    string
		PharmacyID            string
		PharmacySlug          string
		PharmacyName          string
		District              string
		Address               string
		ProductID             string
		Stock                 int
		Price                 float64
		IsAvailable           bool
		DistrictAvgPrice      *float64
		DistrictPharmacyCount int
		CreatedAt             time.Time
		UpdatedAt             time.Time
	}

	// p.city actúa como "distrito" (data DIGEMID); p.street es la dirección.
	// COALESCE para evitar NULL en farmacias scrapeadas con campos relajados.
	// HU-016: la CTE `district_stats` calcula promedio + count por distrito
	// para este producto en una sola pasada. LEFT JOIN para que farmacias sin
	// distrito asignado ('') igual aparezcan (con avg nil).
	query := `
		WITH district_stats AS (
			SELECT
				p2.city                AS district,
				AVG(i2.price)::float8  AS avg_price,
				COUNT(*)               AS pharmacy_count
			FROM pharmacy.pharmacy_inventory i2
			INNER JOIN pharmacy.pharmacies p2 ON p2.id = i2.pharmacy_id
			WHERE i2.product_id = $1
			  AND p2.is_active = true
			  AND p2.deleted_at IS NULL
			  AND p2.city IS NOT NULL
			  AND p2.city <> ''
			GROUP BY p2.city
		)
		SELECT
			i.id                            AS id,
			i.pharmacy_id                   AS pharmacy_id,
			p.slug                          AS pharmacy_slug,
			p.name                          AS pharmacy_name,
			COALESCE(p.city, '')            AS district,
			COALESCE(p.street, '')          AS address,
			i.product_id                    AS product_id,
			i.stock                         AS stock,
			i.price                         AS price,
			i.is_available                  AS is_available,
			ds.avg_price                    AS district_avg_price,
			COALESCE(ds.pharmacy_count, 0)  AS district_pharmacy_count,
			i.created_at                    AS created_at,
			i.updated_at                    AS updated_at
		FROM pharmacy.pharmacy_inventory i
		INNER JOIN pharmacy.pharmacies p ON p.id = i.pharmacy_id
		LEFT JOIN district_stats ds ON ds.district = p.city
		WHERE i.product_id = $1
		  AND p.is_active = true
		  AND p.deleted_at IS NULL
		ORDER BY i.price ASC
	`

	var rows []row
	result := r.db.WithContext(ctx).Raw(query, productID).Scan(&rows)
	if result.Error != nil {
		r.logger.Error("Error buscando inventario por producto",
			zap.String("product_id", productID),
			zap.Error(result.Error),
		)
		return nil, result.Error
	}

	items := make([]repositories.InventoryWithPharmacy, len(rows))
	for i, rr := range rows {
		items[i] = repositories.InventoryWithPharmacy{
			ID:                    rr.ID,
			PharmacyID:            rr.PharmacyID,
			PharmacySlug:          rr.PharmacySlug,
			PharmacyName:          rr.PharmacyName,
			District:              rr.District,
			Address:               rr.Address,
			ProductID:             rr.ProductID,
			Stock:                 rr.Stock,
			Price:                 rr.Price,
			IsAvailable:           rr.IsAvailable,
			DistrictAvgPrice:      rr.DistrictAvgPrice,
			DistrictPharmacyCount: rr.DistrictPharmacyCount,
			CreatedAt:             rr.CreatedAt,
			UpdatedAt:             rr.UpdatedAt,
		}
	}
	return items, nil
}

// findByProductIDWithGeo aplica el patrón PostGIS de pharmacy_repository.FindNearby:
// ST_Distance(geography, point) / 1000 para km, ST_DWithin para filtro
// por radio. ST_MakePoint usa orden (lng, lat). Si radius_km <= 0, no se
// aplica filtro de radio — solo se calcula y ordena por distancia.
//
// HU-016: incluye CTE `district_stats` (idéntica al branch sin geo) que
// calcula promedio + count de farmacias por distrito que ofrecen el
// producto. Importante: el promedio se calcula SOBRE TODO EL UNIVERSO
// (sin filtrar por radio) — el "promedio del distrito" es una propiedad
// del distrito, no de la búsqueda del usuario.
func (r *InventoryRepositoryImpl) findByProductIDWithGeo(ctx context.Context, productID string, geo repositories.GeoFilter) ([]repositories.InventoryWithPharmacy, error) {
	type row struct {
		ID                    string
		PharmacyID            string
		PharmacySlug          string
		PharmacyName          string
		District              string
		Address               string
		ProductID             string
		Stock                 int
		Price                 float64
		IsAvailable           bool
		DistanceKm            float64
		DistrictAvgPrice      *float64
		DistrictPharmacyCount int
		CreatedAt             time.Time
		UpdatedAt             time.Time
	}

	// args: $1=lng, $2=lat, $3=productID, [$4=radiusMeters si aplica]
	radiusFilter := ""
	args := []any{geo.Lng, geo.Lat, productID}
	if geo.RadiusKm > 0 {
		radiusFilter = "AND ST_DWithin(p.location, ST_MakePoint($1, $2)::geography, $4)"
		args = append(args, geo.RadiusKm*1000.0)
	}

	query := `
		WITH district_stats AS (
			SELECT
				p2.city                AS district,
				AVG(i2.price)::float8  AS avg_price,
				COUNT(*)               AS pharmacy_count
			FROM pharmacy.pharmacy_inventory i2
			INNER JOIN pharmacy.pharmacies p2 ON p2.id = i2.pharmacy_id
			WHERE i2.product_id = $3
			  AND p2.is_active = true
			  AND p2.deleted_at IS NULL
			  AND p2.city IS NOT NULL
			  AND p2.city <> ''
			GROUP BY p2.city
		)
		SELECT
			i.id                                                                 AS id,
			i.pharmacy_id                                                        AS pharmacy_id,
			p.slug                                                               AS pharmacy_slug,
			p.name                                                               AS pharmacy_name,
			COALESCE(p.city, '')                                                 AS district,
			COALESCE(p.street, '')                                               AS address,
			i.product_id                                                         AS product_id,
			i.stock                                                              AS stock,
			i.price                                                              AS price,
			i.is_available                                                       AS is_available,
			ST_Distance(p.location, ST_MakePoint($1, $2)::geography) / 1000.0    AS distance_km,
			ds.avg_price                                                         AS district_avg_price,
			COALESCE(ds.pharmacy_count, 0)                                       AS district_pharmacy_count,
			i.created_at                                                         AS created_at,
			i.updated_at                                                         AS updated_at
		FROM pharmacy.pharmacy_inventory i
		INNER JOIN pharmacy.pharmacies p ON p.id = i.pharmacy_id
		LEFT JOIN district_stats ds ON ds.district = p.city
		WHERE i.product_id = $3
		  AND p.is_active = true
		  AND p.deleted_at IS NULL
		  AND p.location IS NOT NULL
		  ` + radiusFilter + `
		ORDER BY distance_km ASC
	`

	var rows []row
	result := r.db.WithContext(ctx).Raw(query, args...).Scan(&rows)
	if result.Error != nil {
		r.logger.Error("Error buscando inventario por producto con geo",
			zap.String("product_id", productID),
			zap.Float64("lat", geo.Lat),
			zap.Float64("lng", geo.Lng),
			zap.Float64("radius_km", geo.RadiusKm),
			zap.Error(result.Error),
		)
		return nil, result.Error
	}

	items := make([]repositories.InventoryWithPharmacy, len(rows))
	for i, rr := range rows {
		d := rr.DistanceKm
		items[i] = repositories.InventoryWithPharmacy{
			ID:                    rr.ID,
			PharmacyID:            rr.PharmacyID,
			PharmacySlug:          rr.PharmacySlug,
			PharmacyName:          rr.PharmacyName,
			District:              rr.District,
			Address:               rr.Address,
			ProductID:             rr.ProductID,
			Stock:                 rr.Stock,
			Price:                 rr.Price,
			IsAvailable:           rr.IsAvailable,
			DistanceKm:            &d,
			DistrictAvgPrice:      rr.DistrictAvgPrice,
			DistrictPharmacyCount: rr.DistrictPharmacyCount,
			CreatedAt:             rr.CreatedAt,
			UpdatedAt:             rr.UpdatedAt,
		}
	}
	return items, nil
}

var _ repositories.InventoryRepository = (*InventoryRepositoryImpl)(nil)
