# PRD: Container IDs Filtering for Runtime Extraction

## Overview

This document describes the changes implemented to support targeted container scanning via a list of container IDs. Previously, the system only scanned all running containers. This enhancement allows clients to request scanning of a specific subset of containers by providing their IDs in a POST request.

## Motivation

- **Performance**: Scanning only specific containers reduces resource usage and response time.

---

## Architecture

### Request Flow
```
HTTP Client → Exporter (Go) → Extractor Server (Rust) → Coordinator (Rust) → Fingerprints
             GET/POST /gather_runtime_info    TCP 3000          /coordinator
```

### Capabilities
- **Exporter**: Supports `GET /gather_runtime_info?hash=true|false` (scans all) and `POST /gather_runtime_info` (scans specific containers)
- **Extractor Server**: Receives comma-separated container IDs via TCP, spawns coordinator with IDs as argument
- **Coordinator**: Accepts optional `container_ids` positional argument (comma-separated list)

---

## Implemented Changes

### 1. Exporter HTTP Server (Go)

**File**: `exporter/cmd/exporter/main.go`

#### Endpoints

| Method | Path | Request Body | Description |
|--------|------|--------------|-------------|
| GET | `/gather_runtime_info` | None | Scan all containers |
| POST | `/gather_runtime_info` | JSON object with container IDs | Scan specific containers |

#### Request Format (POST)
```json
{
  "containerIds": ["abc123...", "def456...", "ghi789..."]
}
```

#### Validation Rules
- Request body must be valid JSON
- `containerIds` field must be present and be an array
- Array can be empty (equivalent to scanning all containers)
- Container IDs must be in full CRI-O format, with or without `cri-o://` prefix

#### Implementation

1. **JSON request struct**:
   ```go
   type GatherRuntimeInfoRequest struct {
       ContainerIds []string `json:"containerIds"`
   }
   ```

2. **HTTP handler** (`gatherRuntimeInfo`):
   - For GET: uses empty `containerIds` slice
   - For POST: decodes JSON body and validates `containerIds` field is present
   - Passes container IDs to `triggerRuntimeInfoExtraction()`

3. **`triggerRuntimeInfoExtraction()` function**:
   - Accepts `containerIds []string` parameter
   - Joins container IDs with commas: `strings.Join(containerIds, ",")`
   - Sends payload via TCP
   - **Calls `tcpConn.CloseWrite()` to signal EOF** - this is critical because the Rust server uses `read_to_string()` which reads until EOF

#### TCP Protocol
- Client sends comma-separated container IDs as plain text
- Client calls `CloseWrite()` to signal end of data
- Server responds with path to extracted data

#### Backward Compatibility
- GET endpoint behavior unchanged (scans all containers)
- GET internally sends an empty string `""` to maintain protocol consistency

---

### 2. Extractor Server (Rust)

**File**: `extractor/src/bin/extractor_server.rs`

#### Implementation

1. **`handle_trigger_extraction()` function**:
   - Reads incoming data from TCP stream using `read_to_string()` (reads until EOF)
   - Trims the received string
   - Passes container IDs as positional argument to coordinator

2. **Coordinator invocation**:
   ```rust
   Command::new("/coordinator")
       .arg("--log-level")
       .arg(log_level)
       .arg(container_ids)  // comma-separated string, passed even if empty
       .output()
   ```

#### TCP Protocol

| Direction | Format | Example |
|-----------|--------|---------|
| Request | Plain text (read until EOF) | `abc123,def456` |
| Response | Plain text (path) | `data/out-1234567890\n` |

---

### 3. Coordinator (Rust)

**File**: `extractor/src/bin/coordinator.rs`

#### CLI Arguments
```rust
#[arg(short, long)]
log_level: Option<String>

#[arg(help = "Comma-separated list of container IDs to scan. If absent, all containers are scanned")]
container_ids: Option<String>
```

#### Implementation

1. **Parsing container IDs**:
   ```rust
   let container_ids: Vec<String> = args
       .container_ids
       .map(|ids| {
           ids.split(',')
               .map(|id| id.trim().to_string())
               .filter(|id| !id.is_empty())
               .collect()
       })
       .unwrap_or_default();
   ```

2. **Container retrieval**:
   - Calls `get_containers(container_ids)` passing the parsed vector
   - Empty vector means "scan all containers"

3. **Logging**:
   - Logs "Scanning X containers" at info level

---

### 4. Container Module (Rust)

**File**: `extractor/src/insights_runtime_extractor/container.rs`

#### Implementation

1. **`get_containers(container_ids: Vec<String>)` function**:
   - Normalizes container IDs by stripping `cri-o://` prefix if present
   - Runs `crictl ps -o json -s RUNNING` to get all running containers
   - If `container_ids` is not empty, filters to only include matching containers
   - Returns `Vec<Container>`

