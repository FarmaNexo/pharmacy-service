-- 000003_add_pharmacy_chains.down.sql

DROP INDEX IF EXISTS pharmacy.idx_pharmacies_chain_id;

ALTER TABLE pharmacy.pharmacies
    DROP COLUMN IF EXISTS chain_id,
    DROP COLUMN IF EXISTS chain_name;
