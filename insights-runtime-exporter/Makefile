.DEFAULT_GOAL := test

clean:
	rm -f ./exporter

build: clean
	go build -o exporter cmd/exporter/main.go

test: build
	go test -v ./...

run: build
	./exporter