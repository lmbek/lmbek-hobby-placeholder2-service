# Build Stage
FROM golang:1.26-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o service main.go

# Minimal Production Stage
FROM alpine:3.22
RUN apk upgrade --no-cache && adduser -D -u 10001 appuser
WORKDIR /app
COPY --from=builder /app/service .
USER 10001:10001
EXPOSE 8081
ENTRYPOINT ["/app/service"]
