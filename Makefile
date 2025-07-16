.PHONY: run test build migrate-up migrate-down docker-build docker-run

# Go configuration
BINARY_NAME=peakstreak-api
BINARY_PATH=./cmd/api

# Database configuration
DB_URL ?= "postgres://user:password@localhost:5432/peakstreak?sslmode=disable"
MIGRATE_PATH = ./migrations

run:
	@echo "Running the application..."
	go run ${BINARY_PATH}/main.go

test:
	@echo "Running tests..."
	go test -v ./...

build:
	@echo "Building the binary..."
	go build -o ./bin/${BINARY_NAME} ${BINARY_PATH}/main.go

migrate-create:
	@echo "Creating migration file: ${name}"
	@migrate create -ext sql -dir ${MIGRATE_PATH} -seq ${name}

migrate-up:
	@echo "Applying migrations..."
	@migrate -database "${DB_URL}" -path ${MIGRATE_PATH} up

migrate-down:
	@echo "Reverting migrations..."
	@migrate -database "${DB_URL}" -path ${MIGRATE_PATH} down

docker-build:
	@echo "Building Docker image..."
	@docker build -t habit-tracker:latest .

run-db:
	@docker run --name peakstreak -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=peakstreak -p 5432:5432 -d postgres:15-alpine
