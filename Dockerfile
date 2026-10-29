# syntax=docker/dockerfile:1.7

FROM node:26.13.0-alpine3.24 AS frontend
WORKDIR /src/frontend
RUN npm install --global pnpm@12.16.1
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN --mount=type=cache,id=uptimelit-pnpm,target=/root/.local/share/pnpm/store pnpm install --frozen-lockfile
COPY frontend/ ./
RUN pnpm run build

FROM golang:1.27.3-alpine3.24 AS backend
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY backend/ ./
COPY --from=frontend /src/frontend/dist/ ./cmd/server/static/
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/uptimelit ./cmd/server

FROM alpine:3.24.3 AS runtime
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 uptimelit \
    && adduser -S -D -H -u 10001 -G uptimelit uptimelit \
    && mkdir /data \
    && chown uptimelit:uptimelit /data
COPY --from=backend /out/uptimelit /usr/local/bin/uptimelit
COPY LICENSE /usr/share/licenses/uptimelit/LICENSE
COPY frontend/LICENSE.dashlit /usr/share/licenses/uptimelit/LICENSE.dashlit
WORKDIR /data
ENV UPTIMELIT_ADDR=:8080 \
    UPTIMELIT_DB=/data/uptimelit.db
USER 10001:10001
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -q -T 4 -O /dev/null http://127.0.0.1:8080/api/auth/config || exit 1
ENTRYPOINT ["/usr/local/bin/uptimelit"]
