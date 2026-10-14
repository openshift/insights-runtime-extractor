.PHONY: build
build:
	cargo build --offline --release

clean:
	cargo clean

unit-tests: build
	cargo test

privileged-tests: build
	cargo test --no-run --lib 2>&1
	sudo cargo test --lib privilege -- --ignored --nocapture --test-threads=1

tests: unit-tests privileged-tests

check:
	@result=0; \
	$(MAKE) fmt-check || result=1; \
	$(MAKE) clippy || result=1; \
	exit $$result

fmt-check:
	cargo fmt -- --check

clippy:
	cargo clippy -- -D warnings
