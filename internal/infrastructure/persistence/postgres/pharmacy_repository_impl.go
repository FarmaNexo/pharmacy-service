// internal/infrastructure/persistence/postgres/pharmacy_repository_impl.go
package postgres

import (
	"context"
	"fmt"
	"regexp"
	"strings"
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
	source_pharmacy_code, ruc, technical_director, hours_raw,
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
			 source_pharmacy_code, ruc, technical_director, hours_raw,
			 street, city, state, postal_code, country, location,
			 is_verified, is_active, is_24h, created_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9,
			 $10, $11, $12,
			 $13, NULLIF($14,''), NULLIF($15,''), NULLIF($16,''),
			 $17, $18, $19, $20, $21, ST_MakePoint($22, $23)::geography,
			 $24, $25, $26, $27, $28)
	`
	now := time.Now()
	result := r.db.WithContext(ctx).Exec(query,
		pharmacy.ID, pharmacy.OwnerUserID, pharmacy.Name, pharmacy.Slug,
		pharmacy.Description, pharmacy.Phone, pharmacy.Email, pharmacy.Website, pharmacy.LogoURL,
		pharmacy.AuthorizationDocumentURL, pharmacy.ChainID, pharmacy.ChainName,
		pharmacy.SourcePharmacyCode, pharmacy.RUC, pharmacy.TechnicalDirector, pharmacy.HoursRaw,
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
			ruc = COALESCE(NULLIF($12,''), ruc),
			technical_director = COALESCE(NULLIF($13,''), technical_director),
			hours_raw = COALESCE(NULLIF($14,''), hours_raw),
			street = $15, city = $16, state = $17,
			postal_code = $18, country = $19, location = ST_MakePoint($20, $21)::geography,
			is_verified = $22, is_active = $23, is_24h = $24, updated_at = $25
		WHERE id = $1
	`
	result := r.db.WithContext(ctx).Exec(query,
		pharmacy.ID, pharmacy.Name, pharmacy.Slug, pharmacy.Description,
		pharmacy.Phone, pharmacy.Email, pharmacy.Website, pharmacy.LogoURL,
		pharmacy.AuthorizationDocumentURL,
		pharmacy.ChainID, pharmacy.ChainName,
		pharmacy.RUC, pharmacy.TechnicalDirector, pharmacy.HoursRaw,
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

// FindBySourceCode busca por la clave natural DIGEMID. Usado por el SQS
// consumer para resolver INVENTORY_DISCOVERED → pharmacy_id local.
func (r *PharmacyRepositoryImpl) FindBySourceCode(ctx context.Context, sourcePharmacyCode string) (*entities.Pharmacy, error) {
	var pharmacy entities.Pharmacy
	query := `SELECT ` + pharmacySelectColumns + `
		FROM pharmacy.pharmacies
		WHERE source_pharmacy_code = $1 AND deleted_at IS NULL
	`
	result := r.db.WithContext(ctx).Raw(query, sourcePharmacyCode).Scan(&pharmacy)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &pharmacy, nil
}

// UpsertBySource hace INSERT ... ON CONFLICT (source_pharmacy_code) DO UPDATE.
// Idempotente. Reglas de merge:
//   - name → siempre se sobrescribe con el valor del evento.
//   - phone/email/ruc/etc. → COALESCE: no borra valor previo si el evento trae vacío.
//   - location (PostGIS) → si el evento trae lat/lng, se actualiza; si no, se preserva.
//
// Slug determinístico: slugify(name + "-" + source_pharmacy_code) — único por fuente.
// owner_user_id queda NULL (la migración 000004 lo permite); se reclamará cuando
// el dueño se registre en FarmaNexo.
func (r *PharmacyRepositoryImpl) UpsertBySource(ctx context.Context, p repositories.PharmacyUpsertParams) (string, error) {
	slug := buildPharmacySlugFromSource(p.CanonicalName, p.SourcePharmacyCode)

	// Distrito → city (convención DIGEMID Perú); Departamento → state.
	city := firstNonEmptyStr(p.Distrito, p.Provincia)
	state := p.Departamento

	var resultID string
	var execErr error

	if p.Latitude != nil && p.Longitude != nil {
		execErr = r.db.WithContext(ctx).Raw(`
			INSERT INTO pharmacy.pharmacies (
				name, slug, phone, email, website, logo_url,
				source_pharmacy_code, ruc, technical_director, hours_raw,
				chain_id, chain_name,
				street, city, state, postal_code, country, location,
				is_verified, is_active, is_24h, created_at, updated_at
			) VALUES (
				?, ?, NULLIF(?, ''), NULLIF(?, ''), '', '',
				?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''),
				NULLIF(?, ''), NULLIF(?, ''),
				NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), '', 'Perú',
				ST_MakePoint(?, ?)::geography,
				false, true, false, NOW(), NOW()
			)
			ON CONFLICT (source_pharmacy_code)
			WHERE source_pharmacy_code IS NOT NULL
			DO UPDATE SET
				name               = EXCLUDED.name,
				phone              = COALESCE(EXCLUDED.phone, pharmacy.pharmacies.phone),
				email              = COALESCE(EXCLUDED.email, pharmacy.pharmacies.email),
				ruc                = COALESCE(EXCLUDED.ruc, pharmacy.pharmacies.ruc),
				technical_director = COALESCE(EXCLUDED.technical_director, pharmacy.pharmacies.technical_director),
				hours_raw          = COALESCE(EXCLUDED.hours_raw, pharmacy.pharmacies.hours_raw),
				chain_id           = COALESCE(EXCLUDED.chain_id, pharmacy.pharmacies.chain_id),
				chain_name         = COALESCE(EXCLUDED.chain_name, pharmacy.pharmacies.chain_name),
				street             = COALESCE(EXCLUDED.street, pharmacy.pharmacies.street),
				city               = COALESCE(EXCLUDED.city, pharmacy.pharmacies.city),
				state              = COALESCE(EXCLUDED.state, pharmacy.pharmacies.state),
				location           = COALESCE(pharmacy.pharmacies.location, EXCLUDED.location),
				updated_at         = NOW()
			RETURNING id::text
		`,
			p.CanonicalName, slug, p.Phone, p.Email,
			p.SourcePharmacyCode, p.RUC, p.TechnicalDirector, p.Hours,
			p.ChainID, p.ChainName,
			p.FullAddress, city, state,
			*p.Longitude, *p.Latitude,
		).Scan(&resultID).Error
	} else {
		execErr = r.db.WithContext(ctx).Raw(`
			INSERT INTO pharmacy.pharmacies (
				name, slug, phone, email, website, logo_url,
				source_pharmacy_code, ruc, technical_director, hours_raw,
				chain_id, chain_name,
				street, city, state, postal_code, country,
				is_verified, is_active, is_24h, created_at, updated_at
			) VALUES (
				?, ?, NULLIF(?, ''), NULLIF(?, ''), '', '',
				?, NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''),
				NULLIF(?, ''), NULLIF(?, ''),
				NULLIF(?, ''), NULLIF(?, ''), NULLIF(?, ''), '', 'Perú',
				false, true, false, NOW(), NOW()
			)
			ON CONFLICT (source_pharmacy_code)
			WHERE source_pharmacy_code IS NOT NULL
			DO UPDATE SET
				name               = EXCLUDED.name,
				phone              = COALESCE(EXCLUDED.phone, pharmacy.pharmacies.phone),
				email              = COALESCE(EXCLUDED.email, pharmacy.pharmacies.email),
				ruc                = COALESCE(EXCLUDED.ruc, pharmacy.pharmacies.ruc),
				technical_director = COALESCE(EXCLUDED.technical_director, pharmacy.pharmacies.technical_director),
				hours_raw          = COALESCE(EXCLUDED.hours_raw, pharmacy.pharmacies.hours_raw),
				chain_id           = COALESCE(EXCLUDED.chain_id, pharmacy.pharmacies.chain_id),
				chain_name         = COALESCE(EXCLUDED.chain_name, pharmacy.pharmacies.chain_name),
				street             = COALESCE(EXCLUDED.street, pharmacy.pharmacies.street),
				city               = COALESCE(EXCLUDED.city, pharmacy.pharmacies.city),
				state              = COALESCE(EXCLUDED.state, pharmacy.pharmacies.state),
				updated_at         = NOW()
			RETURNING id::text
		`,
			p.CanonicalName, slug, p.Phone, p.Email,
			p.SourcePharmacyCode, p.RUC, p.TechnicalDirector, p.Hours,
			p.ChainID, p.ChainName,
			p.FullAddress, city, state,
		).Scan(&resultID).Error
	}

	if execErr != nil {
		return "", execErr
	}
	if resultID == "" {
		return "", fmt.Errorf("UPSERT pharmacy no retornó id (source_pharmacy_code=%s)", p.SourcePharmacyCode)
	}
	return resultID, nil
}

// ----- Helpers internos para UpsertBySource -----

var slugPharmacyReplacer = regexp.MustCompile(`[^a-z0-9]+`)

func slugifyPharmacy(s string) string {
	s = strings.ToLower(s)
	s = slugPharmacyReplacer.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

func buildPharmacySlugFromSource(name, sourceCode string) string {
	return slugifyPharmacy(name + "-" + sourceCode)
}

func firstNonEmptyStr(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

var _ repositories.PharmacyRepository = (*PharmacyRepositoryImpl)(nil)
