.DEFAULT_GOAL = build-image
.PHONY: runtime-samples

build-rust-dev-image:
	./scripts/build-dev.sh

rust-dev:
	./scripts/rust-dev.sh

build-image:
	./scripts/build.sh

e2e-test: build-image
	cd exporter && make e2e-test

runtime-samples:
	cd runtime-samples && make