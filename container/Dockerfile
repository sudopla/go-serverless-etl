# Multi-Stage image
FROM golang:1.17-alpine AS build_base

RUN apk add --no-cache git build-base

WORKDIR /tmp/go-ingestor

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./out/go-ingestor ./cmd/ingestor

# Start fresh from a smaller image
FROM alpine:3.9 
RUN apk add ca-certificates

COPY --from=build_base /tmp/go-ingestor/out/go-ingestor /app/go-ingestor

CMD ["/app/go-ingestor"]