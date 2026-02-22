-- Create separate databases for each microservice
-- This script runs on first PostgreSQL container start via /docker-entrypoint-initdb.d/

CREATE DATABASE auth_db;
CREATE DATABASE users_db;
CREATE DATABASE projects_db;
CREATE DATABASE applications_db;
CREATE DATABASE chat_db;
CREATE DATABASE reviews_db;
CREATE DATABASE notifications_db;
