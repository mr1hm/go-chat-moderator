FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build all binaries
RUN CGO_ENABLED=0 go build -o bin/api ./cmd/api
RUN CGO_ENABLED=0 go build -o bin/migrate ./cmd/migrate
RUN CGO_ENABLED=0 go build -o bin/moderation ./cmd/moderation-service

# Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/bin ./bin
COPY entrypoint.sh .
RUN chmod +x entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["./entrypoint.sh"]
