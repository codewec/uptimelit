.PHONY: build run test clean

BINARY=uptimelit

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINARY) ./cmd/uptimelit

run:
	go run ./cmd/uptimelit

test:
	go test -v ./...

clean:
	rm -f $(BINARY)
