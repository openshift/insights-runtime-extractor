.DEFAULT_GOAL := test

clean:
	rm -rf bin

build: clean
	go build -o ./bin/exporter cmd/exporter/main.go

test: build
	go test -v ./...

run: build
	go run cmd/exporter/main.go