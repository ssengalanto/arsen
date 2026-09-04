# ============================================================================
# Dev stage: hot-reload with Air, debug tools, mounted source
# ============================================================================
FROM golang:1.26-alpine AS dev

RUN apk add --no-cache git curl
RUN go install github.com/air-verse/air@latest

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

# ============================================================================
# Build stage: compile statically linked binary
# ============================================================================
FROM golang:1.26-alpine AS build

RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/arsen ./cmd/api

# ============================================================================
# Prod stage: minimal image, non-root user
# ============================================================================
FROM gcr.io/distroless/static-debian12 AS prod

COPY --from=build /bin/arsen /bin/arsen

USER nonroot:nonroot

EXPOSE 8080
ENTRYPOINT ["/bin/arsen"]
