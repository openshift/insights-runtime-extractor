.DEFAULT_GOAL = build-image

build-rust-dev-image:
	./scripts/build-dev.sh

rust-dev:
	./scripts/rust-dev.sh

build-image:
	./scripts/build.sh

e2e-test: build-image
	cd exporter && make e2e-test