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

clippy:
	cargo clippy -- -D warnings