---

## Data Flow Diagram

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              HTTP Client                                    │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                POST /gather_runtime_info
                Body: {"containerIds": ["abc123", "def456"]}
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        EXPORTER (Go) - Port 8000                            │
│                                                                             │
│  1. Parse POST body as JSON                                                 │
│  2. Validate containerIds field is present                                  │
│  3. Send comma-separated containerIds to extractor_server via TCP           │
│  4. Call CloseWrite() to signal EOF                                         │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                    TCP: "abc123,def456" + EOF
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                    EXTRACTOR_SERVER (Rust) - Port 3000                      │
│                                                                             │
│  1. Read String from TCP stream until EOF (read_to_string)                  │
│  2. Execute: /coordinator --log-level info "abc123,def456"                  │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                       Subprocess with args
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         COORDINATOR (Rust)                                  │
│                                                                             │
│  1. Parse comma-separated container IDs from argument                       │
│  2. Get all running containers via crictl ps                                │
│  3. Filter to only requested containers (if any specified)                  │
│  4. Scan filtered containers                                                │
│  5. Output path to stdout                                                   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## API Specification

### POST /gather_runtime_info

#### Request

**Headers**:
```
Content-Type: application/json
```

**Body**:
```json
{
  "containerIds": ["container-id-1", "container-id-2"]
}
```

**Field Descriptions**:
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| containerIds | string[] | Yes | List of container IDs to scan. Empty array scans all containers. Supports full CRI-O IDs, with or without `cri-o://` prefix. |

#### Response

Same as existing GET endpoint response:

```json
{
  "namespace-1": {
    "pod-name-1": {
      "container-id-1": {
        "os": "rhel",
        "osVersion": "8.9",
        "kind": "Java",
        "kindVersion": "17.0.9",
        "kindImplementer": "Red Hat, Inc.",
        "runtimes": [
          {"name": "Quarkus", "version": "3.2.0"}
        ]
      }
    }
  }
}
```

#### Error Responses

| Status | Condition | Response Body |
|--------|-----------|---------------|
| 400 | Invalid JSON | `{"error": "Invalid JSON in request body"}` |
| 400 | Missing containerIds | `{"error": "containerIds field is required"}` |
| 500 | Extraction failed | Error message from extraction process |

---

### Integration Tests

1. **End-to-end with specific containers**:
   - POST request with known container IDs
   - Verify only those containers appear in response

2. **Empty array behavior**:
   - POST with `{"containerIds": []}` should scan all containers
   - Result should match GET endpoint

3. **Non-existent container IDs**:
   - POST with IDs that don't match running containers
   - Should return empty/partial results gracefully

4. **Mixed valid/invalid IDs**:
   - Some IDs exist, some don't
   - Should scan existing ones only

### E2E Tests

Add test cases to existing e2e test suite in `runtime-samples/`:
- Test POST endpoint with subset of containers
- Verify filtering works correctly in Kubernetes environment

---

## Backward Compatibility

| Component | Backward Compatible | Notes |
|-----------|---------------------|-------|
| Exporter | Yes | GET endpoint unchanged |
| Extractor Server | Yes | Empty string treated as "scan all" |
| Coordinator | Yes | Empty/missing argument scans all containers |

---

## Security Considerations

- No new privileges required
- No shell injection risk (IDs passed as single argument, not executed)
- Logging does not expose sensitive data

---

## Design Decisions

1. **Container ID format**: Full (64-char) container IDs are supported. The `cri-o://` prefix is automatically stripped if present.

2. **Maximum number of container IDs**: No hard limit imposed.

3. **Response for non-existent IDs**: Missing containers are silently skipped. Results are returned for found containers only.

4. **TCP EOF signaling**: The Go exporter uses `CloseWrite()` to signal EOF because the Rust server uses `read_to_string()` which reads until EOF.

5. **Filtering approach**: Instead of calling `crictl inspect` for each container ID, the implementation fetches all running containers with `crictl ps` and filters in memory. This is simpler and avoids multiple subprocess calls.

---

## Files Modified

| File | Changes |
|------|---------|
| `exporter/cmd/exporter/main.go` | Add POST handler, JSON parsing, TCP EOF signaling with CloseWrite() |
| `extractor/src/bin/extractor_server.rs` | Read container IDs from TCP stream, pass to coordinator |
| `extractor/src/bin/coordinator.rs` | Accept container_ids positional argument, parse and pass to get_containers |
| `extractor/src/insights_runtime_extractor/container.rs` | Update get_containers to accept Vec<String> and filter containers |
