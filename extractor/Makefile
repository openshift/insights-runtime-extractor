.PHONY: build
build:
	cargo build --offline --release

clean:
	cargo clean

unit-tests: build
	cargo test

check:
	cargo fmt -- --check