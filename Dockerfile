FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

FROM golang:1.21-alpine AS go-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/build ./frontend/build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /peakstreak ./cmd/api/main.go

FROM gcr.io/distroless/static-debian11
COPY --from=go-builder /peakstreak /peakstreak
COPY --from=go-builder /app/frontend/build /frontend/build
EXPOSE 8080
ENTRYPOINT ["/peakstreak"]
