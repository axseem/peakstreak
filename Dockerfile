FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the application
# -ldflags="-w -s" strips debug information, reducing binary size
# CGO_ENABLED=0 is important for a static binary for scratch/distroless images
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /peakstreak ./cmd/api/main.go

FROM gcr.io/distroless/static-debian11

COPY --from=builder /peakstreak /peakstreak

EXPOSE 8080

ENTRYPOINT ["/peakstreak"]
