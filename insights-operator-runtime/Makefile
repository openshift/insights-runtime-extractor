.PHONY: build
build: export RUSTFLAGS=-C target-feature=+crt-static
build:
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