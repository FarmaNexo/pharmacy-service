-- 000001_init_schema.up.sql
-- Pharmacy Service - Schema inicial con PostGIS

CREATE SCHEMA IF NOT EXISTS pharmacy;

-- Habilitar PostGIS
CREATE EXTENSION IF NOT EXISTS postgis;

-- Farmacias
CREATE TABLE pharmacy.pharmacies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    phone VARCHAR(20),
    email VARCHAR(255),
    website VARCHAR(500),
    logo_url VARCHAR(500),

    -- Dirección
    street VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) NOT NULL DEFAULT 'Perú',

    -- Geolocalización (PostGIS)
    location GEOGRAPHY(POINT, 4326),

    -- Estado
    is_verified BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_24h BOOLEAN NOT NULL DEFAULT false,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_pharmacies_location ON pharmacy.pharmacies USING GIST(location);
CREATE INDEX idx_pharmacies_owner_user_id ON pharmacy.pharmacies(owner_user_id);
CREATE INDEX idx_pharmacies_slug ON pharmacy.pharmacies(slug);
CREATE INDEX idx_pharmacies_is_active ON pharmacy.pharmacies(is_active);
CREATE INDEX idx_pharmacies_is_verified ON pharmacy.pharmacies(is_verified);

-- Horarios de atención
CREATE TABLE pharmacy.pharmacy_hours (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pharmacy_id UUID NOT NULL REFERENCES pharmacy.pharmacies(id) ON DELETE CASCADE,
    day_of_week INTEGER NOT NULL,
    open_time TIME NOT NULL,
    close_time TIME NOT NULL,
    is_closed BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(pharmacy_id, day_of_week)
);

CREATE INDEX idx_pharmacy_hours_pharmacy_id ON pharmacy.pharmacy_hours(pharmacy_id);

-- Inventario
CREATE TABLE pharmacy.pharmacy_inventory (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pharmacy_id UUID NOT NULL REFERENCES pharmacy.pharmacies(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,

    stock INTEGER NOT NULL DEFAULT 0,
    price DECIMAL(10, 2) NOT NULL,

    is_available BOOLEAN NOT NULL DEFAULT true,

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(pharmacy_id, product_id)
);

CREATE INDEX idx_pharmacy_inventory_pharmacy_id ON pharmacy.pharmacy_inventory(pharmacy_id);
CREATE INDEX idx_pharmacy_inventory_product_id ON pharmacy.pharmacy_inventory(product_id);
CREATE INDEX idx_pharmacy_inventory_is_available ON pharmacy.pharmacy_inventory(is_available);
