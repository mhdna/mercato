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
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/mhdna/kashi/db/sqlc Store

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
    --go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative\
    proto/*.proto

evans:
	evans --host localhost --port 8088 --package pb -r repl

.PHONY: postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 sqlc test server mock proto
