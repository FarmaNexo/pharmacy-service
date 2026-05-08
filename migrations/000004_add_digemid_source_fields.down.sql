-- Revierte 000004.
--
-- OJO: el DROP NOT NULL no se revierte literalmente porque puede haber ya
-- filas con NULL. El down solo restaura columnas/indices agregados.

DROP INDEX IF EXISTS pharmacy.idx_pharmacies_unclaimed;
DROP INDEX IF EXISTS pharmacy.idx_pharmacies_ruc;
DROP INDEX IF EXISTS pharmacy.uq_pharmacies_source_digemid;

ALTER TABLE pharmacy.pharmacies
    DROP COLUMN IF EXISTS hours_raw,
    DROP COLUMN IF EXISTS technical_director,
    DROP COLUMN IF EXISTS ruc,
    DROP COLUMN IF EXISTS source_pharmacy_code;
