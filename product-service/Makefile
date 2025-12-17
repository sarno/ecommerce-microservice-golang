# =========================
# Configuration
# =========================
ENV_FILE=.env
MIGRATIONS_DIR=database/migrations

# Load env file
ifneq (,$(wildcard $(ENV_FILE)))
	include $(ENV_FILE)
	export
endif

# =========================
# Migrate commands
# =========================

migrate-version:
	migrate -database "$(POSTGRESQL_URL)" -path $(MIGRATIONS_DIR) version

migrate-up:
	migrate -database "$(POSTGRESQL_URL)" -path $(MIGRATIONS_DIR) up

migrate-up-one:
	migrate -database "$(POSTGRESQL_URL)" -path $(MIGRATIONS_DIR) up 1

migrate-down:
	migrate -database "$(POSTGRESQL_URL)" -path $(MIGRATIONS_DIR) down

migrate-down-one:
	migrate -database "$(POSTGRESQL_URL)" -path $(MIGRATIONS_DIR) down 1

migrate-force:
	migrate -database "$(POSTGRESQL_URL)" -path $(MIGRATIONS_DIR) force $(version)

migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

# =========================
# Helpers
# =========================

.PHONY: migrate-up migrate-down migrate-version migrate-force migrate-create
