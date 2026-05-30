# syntax=docker/dockerfile:1.7

# ---------- build stage ----------
FROM golang:1.25-alpine AS build
WORKDIR /src

RUN apk add --no-cache git

# Cache dependencies first
COPY go.mod go.sum* ./
RUN go mod download

# Copy source and build a static binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/barterswap .

# ---------- runtime stage ----------
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /app

COPY --from=build /out/barterswap /app/barterswap
COPY migrations /app/migrations
COPY api /app/api

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -fsS http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/barterswap"]
