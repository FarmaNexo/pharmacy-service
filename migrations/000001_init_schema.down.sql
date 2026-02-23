-- 000001_init_schema.down.sql
DROP TABLE IF EXISTS pharmacy.pharmacy_inventory;
DROP TABLE IF EXISTS pharmacy.pharmacy_hours;
DROP TABLE IF EXISTS pharmacy.pharmacies;
DROP SCHEMA IF EXISTS pharmacy;
DROP EXTENSION IF EXISTS postgis;
