.DEFAULT_GOAL := unit-tests

clean:
	rm -rf ./bin

build: clean
	go build -o ./bin/fpr_java_jboss_modules cmd/fpr_java_jboss_modules/main.go
	go build -o ./bin/fpr_java_runtimes cmd/fpr_java_runtimes/main.go
	go build -o ./bin/fpr_java_version cmd/fpr_java_version/main.go
	go build -o ./bin/fpr_kind_executable cmd/fpr_kind_executable/main.go
	go build -o ./bin/fpr_native_executable cmd/fpr_native_executable/main.go
	go build -o ./bin/fpr_os cmd/fpr_os/main.go

unit-tests: build
	go test -v ./...