.PHONY: run test build migrate-up migrate-down docker-build docker-run build-frontend clean-frontend
.PHONY: run test build migrate-up migrate-down docker-build docker-run install-frontend build-frontend dev-frontend clean

BINARY_NAME=peakstreak
BINARY_PATH=./cmd/api

FRONTEND_DIR=./frontend

DB_URL ?= "postgres://user:password@localhost:5432/peakstreak?sslmode=disable"
MIGRATE_PATH = ./migrations

run:
	@echo "Running the Go application..."
	go run ${BINARY_PATH}/main.go

dev: build-frontend
	@echo "Running the Go application with hot-reloaded frontend..."
	@# In one terminal, run `make dev-frontend`. In another, run `make run`.
	@echo "Please run 'make dev-frontend' in a separate terminal for frontend hot-reloading."
	go run ${BINARY_PATH}/main.go

test:
	@echo "Running tests..."
	go test -v ./...

build: build-frontend
	@echo "Building the Go binary..."
	go build -o ./bin/${BINARY_NAME} ${BINARY_PATH}/main.go

clean:
	@echo "Cleaning up build artifacts..."
	rm -f ./bin/${BINARY_NAME}
	rm -rf ${FRONTEND_DIR}/build
	rm -rf ${FRONTEND_DIR}/node_modules

install-frontend:
	@echo "Installing frontend dependencies..."
	@cd ${FRONTEND_DIR} && npm install

build-frontend:
	@echo "Building frontend for production..."
	@cd ${FRONTEND_DIR} && npm install && npm run build

dev-frontend:
	@echo "Starting frontend development server..."
	@cd ${FRONTEND_DIR} && npm run dev

migrate-create:
	@echo "Creating migration file: ${name}"
	@migrate create -ext sql -dir ${MIGRATE_PATH} -seq ${name}

migrate-up:
	@echo "Applying migrations..."
	@migrate -database "${DB_URL}" -path ${MIGRATE_PATH} up

migrate-down:
	@echo "Reverting migrations..."
	@migrate -database "${DB_URL}" -path ${MIGRATE_PATH} down

docker-build: build-frontend
	@echo "Building Docker image..."
	@docker build -t peakstreak:latest .

run-db:
	@podman run --name peakstreak-db -e POSTGRES_USER=user -e POSTGRES_PASSWORD=password -e POSTGRES_DB=peakstreak -p 5432:5432 -d postgres:15-alpine
