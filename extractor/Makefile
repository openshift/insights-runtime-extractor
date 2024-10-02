PATH := "$(HOME)/.cargo/bin:$(PATH)"

.PHONY: build
build: export RUSTFLAGS=-C target-feature=+crt-static
build: check
ifeq ($(TARGETARCH),arm64)
	rustup target add aarch64-unknown-linux-musl
	cargo build --release --target aarch64-unknown-linux-musl
else
	rustup target add x86_64-unknown-linux-musl
	cargo build --release --target x86_64-unknown-linux-musl
endif

clean:
	cargo clean

unit-tests: build
	cargo test

check:
	cargo fmt -- --check