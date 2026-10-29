FROM golang:1.26.2-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/tinypulse ./cmd/tinypulse

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 tinypulse \
    && adduser -S -D -H -u 10001 -G tinypulse tinypulse \
    && mkdir /data \
    && chown tinypulse:tinypulse /data

COPY --from=build /out/tinypulse /usr/local/bin/tinypulse
WORKDIR /data
ENV TINYPULSE_ADDR=:8080 \
    TINYPULSE_DB=/data/tinypulse.db
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/tinypulse"]
