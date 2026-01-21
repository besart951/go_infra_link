-- Rollback user + auth schema migration

DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS business_details;
DROP TABLE IF EXISTS users;
