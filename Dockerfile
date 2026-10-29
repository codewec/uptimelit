FROM golang:1.26.2-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/uptimelit ./cmd/uptimelit

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S -g 10001 uptimelit \
    && adduser -S -D -H -u 10001 -G uptimelit uptimelit \
    && mkdir /data \
    && chown uptimelit:uptimelit /data

COPY --from=build /out/uptimelit /usr/local/bin/uptimelit
WORKDIR /data
ENV UPTIMELIT_ADDR=:8080 \
    UPTIMELIT_DB=/data/uptimelit.db
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/uptimelit"]
