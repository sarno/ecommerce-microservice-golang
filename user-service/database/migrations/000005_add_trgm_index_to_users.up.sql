CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS users_name_phone_email_trgm_idx ON users USING gin ((name || ' ' || phone || ' ' || email) gin_trgm_ops);
