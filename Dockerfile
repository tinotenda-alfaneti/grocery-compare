# syntax=docker/dockerfile:1.6

FROM node:20-bookworm AS web-builder
WORKDIR /web
COPY apps/web/package*.json ./
RUN npm ci
COPY apps/web ./
RUN npm run build

FROM golang:1.24-bookworm AS api-builder
WORKDIR /app

# Install build dependencies for SQLite (CGO), matching homelabsite's proven pattern.
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    libc6-dev \
    && rm -rf /var/lib/apt/lists/*

COPY apps/api/go.mod apps/api/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY apps/api ./
RUN --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o /out/grocery-compare .

FROM gcr.io/distroless/base-debian12
WORKDIR /srv

COPY --from=api-builder /out/grocery-compare /usr/local/bin/grocery-compare
COPY --from=web-builder /web/dist /srv/web

ENV APP_ENV=production
ENV HTTP_PORT=8080
ENV WEB_ROOT=/srv/web
ENV DB_PATH=/data/grocery.db

EXPOSE 8080
VOLUME ["/data"]

ENTRYPOINT ["/usr/local/bin/grocery-compare"]
