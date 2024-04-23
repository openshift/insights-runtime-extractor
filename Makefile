.DEFAULT_GOAL = build-image

build-dev-image:
	./scripts/build-dev.sh

rust-dev:
	./scripts/rust-dev.sh

build-image:
	./scripts/build.sh
