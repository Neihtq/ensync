FROM golang:1.26-alpine AS builder

RUN apk add --no-cache build-base alsa-lib-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build  binaries
RUN CGO_ENABLED=1 GOOS=linux go build -o grandmaster ./cmd/grandmaster/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -o follower ./cmd/follower/main.go

FROM alpine:latest

RUN apk add --no-cache alsa-lib

WORKDIR /app

COPY --from=builder /app/grandmaster .
COPY --from=builder /app/follower .

COPY --from=builder /app/static ./static

# Grandmaster: 65533 (Web UI), 65534 (NTP/Clock Sync), 65535 (Follower Registration)
# Follower: 65532 (Audio Stream), 65531 (Control Plane)
EXPOSE 65531 65532 65533 65534 65535

# Default command is grandmaster
CMD ["./grandmaster"]
