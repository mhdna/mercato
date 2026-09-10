postgres:
	docker run --name postgres -p 5433:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root kashi

dropdb:
	docker exec -it postgres dropdb kashi

migrateup:
	migrate -path ./db/migrations -database "postgresql://root:secret@localhost:5433/kashi?sslmode=disable" -verbose up

# migrate up the last migration
migrateup1:
	migrate -path ./db/migrations -database "postgresql://root:secret@localhost:5433/kashi?sslmode=disable" -verbose up 1

migratedown:
	migrate -path ./db/migrations -database "postgresql://root:secret@localhost:5433/kashi?sslmode=disable" -verbose down

# migrate down the last migration
migratedown1:
	migrate -path ./db/migrations -database "postgresql://root:secret@localhost:5433/kashi?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	# go test -v -cover ./...
	richgo test -v -cover ./...

server:
	go run .

# Live-reloading server for local dev. Needs `air` on PATH:
#   go install github.com/air-verse/air@latest
dev:
	air

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/mhdna/kashi/db/sqlc Store

proto:
	rm -f pb/*.go
	rm -rf doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=kashi \
    proto/*.proto
	statik -src=./doc/swagger/ -dest=./doc

evans:
	evans --host localhost --port 8088 --package pb -r repl

seed:
	go run ./cmd/seed/main.go

# --- cloud deploy (grandbz.com droplet) -----------------------------------
# One-time server setup lives in ui/deploy/SETUP.md. Needs an ssh alias
# `kashi` -> root@178.62.8.143 (in ~/.ssh/config).
SSH      ?= kashi
UI_ROOT  ?= /var/www/kashi-ui
APP_DIR  ?= /home/kashi/app
PROD_DB  ?= postgresql://kashi:kashi@localhost:5432/kashi?sslmode=disable

# Full release: API binary + migrations, then the UI.
deploy: deploy-api deploy-ui
	@echo "Deployed -> https://grandbz.com/"

# Cross-compile the API, ship it with the current migrations, run any
# pending ones on the prod DB (idempotent), restart the service.
deploy-api:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o /tmp/kashi-api-linux .
	rsync -avz /tmp/kashi-api-linux $(SSH):$(APP_DIR)/api-linux
	rsync -avz --delete ./db/migrations/ $(SSH):$(APP_DIR)/migrations/
	ssh $(SSH) 'chown -R kashi:kashi $(APP_DIR)/api-linux $(APP_DIR)/migrations && \
		chmod +x $(APP_DIR)/api-linux && \
		migrate -path $(APP_DIR)/migrations -database "$(PROD_DB)" up && \
		systemctl restart kashi-api && sleep 1 && systemctl is-active kashi-api'

# Build the SPA (uses ui/.env.production -> VITE_API_URL=/api) and mirror it
# into the Caddy web root. Fast path for front-end-only changes.
deploy-ui:
	$(MAKE) -C ui build
	rsync -avz --delete --chmod=D755,F644 ui/dist/ $(SSH):$(UI_ROOT)/

# Run pending migrations on prod without shipping anything.
migrate-prod:
	rsync -avz --delete ./db/migrations/ $(SSH):$(APP_DIR)/migrations/
	ssh $(SSH) 'migrate -path $(APP_DIR)/migrations -database "$(PROD_DB)" up'

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server dev mock proto evans seed deploy deploy-api deploy-ui migrate-prod
