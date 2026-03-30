# PRD: TLS Profile Compliance (OCPBUGS-78774)

## Overview

This document describes the changes needed to make insights-runtime-extractor fully compliant with the OpenShift TLS security profile requirements (OCPBUGS-78774). The component must dynamically fetch and apply TLS configuration from the API Server rather than using hardcoded settings.

## Motivation

- **Compliance**: OpenShift requires all components to obey the cluster-wide TLS security profile by OCP 4.22 GA. This is a release blocker.
- **PQC Readiness**: Components must be set up for Post-Quantum Cryptography readiness by properly adhering to all aspects of the configured TLS profile.
- **Security**: All communication channels must be encrypted using the cluster's TLS policy.

### Why Not Hardcode TLS Settings

Hardcoding TLS configuration (e.g., "TLS 1.3") is explicitly **not acceptable**. The component must dynamically inherit TLS settings from the API Server configuration so that:
- Customers can adjust TLS profiles via [three documented knobs](https://docs.redhat.com/en/documentation/openshift_container_platform/4.20/html/security_and_compliance/tls-security-profiles#tls-profiles-kubernetes-configuring_tls-security-profiles) and have all components obey.
- Custom TLS profiles are respected platform-wide.
- PQC-resilient ciphers can be adopted in one pass.

---

## Architecture

### TLS Endpoints

There are two TLS endpoints to secure:

| # | Connection | Current State | TLS Role | TLS Source |
|---|-----------|---------------|----------|------------|
| 1 | Exporter HTTP server (port 8000) | Plain HTTP | TLS Server | API Server TLS profile |
| 2 | Extractor TCP server (port 3000) | Plain TCP | TLS Server | API Server TLS profile |

### Request Flow

```
HTTPS Client (8000) --> Exporter (Go, TLS Server) --> Extractor Server (Rust, TLS Server) --> Coordinator
                        TLS from API Server profile      TLS from API Server profile
```

### TLS Configuration Source

**Endpoint 1 (Exporter HTTP Server)**: Uses the **API Server configuration** (default source). The exporter fetches `apiservers.config.openshift.io/cluster` and applies the TLS profile (min version, cipher suites) to its HTTPS server. This is the externally-facing endpoint that the TLS scanner validates.

**Endpoint 2 (Extractor TCP Server)**: Uses the **API Server configuration** (default source). The extractor fetches `apiservers.config.openshift.io/cluster` at startup and applies the TLS profile (min version, cipher suites) to its TLS server. Certificate and key files are mounted from a Kubernetes Secret. The extractor does not watch for profile changes -- the exporter handles that by exiting the process, which causes the DaemonSet controller to restart the entire pod.

### No Backward Compatibility

Plain HTTP and plain TCP are **removed**. TLS is mandatory on both endpoints. The manifests must always provide TLS certificates.

---

## Planned Changes

### 1. Extractor Server (Rust) - TLS Server on TCP

**File**: `extractor/src/bin/extractor_server.rs`

#### CLI Arguments

Replace the current `Args` struct with mandatory TLS arguments:

```rust
#[derive(Parser, Debug)]
#[command(about, long_about = None)]
struct Args {
    #[arg(short, long, help = "Log level")]
    log_level: Option<String>,

    #[arg(long, required = true, help = "Path to TLS certificate file (PEM format)")]
    tls_cert: String,

    #[arg(long, required = true, help = "Path to TLS private key file (PEM format)")]
    tls_key: String,
}
```

Both `--tls-cert` and `--tls-key` are required. The server will not start without them.

#### TLS Library

Use `openssl` crate (not `rustls`) for FIPS compliance via system OpenSSL. The OpenShift TLS profile uses OpenSSL cipher names natively, so no cipher name conversion is needed.

- FIPS compliance: system OpenSSL provides FIPS-validated cryptography
- `GOEXPERIMENT=strictfipsruntime` is already used for the Go exporter, signaling FIPS is a requirement
- OpenSSL cipher names from the TLS profile can be used directly without conversion
- The base image (`base-rhel9`) already includes OpenSSL

#### TLS Initialization

At startup, the extractor fetches the API Server TLS profile and applies it to the `SslAcceptor`:

1. Use the `kube` crate to fetch the `APIServer` resource named `cluster` from `config.openshift.io/v1`
2. Resolve the TLS profile spec (handle built-in profiles: Old, Intermediate, Modern, and Custom; default to Intermediate if unset)
3. Create an `SslAcceptor` with `SslMethod::tls_server()`
4. Set `min_proto_version` from the resolved profile's `minTLSVersion`
5. Set `cipher_list` from the resolved profile's `ciphers` (OpenSSL names used directly, no conversion needed)
6. Load certificate chain and private key from the paths provided via CLI arguments
7. For each accepted `TcpStream`, perform a TLS handshake using the configured `SslAcceptor`

The extractor does **not** watch for profile changes. The exporter handles this by exiting the process when the TLS profile changes, causing the DaemonSet controller to restart the entire pod (both containers).

#### Refactor `handle_trigger_extraction`

Change the function signature to accept `impl Read + Write` instead of `TcpStream` so it works with `SslStream`. Move the TCP shutdown call to the caller.

#### Dependencies

**File**: `extractor/Cargo.toml`

Add:
```toml
openssl = "0.10"
kube = { version = "0.98", features = ["client", "openssl-tls"], default-features = false }
k8s-openapi = { version = "0.24", features = ["latest"] }
tokio = { version = "1", features = ["rt"] }
```

The `kube` client is used for the one-time API Server fetch at startup. Since the extractor is synchronous, `tokio::runtime::Runtime::block_on()` is used for the async fetch.

Run `cargo vendor` to update `extractor/vendor/` for the offline build.

#### Containerfile

**File**: `Containerfile-extractor`

Add `openssl-devel` and `perl` to build dependencies for the `openssl` crate:
```dockerfile
RUN dnf -y install gcc make wget rust-toolset rustfmt openssl-devel perl
```

---

### 2. Exporter (Go) - TLS Server on HTTP + TLS Client to Extractor

**File**: `exporter/cmd/exporter/main.go`

The exporter has two TLS roles:
- **TLS Server**: Serves HTTPS on port 8000 with the API Server TLS profile
- **TLS Client**: Connects to the extractor over TLS on port 3000

#### CLI Flags

Replace the current `-bind` flag:

```go
bindAddress := flag.String("bind", "0.0.0.0", "Bind address")
tlsCert     := flag.String("tls-cert", "", "Path to TLS certificate file (PEM format)")
tlsKey      := flag.String("tls-key", "", "Path to TLS private key file (PEM format)")
tlsCAFile      := flag.String("tls-ca", "", "Path to CA certificate for extractor TLS verification")
tlsServerName  := flag.String("tls-server-name", "", "Server name for extractor TLS certificate verification")
```

All TLS flags are required. The exporter will not start without them.

#### TLS Server (HTTPS)

Fetch the TLS profile from the API Server and apply it to the HTTPS server:

1. Use in-cluster Kubernetes client configuration to connect to the API Server
2. Fetch `apiservers.config.openshift.io/cluster` using the `openshift/api` client library
3. Resolve the TLS profile spec (handle built-in profiles: Old, Intermediate, Modern, and Custom)
4. Convert OpenSSL cipher names to Go `crypto/tls` constants
5. Build a `tls.Config` with `MinVersion` and `CipherSuites` from the profile
6. Use `http.Server.ListenAndServeTLS()` with the profile-derived `tls.Config`

For the profile resolution and cipher mapping, use the "Direct Go" approach from the compliance guidance since this component is not a controller-runtime or library-go operator.

#### TLS Client (to Extractor)

For the connection to the extractor TCP server:

1. Load the CA certificate from `--tls-ca` flag
2. Create a `tls.Config` with the CA cert pool and `ServerName` from `--tls-server-name`
3. Use `tls.Dial("tcp", EXTRACTOR_ADDRESS, tlsConfig)` instead of `net.Dial()`

The `--tls-server-name` flag is needed because the exporter dials `127.0.0.1:3000` but the certificate is typically issued for a DNS name. Go's TLS client verifies the certificate's Subject Alternative Names against the `ServerName`, not the dial address.

Go's `tls.Conn` supports `CloseWrite()`, so the existing EOF signaling protocol works unchanged.

#### Watch for Profile Changes

Implement the exit-on-change pattern: watch the `APIServer` resource for TLS profile changes and exit the process. The DaemonSet controller will restart the pod with the new profile.

#### Dependencies

**File**: `exporter/go.mod`

Add:
```
github.com/openshift/api
github.com/openshift/client-go
```

These are needed to fetch and parse the `APIServer` TLS profile configuration.

---

### 3. Kubernetes Manifests

**File**: `manifests/insights-runtime-extractor.yaml`

Update the DaemonSet spec:

```yaml
containers:
  - name: exporter
    command:
      - /exporter
      - -bind
      - 0.0.0.0
      - --tls-cert
      - /tls/tls.crt
      - --tls-key
      - /tls/tls.key
      - --tls-ca
      - /tls/tls.crt
      - --tls-server-name
      - localhost
    volumeMounts:
      - mountPath: /data
        name: data-volume
      - mountPath: /tls
        name: tls-certs
        readOnly: true
  - name: extractor
    command:
      - /extractor_server
      - --tls-cert
      - /tls/tls.crt
      - --tls-key
      - /tls/tls.key
    volumeMounts:
      - mountPath: /crio.sock
        name: crio-socket
      - mountPath: /data
        name: data-volume
      - mountPath: /tls
        name: tls-certs
        readOnly: true
volumes:
  - name: crio-socket
    hostPath:
      path: /run/crio/crio.sock
      type: Socket
  - name: data-volume
    emptyDir: {}
  - name: tls-certs
    secret:
      secretName: extractor-tls
```

Both containers mount the same TLS secret. The exporter uses the cert as both its server certificate and the CA to verify the extractor's certificate (same cert for pod-internal communication).

**File**: `manifests/insights-runtime-extractor-scc.yaml`

Add RBAC permissions for the exporter's service account to read the `APIServer` resource:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: insights-runtime-extractor-tls
rules:
  - apiGroups: ["config.openshift.io"]
    resources: ["apiservers"]
    verbs: ["get", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: insights-runtime-extractor-tls
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: insights-runtime-extractor-tls
subjects:
  - kind: ServiceAccount
    name: insights-runtime-extractor-sa
    namespace: openshift-insights
```

---

## Protocols

### HTTPS (Exporter, port 8000)

The HTTP API is unchanged. The only difference is TLS encryption:

| Method | Path | Description |
|--------|------|-------------|
| GET | `/gather_runtime_info?hash=true\|false` | Scan all containers |
| POST | `/gather_runtime_info` | Scan specific containers (JSON body with `containerIds`) |

### TLS TCP (Extractor, port 3000)

The TCP protocol is unchanged. TLS adds an encryption layer:

| Direction | Format | Example |
|-----------|--------|---------|
| Request | Plain text (read until EOF) | `abc123,def456` |
| Response | Plain text (path + newline) | `data/out-1234567890\n` |

EOF is signaled by `CloseWrite()` (Go) which sends a TLS `close_notify` alert.

---

## Testing

### Extractor Tests (Rust)

Unit and integration tests for the Rust TLS server:

1. **TLS handshake**: Verify the extractor accepts TLS connections with valid certificates
2. **TLS rejection**: Verify connections without TLS are rejected
3. **Certificate validation**: Verify the server loads cert and key correctly, fails on invalid/missing files
4. **Protocol over TLS**: Verify the existing read-until-EOF / respond-with-path protocol works identically over TLS
5. **Mandatory arguments**: Verify the server exits with error if `--tls-cert` or `--tls-key` is missing

### Exporter Tests (Go)

Unit and integration tests for the Go TLS implementation:

1. **TLS profile fetch**: Verify the exporter correctly fetches and parses the API Server TLS profile (Old, Intermediate, Modern, Custom)
2. **Cipher mapping**: Verify OpenSSL cipher names are correctly mapped to Go `crypto/tls` constants
3. **TLS version mapping**: Verify `VersionTLS12`, `VersionTLS13` etc. map to correct `tls.VersionTLS*` constants
4. **HTTPS server**: Verify the HTTP server starts with TLS and serves requests over HTTPS
5. **TLS client to extractor**: Verify the exporter connects to the extractor over TLS with CA verification
6. **Profile change detection**: Verify the exporter exits when the API Server TLS profile changes
7. **Custom profile**: Verify a custom TLS profile (specific min version + cipher list) is correctly applied
8. **Mandatory arguments**: Verify the exporter exits with error if TLS flags are missing

### E2E Tests

Update existing e2e tests in `exporter/test/e2e/`:

1. **HTTPS client**: Update test helpers to use HTTPS instead of HTTP when connecting to the exporter
2. **TLS certificate provisioning**: Test setup must create TLS secrets in the test namespace
3. **Profile compliance**: Verify the exporter only accepts connections matching the configured TLS profile
4. **Existing test scenarios**: All existing e2e tests (Java, Go, Node.js, Python, Quarkus, Spring Boot, Tomcat, JBoss EAP, WildFly, sidecar, OS) continue to pass over HTTPS

---

## Acceptance Criteria

From the [TLS Profile Compliance Remediation Guidance](https://redhat.atlassian.net/browse/OCPBUGS-78774):

- [ ] All hardcoded TLS configurations (protocols, ciphers) have been removed from the codebase
- [ ] Both the exporter and extractor fetch and apply the TLS policy from the API Server configuration
- [ ] A tls-scanner re-scan confirms the exporter HTTPS endpoint is fully compliant with the cluster TLS policy
- [ ] The service remains stable, functional, and accessible to legitimate clients after changes are deployed
- [ ] The component explicitly respects all TLS profile settings (does not rely on Go defaults)
- [ ] Functional testing confirms the component accepts only permitted TLS profile settings (including custom profiles)
- [ ] The component is set up for PQC readiness by properly adhering to all aspects of the configured TLS profile
- [ ] The exporter exits and restarts when the cluster TLS profile is updated

---

## Security Considerations

- Private key files must be mounted read-only from Kubernetes Secrets
- No client certificate authentication (server-only TLS for both endpoints)
- Certificate lifecycle management (rotation, expiry) is out of scope -- handled by external tooling (e.g., cert-manager, service-ca operator)
- Never use `InsecureSkipVerify` for the exporter-to-extractor TLS client
- The exporter must explicitly set all TLS configuration from the profile rather than relying on Go defaults
- Golang 1.24+ is required for MLKEM (PQC) support -- current go.mod already uses Go 1.25

---

## Design Decisions

1. **`openssl` over `rustls` for Rust**: this PRD uses `openssl` for FIPS compliance. The base image (`base-rhel9`) includes system OpenSSL with FIPS-validated cryptography, and OpenSSL cipher names from the TLS profile can be used directly.

2. **No backward compatibility**: Plain TCP and plain HTTP are removed entirely. TLS is mandatory. This simplifies the code (no dual-mode logic) and ensures compliance.

3. **Both endpoints fetch the API Server TLS profile**: Both the exporter and extractor dynamically fetch the TLS profile from the API Server at startup. Only the exporter watches for profile changes (exit-on-change pattern); the extractor relies on the pod restart triggered by the exporter to pick up profile updates.

4. **Exit-on-change pattern for profile updates**: When the cluster TLS profile changes, the exporter exits. The DaemonSet controller restarts the pod with the new configuration. This is the recommended pattern from the compliance guidance.

5. **Direct Go approach**: The exporter uses the "Direct Go" approach (fetch APIServer, build `tls.Config` manually) rather than controller-runtime or library-go patterns, since it is not an operator.

6. **Same certificate for both endpoints**: Both containers in the pod share the same TLS secret volume. The exporter uses the certificate as both its server cert and the CA for verifying the extractor.

---

## Files Modified

| File | Changes |
|------|---------|
| `extractor/src/bin/extractor_server.rs` | Add mandatory TLS args, fetch API Server TLS profile, TLS initialization with `openssl` using profile settings, refactor to generic `Read + Write` |
| `extractor/Cargo.toml` | Add `openssl`, `kube`, `k8s-openapi`, `tokio` dependencies |
| `extractor/vendor/` | Vendor new dependencies |
| `Containerfile-extractor` | Add `openssl-devel` and `perl` build dependencies |
| `exporter/cmd/exporter/main.go` | Add TLS flags, fetch API Server profile, HTTPS server, TLS client to extractor, profile change watcher |
| `exporter/go.mod` | Add `github.com/openshift/api`, `github.com/openshift/client-go` |
| `manifests/insights-runtime-extractor.yaml` | Add TLS secret volume, TLS CLI args for both containers |
| `manifests/insights-runtime-extractor-scc.yaml` | Add ClusterRole/ClusterRoleBinding for APIServer read access |
| `exporter/test/e2e/` | Update e2e tests for HTTPS |

---

## Implementation Order

1. Add `openssl` dependency to Rust extractor, implement TLS server, refactor `handle_trigger_extraction` to generic `Read + Write`
2. Add API Server profile fetching to Go exporter (new dependencies, cipher mapping, profile resolution)
3. Convert exporter HTTP server to HTTPS with profile-derived `tls.Config`
4. Convert exporter TCP client to TLS client with CA verification
5. Add profile change watcher (exit-on-change)
6. Update Kubernetes manifests (TLS secret volume, CLI args, RBAC)
7. Update Containerfile for `openssl-devel` build dependency
8. Update e2e tests for HTTPS
