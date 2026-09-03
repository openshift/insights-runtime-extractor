.PHONY: build
build:
	cargo build --offline --release

clean:
	cargo clean

unit-tests: build
	cargo test

check:
	@result=0; \
	$(MAKE) fmt-check || result=1; \
	$(MAKE) clippy || result=1; \
	exit $$result

fmt-check:
	cargo fmt -- --check

fmt:
	cargo fmt

clippy:
	cargo clippy -- -D warnings

clippy-fix:
	cargo clippy --fix -- -D warnings

fix: clippy-fix fmt
