.DEFAULT_GOAL := unit-tests

PATH=$(shell printenv PATH):$(HOME)/bin

clean:
	rm -rf bin

build: clean
	go build -o ./bin/exporter cmd/exporter/main.go

unit-tests: build
	go test -v ./pkg/...

run: build
	go run cmd/exporter/main.go

install-helm:  ## Install Helm toolchain for 3rd party integration
	./hack/install-helm.sh

export TEST_NAMESPACE=test-openshift-insights-00001
e2e-setup: install-helm
	./test/e2e/setup.sh

export TEST_NAMESPACE=test-openshift-insights-00001
e2e-test: e2e-setup
	./test/e2e/test.sh

clean-e2e:
	oc delete namespace ${TEST_NAMESPACE}
