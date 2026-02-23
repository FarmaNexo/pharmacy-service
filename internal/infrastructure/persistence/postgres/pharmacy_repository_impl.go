// internal/infrastructure/persistence/postgres/pharmacy_repository_impl.go
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/farmanexo/pharmacy-service/internal/domain/entities"
	"github.com/farmanexo/pharmacy-service/internal/domain/repositories"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PharmacyRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewPharmacyRepository(db *gorm.DB, logger *zap.Logger) repositories.PharmacyRepository {
	return &PharmacyRepositoryImpl{db: db, logger: logger}
}

// pharmacySelectColumns is the standard SELECT clause for pharmacy queries
const pharmacySelectColumns = `
	id, owner_user_id, name, slug, description, phone, email, website, logo_url,
	authorization_document_url, chain_id, chain_name,
	street, city, state, postal_code, country,
	ST_Y(location::geometry) AS latitude,
	ST_X(location::geometry) AS longitude,
	is_verified, is_active, is_24h, created_at, updated_at
`

func (r *PharmacyRepositoryImpl) Create(ctx context.Context, pharmacy *entities.Pharmacy) error {
	query := `
		INSERT INTO pharmacy.pharmacies
			(id, owner_user_id, name, slug, description, phone, email, website, logo_url,
			 authorization_document_url, chain_id, chain_name,
			 street, city, state, postal_code, country, location,
			 is_verified, is_active, is_24h, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9,
			 $10, $11, $12,
			 $13, $14, $15, $16, $17, ST_MakePoint($18, $19)::geography,
			 $20, $21, $22, $23, $24)
	`
	now := time.Now()
	result := r.db.WithContext(ctx).Exec(query,
		pharmacy.ID, pharmacy.OwnerUserID, pharmacy.Name, pharmacy.Slug,
		pharmacy.Description, pharmacy.Phone, pharmacy.Email, pharmacy.Website, pharmacy.LogoURL,
		pharmacy.AuthorizationDocumentURL, pharmacy.ChainID, pharmacy.ChainName,
		pharmacy.Street, pharmacy.City, pharmacy.State, pharmacy.PostalCode, pharmacy.Country,
		pharmacy.Longitude, pharmacy.Latitude,
		pharmacy.IsVerified, pharmacy.IsActive, pharmacy.Is24h, now, now,
	)
	if result.Error != nil {
		r.logger.Error("Error creando farmacia", zap.Error(result.Error))
		return result.Error
	}
	pharmacy.CreatedAt = now
	pharmacy.UpdatedAt = now
	return nil
}

func (r *PharmacyRepositoryImpl) FindByID(ctx context.Context, id string) (*entities.Pharmacy, error) {
	var pharmacy entities.Pharmacy
	query := `SELECT ` + pharmacySelectColumns + `, deleted_at
		FROM pharmacy.pharmacies
		WHERE id = $1 AND deleted_at IS NULL
	`
	result := r.db.WithContext(ctx).Raw(query, id).Scan(&pharmacy)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pharmacy, nil
}

func (r *PharmacyRepositoryImpl) FindByIDWithDeleted(ctx context.Context, id string) (*entities.Pharmacy, error) {
	var pharmacy entities.Pharmacy
	query := `SELECT ` + pharmacySelectColumns + `, deleted_at
		FROM pharmacy.pharmacies
		WHERE id = $1
	`
	result := r.db.WithContext(ctx).Raw(query, id).Scan(&pharmacy)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pharmacy, nil
}

func (r *PharmacyRepositoryImpl) FindBySlug(ctx context.Context, slug string) (*entities.Pharmacy, error) {
	var pharmacy entities.Pharmacy
	query := `SELECT ` + pharmacySelectColumns + `
		FROM pharmacy.pharmacies
		WHERE slug = $1 AND deleted_at IS NULL
	`
	result := r.db.WithContext(ctx).Raw(query, slug).Scan(&pharmacy)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pharmacy, nil
}

func (r *PharmacyRepositoryImpl) FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*entities.Pharmacy, error) {
	var pharmacies []*entities.Pharmacy
	query := `SELECT ` + pharmacySelectColumns + `
		FROM pharmacy.pharmacies
		WHERE owner_user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`
	result := r.db.WithContext(ctx).Raw(query, ownerUserID).Scan(&pharmacies)
	if result.Error != nil {
		return nil, result.Error
	}
	return pharmacies, nil
}

