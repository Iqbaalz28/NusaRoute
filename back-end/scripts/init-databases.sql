-- init-databases.sql
-- Creates separate databases for each microservice (Database-per-Service pattern)

CREATE DATABASE nusaroute_payment;
CREATE DATABASE nusaroute_pricing;
CREATE DATABASE nusaroute_order;
CREATE DATABASE nusaroute_courier;
CREATE DATABASE nusaroute_dispatch;
CREATE DATABASE nusaroute_hub;
CREATE DATABASE nusaroute_resolution;

-- Enable PostGIS extension on dispatch database for spatial queries
\c nusaroute_dispatch;
CREATE EXTENSION IF NOT EXISTS postgis;

-- Grant all privileges to the nusaroute user
GRANT ALL PRIVILEGES ON DATABASE nusaroute_payment TO nusaroute;
GRANT ALL PRIVILEGES ON DATABASE nusaroute_pricing TO nusaroute;
GRANT ALL PRIVILEGES ON DATABASE nusaroute_order TO nusaroute;
GRANT ALL PRIVILEGES ON DATABASE nusaroute_courier TO nusaroute;
GRANT ALL PRIVILEGES ON DATABASE nusaroute_dispatch TO nusaroute;
GRANT ALL PRIVILEGES ON DATABASE nusaroute_hub TO nusaroute;
GRANT ALL PRIVILEGES ON DATABASE nusaroute_resolution TO nusaroute;
