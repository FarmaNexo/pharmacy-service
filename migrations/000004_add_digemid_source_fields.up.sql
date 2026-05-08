-- Migration 000004 — soporte para farmacias scrapeadas (fuente DIGEMID)
--
-- Contexto: el scraper-service ingesta establecimientos del Observatorio
-- DIGEMID. Estas farmacias existen en la realidad pero aún no tienen dueño
-- registrado en FarmaNexo. Esta migración:
--   1. Permite farmacias sin owner_user_id (se "reclaman" cuando alguien se registra)
--   2. Añade source_pharmacy_code para UPSERT idempotente por fuente externa
--   3. Añade campos regulatorios peruanos (RUC, director técnico) que hoy no existen
--   4. Añade hours_raw para guardar el string de horarios tal como viene del DIGEMID

-- 1. owner_user_id NULLABLE: las farmacias scrapeadas no tienen dueño.
-- Cuando alguien se registre, el admin asigna el owner_user_id y
-- opcionalmente flag is_verified=true.
ALTER TABLE pharmacy.pharmacies
    ALTER COLUMN owner_user_id DROP NOT NULL;

-- 2. source_pharmacy_code: código del establecimiento en DIGEMID (ej. "0022705").
-- Clave natural para evitar duplicados cuando el scraper re-corre.
ALTER TABLE pharmacy.pharmacies
    ADD COLUMN IF NOT EXISTS source_pharmacy_code VARCHAR(20);

-- 3. ruc: Registro Único de Contribuyentes (11 dígitos en Perú).
-- Obligatorio legalmente; hoy no existía.
ALTER TABLE pharmacy.pharmacies
    ADD COLUMN IF NOT EXISTS ruc VARCHAR(20);

-- 4. technical_director: nombre del Director Técnico (químico farmacéutico).
-- Obligatorio por ley peruana para cada farmacia.
ALTER TABLE pharmacy.pharmacies
    ADD COLUMN IF NOT EXISTS technical_director VARCHAR(255);

-- 5. hours_raw: string de horarios tal como viene del portal DIGEMID
-- (ej. "LUN A VIE: 08:00 A 23:00; SAB: 08:00 A 23:00; DOM: 08:00 A 23:00").
-- El parser a pharmacy_hours (tabla normalizada) se agregará en otra iteración.
ALTER TABLE pharmacy.pharmacies
    ADD COLUMN IF NOT EXISTS hours_raw TEXT;

-- Street/city se relajan a NULLABLE porque algunas farmacias scrapeadas
-- vienen con dirección vacía. Los datos reales las traen, pero no queremos
-- fallar el INSERT por direcciones ausentes.
ALTER TABLE pharmacy.pharmacies
    ALTER COLUMN street DROP NOT NULL;

ALTER TABLE pharmacy.pharmacies
    ALTER COLUMN city DROP NOT NULL;

-- Índice único parcial en source_pharmacy_code — permite UPSERT por fuente DIGEMID
-- y deja convivir farmacias registradas manualmente (source_pharmacy_code NULL).
CREATE UNIQUE INDEX IF NOT EXISTS uq_pharmacies_source_digemid
    ON pharmacy.pharmacies (source_pharmacy_code)
    WHERE source_pharmacy_code IS NOT NULL;

-- Índice en ruc para lookups (búsquedas por RUC son frecuentes).
CREATE INDEX IF NOT EXISTS idx_pharmacies_ruc
    ON pharmacy.pharmacies (ruc)
    WHERE ruc IS NOT NULL;

-- Partial index útil para el dashboard admin: "pharmacies sin dueño".
CREATE INDEX IF NOT EXISTS idx_pharmacies_unclaimed
    ON pharmacy.pharmacies (source_pharmacy_code)
    WHERE owner_user_id IS NULL AND source_pharmacy_code IS NOT NULL;
