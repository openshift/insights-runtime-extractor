# PRD: TLS Support for Extractor TCP Server

## Overview

This document describes the changes needed to add TLS encryption to the TCP communication between the Exporter (Go) and the Extractor Server (Rust). TLS is configured by providing a private key and certificate file. When these are not provided, the server runs plain TCP for backward compatibility.

## Motivation

- **Security**: Encrypt communication between the exporter and extractor to protect data in transit, even within the same pod.
- **Compliance**: Meet security requirements for encrypted inter-process communication in production deployments.

---

## Architecture

### Request Flow
```
HTTP Client → Exporter (Go) → Extractor Server (Rust) → Coordinator (Rust) → Fingerprints
             GET/POST /gather_runtime_info   TLS/TCP 3000        /coordinator
```

### TLS Configuration

| Component | TLS Role | Configuration |
|-----------|----------|---------------|
| Extractor Server (Rust) | TLS Server | `--tls-cert` and `--tls-key` CLI arguments |
| Exporter (Go) | TLS Client | `--tls-cert` CLI flag (CA/server cert for verification) |

Both sides are configured in lockstep. If the Rust server has TLS enabled, the Go exporter must also be configured with TLS.

### TLS Library

**Rust**: `rustls` (pure-Rust TLS implementation) with the `ring` crypto backend.

Rationale:
- ADR-0002 mandates static compilation and minimal dependencies
- `rustls` compiles statically without needing a system OpenSSL library
- The `ring` backend avoids `aws-lc-rs` C build complexity
- Memory-safe, well-audited Rust crate

**Go**: Standard library `crypto/tls` package (no additional dependencies).

---

## Planned Changes

### 1. Extractor Server (Rust)

**File**: `extractor/src/bin/extractor_server.rs`

#### CLI Arguments

Add two optional arguments to the existing `Args` struct:

```rust
#[arg(long, help = "Path to TLS certificate file (PEM format)")]
tls_cert: Option<String>,

#[arg(long, help = "Path to TLS private key file (PEM format)")]
tls_key: Option<String>,
```

Validation: both must be provided together or neither. Exit with error if only one is given.

#### TLS Initialization

When both `--tls-cert` and `--tls-key` are provided:

1. Load PEM certificate and private key files using `rustls-pemfile`
2. Build a `rustls::ServerConfig` with no client authentication
3. Wrap each accepted `TcpStream` in a `rustls::StreamOwned` before handling

When neither is provided, accept plain TCP connections as before.

#### Refactor `handle_trigger_extraction`

Change the function to accept `impl Read + Write` instead of `TcpStream` so it works with both plain TCP and TLS streams. Move the `stream.shutdown(Shutdown::Both)` call to the caller level (TCP-specific).

---

### 2. Rust Dependencies

**File**: `extractor/Cargo.toml`

Add:
```toml
rustls = { version = "0.23", default-features = false, features = ["ring", "logging", "std", "tls12"] }
rustls-pemfile = "2"
```

Run `cargo vendor` to update `extractor/vendor/` for the offline build.

---

### 3. Exporter (Go)

**File**: `exporter/cmd/exporter/main.go`

#### CLI Flag

Add a `--tls-cert` flag specifying the path to the CA or server certificate for TLS verification.

#### Connection Logic

In `triggerRuntimeInfoExtraction()`:

- If `--tls-cert` is set: load the certificate, create a `tls.Config` with the cert pool, and use `tls.Dial("tcp", EXTRACTOR_ADDRESS, tlsConfig)`
- If not set: use `net.Dial("tcp", EXTRACTOR_ADDRESS)` as before

Go's `tls.Conn` supports `CloseWrite()`, so the existing EOF signaling protocol works unchanged.

---

### 4. Containerfile

**File**: `Containerfile-extractor`

May need to add `perl` to the build dependencies for the `ring` crate's build scripts:
```dockerfile
RUN dnf -y install gcc make wget rust-toolset rustfmt perl
```

---

### 5. Kubernetes Manifests

**File**: `manifests/insights-runtime-extractor.yaml`

When TLS is enabled:

```yaml
containers:
  - name: extractor
    command:
      - /extractor_server
      - --tls-cert
      - /tls/tls.crt
      - --tls-key
      - /tls/tls.key
    volumeMounts:
      - mountPath: /tls
        name: tls-certs
        readOnly: true
  - name: exporter
    command:
      - /exporter
      - --tls-cert
      - /tls/tls.crt
    volumeMounts:
      - mountPath: /tls
        name: tls-certs
        readOnly: true
volumes:
  - name: tls-certs
    secret:
      secretName: extractor-tls
```

---

## TCP Protocol

The protocol is unchanged. TLS adds an encryption layer but the application-level data flow remains the same:

| Direction | Format | Example |
|-----------|--------|---------|
| Request | Plain text (read until EOF) | `abc123,def456` |
| Response | Plain text (path) | `data/out-1234567890\n` |

EOF is signaled by `CloseWrite()` (Go) which sends a TLS `close_notify` alert. The Rust side detects EOF on `read_to_string()`.

---

## Backward Compatibility

| Component | Backward Compatible | Notes |
|-----------|---------------------|-------|
| Extractor Server | Yes | Plain TCP when `--tls-cert`/`--tls-key` absent |
| Exporter | Yes | Plain TCP when `--tls-cert` absent |
| Coordinator | N/A | No changes needed |
| Kubernetes Manifests | Yes | TLS volume mount is optional |

---

## Security Considerations

- Private key files must be mounted read-only from Kubernetes Secrets
- No client certificate authentication (server-only TLS)
- Certificate lifecycle management (rotation, expiry) is out of scope — handled by external tooling (e.g., cert-manager)
- `rustls` and `ring` are well-audited, memory-safe Rust crates

---

## Design Decisions

1. **`rustls` over OpenSSL**: Pure Rust, static compilation, no system library dependency. Aligns with ADR-0002.

2. **`ring` crypto backend**: Avoids `aws-lc-rs` C build dependency. `ring` includes some C/ASM but is well-supported.

3. **No separate TLS toggle**: Both sides are configured in lockstep. The presence of cert/key parameters enables TLS — no separate enable/disable flag.

4. **No client certificate authentication**: The extractor and exporter run in the same pod. Server-side TLS is sufficient for encrypting the channel.

5. **Generic stream handling**: Refactoring `handle_trigger_extraction` to `impl Read + Write` keeps the code clean and avoids duplicating logic for TLS vs plain TCP.

---

## Files Modified

| File | Changes |
|------|---------|
| `extractor/src/bin/extractor_server.rs` | Add TLS CLI args, TLS initialization, refactor to generic `Read + Write` |
| `extractor/Cargo.toml` | Add `rustls` and `rustls-pemfile` dependencies |
| `extractor/vendor/` | Vendor new dependencies for offline build |
| `exporter/cmd/exporter/main.go` | Add `--tls-cert` flag, use `tls.Dial()` when configured |
| `Containerfile-extractor` | Add `perl` build dependency if needed for `ring` |
| `manifests/insights-runtime-extractor.yaml` | Add TLS secret volume mount and CLI args |

---

## Implementation Order

1. Refactor `handle_trigger_extraction` to generic `Read + Write` (independent, testable change)
2. Add `rustls`/`rustls-pemfile` to `Cargo.toml` and vendor
3. Add TLS initialization and accept logic to `extractor_server.rs`
4. Update Go exporter with TLS client support
5. Update Containerfile and Kubernetes manifests
6. Test with self-signed certificates in both TLS and non-TLS modes
