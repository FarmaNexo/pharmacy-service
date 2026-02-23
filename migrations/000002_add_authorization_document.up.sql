-- 000002_add_authorization_document.up.sql
-- Agrega columna para documento de autorización sanitaria

ALTER TABLE pharmacy.pharmacies
    ADD COLUMN IF NOT EXISTS authorization_document_url VARCHAR(500);

-- Índice para búsquedas de farmacias sin documento
CREATE INDEX IF NOT EXISTS idx_pharmacies_authorization_doc
    ON pharmacy.pharmacies (authorization_document_url)
    WHERE authorization_document_url IS NOT NULL;
