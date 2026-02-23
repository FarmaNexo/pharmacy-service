-- 000002_add_authorization_document.down.sql

DROP INDEX IF EXISTS pharmacy.idx_pharmacies_authorization_doc;

ALTER TABLE pharmacy.pharmacies
    DROP COLUMN IF EXISTS authorization_document_url;
