.PHONY: run build tidy test clean

run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/approval-engine ./cmd/server

tidy:
	go mod tidy

test:
	go test ./...

clean:
	rm -rf bin data/approval.db