func (r *PharmacyRepositoryImpl) FindAll(ctx context.Context, page, limit int) (*repositories.PaginatedResult, error) {
	var total int64
	r.db.WithContext(ctx).Raw("SELECT COUNT(*) FROM pharmacy.pharmacies WHERE deleted_at IS NULL AND is_active = true").Scan(&total)

	offset := (page - 1) * limit
	var pharmacies []*entities.Pharmacy
	query := `SELECT ` + pharmacySelectColumns + `
		FROM pharmacy.pharmacies
		WHERE deleted_at IS NULL AND is_active = true
		ORDER BY name ASC
		LIMIT $1 OFFSET $2
	`
	result := r.db.WithContext(ctx).Raw(query, limit, offset).Scan(&pharmacies)
	if result.Error != nil {
		return nil, result.Error
	}

	return &repositories.PaginatedResult{Total: total, Items: pharmacies}, nil
}

func (r *PharmacyRepositoryImpl) FindNearby(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]*entities.Pharmacy, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	radiusMeters := radiusKm * 1000

	var pharmacies []*entities.Pharmacy
	query := `SELECT ` + pharmacySelectColumns + `,
			   ST_Distance(location, ST_MakePoint($1, $2)::geography) / 1000.0 AS distance_km
		FROM pharmacy.pharmacies
		WHERE
			ST_DWithin(location, ST_MakePoint($1, $2)::geography, $3)
			AND is_active = true
			AND is_verified = true
			AND deleted_at IS NULL
		ORDER BY distance_km ASC
		LIMIT $4
	`
	result := r.db.WithContext(ctx).Raw(query, lng, lat, radiusMeters, limit).Scan(&pharmacies)
	if result.Error != nil {
		r.logger.Error("Error buscando farmacias cercanas", zap.Error(result.Error))
		return nil, result.Error
	}
	return pharmacies, nil
}

func (r *PharmacyRepositoryImpl) FindByChainID(ctx context.Context, chainID string, page, limit int) (*repositories.PaginatedResult, error) {
	var total int64
	r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM pharmacy.pharmacies WHERE chain_id = $1 AND deleted_at IS NULL AND is_active = true",
		chainID,
	).Scan(&total)

	offset := (page - 1) * limit
	var pharmacies []*entities.Pharmacy
	query := `SELECT ` + pharmacySelectColumns + `
		FROM pharmacy.pharmacies
		WHERE chain_id = $1 AND deleted_at IS NULL AND is_active = true
		ORDER BY name ASC
		LIMIT $2 OFFSET $3
	`
	result := r.db.WithContext(ctx).Raw(query, chainID, limit, offset).Scan(&pharmacies)
	if result.Error != nil {
		r.logger.Error("Error buscando farmacias por cadena", zap.Error(result.Error), zap.String("chain_id", chainID))
		return nil, result.Error
	}

	return &repositories.PaginatedResult{Total: total, Items: pharmacies}, nil
}

func (r *PharmacyRepositoryImpl) Update(ctx context.Context, pharmacy *entities.Pharmacy) error {
	pharmacy.UpdatedAt = time.Now()
	query := `
		UPDATE pharmacy.pharmacies SET
			name = $2, slug = $3, description = $4, phone = $5, email = $6,
			website = $7, logo_url = $8, authorization_document_url = $9,
			chain_id = $10, chain_name = $11,
			street = $12, city = $13, state = $14,
			postal_code = $15, country = $16, location = ST_MakePoint($17, $18)::geography,
			is_verified = $19, is_active = $20, is_24h = $21, updated_at = $22
		WHERE id = $1
	`
	result := r.db.WithContext(ctx).Exec(query,
		pharmacy.ID, pharmacy.Name, pharmacy.Slug, pharmacy.Description,
		pharmacy.Phone, pharmacy.Email, pharmacy.Website, pharmacy.LogoURL,
		pharmacy.AuthorizationDocumentURL,
		pharmacy.ChainID, pharmacy.ChainName,
		pharmacy.Street, pharmacy.City, pharmacy.State, pharmacy.PostalCode, pharmacy.Country,
		pharmacy.Longitude, pharmacy.Latitude,
		pharmacy.IsVerified, pharmacy.IsActive, pharmacy.Is24h, pharmacy.UpdatedAt,
	)
	if result.Error != nil {
		r.logger.Error("Error actualizando farmacia", zap.Error(result.Error))
		return result.Error
	}
	return nil
}

func (r *PharmacyRepositoryImpl) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Exec(
		"UPDATE pharmacy.pharmacies SET deleted_at = $1, is_active = false, updated_at = $1 WHERE id = $2",
		now, id,
	)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("pharmacy not found")
	}
	return nil
}

var _ repositories.PharmacyRepository = (*PharmacyRepositoryImpl)(nil)
