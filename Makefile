.PHONY: build build-frontend clean dev dev-backend dev-frontend docker-build install-frontend migrate-create migrate-down migrate-up run run-db test

BINARY_NAME=peakstreak
DEV_BINARY_NAME=peakstreak-dev
BINARY_PATH=./cmd/api

FRONTEND_DIR=./frontend

DB_URL ?= "postgres://user:password@localhost:5432/peakstreak?sslmode=disable"
MIGRATE_PATH = ./migrations

run:
	@echo "Running the Go application..."
	go run ${BINARY_PATH}/main.go

dev:
	@trap 'echo "Stopping servers..."; kill 0' INT; \
	make dev-frontend & \
	make dev-backend & \
	wait

dev-backend:
	@echo "Starting backend development server with hot-reloading (air)..."
	@air

dev-frontend:
	@echo "Starting frontend development server..."
	@cd ${FRONTEND_DIR} && npm run dev

test:
	@echo "Running tests..."
	go test -v ./...

build: build-frontend
	@echo "Building the Go binary..."
	go build -o ./bin/${BINARY_NAME} ${BINARY_PATH}/main.go

clean:
	@echo "Cleaning up build artifacts..."
	rm -f ./bin/${BINARY_NAME}
	rm -f ./bin/${DEV_BINARY_NAME}
	rm -rf ${FRONTEND_DIR}/build
	rm -rf ${FRONTEND_DIR}/node_modules

install-frontend:
	@echo "Installing frontend dependencies..."
	@cd ${FRONTEND_DIR} && npm install

build-frontend:
	@echo "Building frontend for production..."
	@cd ${FRONTEND_DIR} && npm install && npm run build

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
