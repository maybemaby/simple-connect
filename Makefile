dev:
	air

test:
	go test -v ./...

lint:
	buf lint

generate:
	buf generate

migration-create:
	echo "Enter migration name: "; \
	read NAME; \
	goose -dir ./migrations create $$NAME sql

migration-up:
	goose -dir ./migrations postgres $(DB_URL) up

migration-down:
	goose -dir ./migrations postgres $(DB_URL) down

build:
	go build -o bin/api cmd/api/main.go
	
db-up:
	docker run --name gopg -e POSTGRES_PASSWORD=postgres -e POSTGRES_USER=postgres -e POSTGRES_DB=gopg -p 5432:5432 -d postgres