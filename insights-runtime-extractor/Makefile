.PHONY: build
build: export RUSTFLAGS=-C target-feature=+crt-static
build: check
ifeq ($(TARGETARCH),arm64)
	cargo build --release --target aarch64-unknown-linux-musl
else
	cargo build --release --target x86_64-unknown-linux-musl
endif

.PHONY: clean
clean:
	cargo clean

.PHONY: test
test:
	cargo test

.PHONY: check
check:
	cargo fmt -- --check

export TEST_NAMESPACE=test-openshift-insights-00001

e2e-setup:
	export TEST_NAMESPACE=${TEST_NAMESPACE} && cd e2e && ./setup.sh

e2e-test: e2e-setup
	export TEST_NAMESPACE=${TEST_NAMESPACE} && cd e2e && go test -count=1 -v .

clean-e2e:
	oc delete namespace ${TEST_NAMESPACE}