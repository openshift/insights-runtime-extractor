.DEFAULT_GOAL := test

clean:
	rm -rf ./bin

build: clean
	go build -o ./bin/fpr_native_executable cmd/fpr_native_executable/main.go

test: build
	go test -v ./...