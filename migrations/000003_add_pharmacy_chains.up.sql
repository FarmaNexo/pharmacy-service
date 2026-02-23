-- 000003_add_pharmacy_chains.up.sql
-- Agrega soporte para cadenas de farmacias

ALTER TABLE pharmacy.pharmacies
    ADD COLUMN IF NOT EXISTS chain_id VARCHAR(100),
    ADD COLUMN IF NOT EXISTS chain_name VARCHAR(255);

-- Índice para búsquedas por cadena
CREATE INDEX IF NOT EXISTS idx_pharmacies_chain_id
    ON pharmacy.pharmacies (chain_id)
    WHERE chain_id IS NOT NULL;
