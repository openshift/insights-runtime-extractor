.DEFAULT_GOAL = build-image

build-dev-image:
	./scripts/build-dev.sh

dev:
	./scripts/dev.sh

build-image:
	./scripts/build.sh
