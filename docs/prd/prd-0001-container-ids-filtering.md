# PRD: Container IDs Filtering for Runtime Extraction

## Overview

This document describes the changes required to support targeted container scanning via a list of container IDs. Currently, the system either scans all running containers or a single container. This enhancement will allow clients to request scanning of a specific subset of containers by providing their IDs in a POST request.

## Motivation

- **Performance**: Scanning only specific containers reduces resource usage and response time.

---

## Current Architecture

### Request Flow
```
HTTP Client → Exporter (Go) → Extractor Server (Rust) → Coordinator (Rust) → Fingerprints
             GET /gather_runtime_info    TCP 3000          /coordinator
```

### Current Capabilities
- **Exporter**: Only supports `GET /gather_runtime_info?hash=true|false`
- **Extractor Server**: Receives empty TCP payload, spawns coordinator
- **Coordinator**: Accepts optional single `--container-id <id>` argument (not used by the extractor server)

---

## Proposed Changes

### 1. Exporter HTTP Server (Go)

**File**: `exporter/cmd/exporter/main.go`

#### New Endpoint
Add a new HTTP POST endpoint alongside the existing GET endpoint:

| Method | Path | Request Body | Description |
|--------|------|--------------|-------------|
| POST | `/gather_runtime_info` | JSON object with container IDs | Scan specific containers |

#### Request Format
```json
{
  "containerIds": ["abc123...", "def456...", "ghi789..."]
}
```

#### Validation Rules
- Request body must be valid JSON
- `containerIds` field must be present and be an array
- Array can be empty (equivalent to scanning all containers)
- Each container ID must be a non-empty string
- Container IDs should be the short or full CRI-O container ID format

#### Changes Required

1. **Add JSON request struct**:
   ```go
   type GatherRuntimeInfoRequest struct {
       ContainerIds []string `json:"containerIds"`
   }
   ```

2. **Update HTTP handler**:
   - Parse request method (GET vs POST)
   - For POST: parse JSON body and validate
   - Pass container IDs to `triggerRuntimeInfoExtraction()`

3. **Update `triggerRuntimeInfoExtraction()` function**:
   - Add parameter: `containerIds []string`
   - Serialize container IDs as a single string of comma-separated container IDs before sending over TCP
   - Send string payload instead of empty payload to extractor server

4. **TCP Protocol Change**:
   - Current: Empty payload triggers extraction
   - New: String payload with a list of comma-separated container IDs
   ```
   "abc123,def456,..."
   ```
   - Empty string means "scan all containers"

#### Backward Compatibility
- GET endpoint behavior remains unchanged (scans all containers)
- GET internally sends an empty string `""` to maintain protocol consistency

---

### 2. Extractor Server (Rust)

**File**: `extractor/src/bin/extractor_server.rs`

#### Changes Required

1. **Update `handle_trigger_extraction()` function**:
   - Read incoming data from TCP stream (currently ignored) as a String
   - Pas the incoming data to the CLI argument if not empty. Wrap them in a String.
   
2. **Update coordinator invocation**:
   - Current:
     ```rust
     Command::new("/coordinator")
         .arg("--log-level")
         .arg(log_level)
         .output()
     ```
   - New (when container IDs provided):
     ```rust
     Command::new("/coordinator")
         .arg("--log-level")
         .arg(log_level)
         .arg(container_ids)
         .output()
     ```

3. **Error Handling**:
   - Invalid String: Return error response via TCP
   - Log validation errors at warn level

#### TCP Protocol

| Direction | Format | Example |
|-----------|--------|---------|
| Request | Plain text | `abc123,def456` |
| Response | Plain text (path) | `data/out-1234567890` |

---

### 3. Coordinator (Rust)

**File**: `extractor/src/bin/coordinator.rs`

#### Current Arguments
```rust
#[arg(short, long)]
log_level: Option<String>

#[arg(help = "ID of the container to scan. If absent, all containers are scanned")]
container_id: Option<String>
```

#### New Arguments
```rust
#[arg(short, long)]
log_level: Option<String>

#[arg(help = "Comma-separated list of container IDs to scan")]
container_ids: Option<String>
```

#### Changes Required

1. **Add new CLI argument**:
   - Change `container_id` to `container_ids` so that multiple containers ID can be passed (separated by commas)
   - Parse them into `Vec<String>`

2. **Update container filtering logic**:

   - Current location in coordinator.rs `main()`:
     ```rust
      let containers = match args.container_id {
        None => get_containers(),
        Some(container_id) => match get_container(&container_id) {
            Some(container) => vec![container],
            None => vec![],
        },
    };
     ```
   - New logic:
     ```rust
      let containers = match args.container_ids {         
        None => get_containers(),
        Some(container_ids) => {
         // split them from commas,
         // iterate on them: for each trimmed non-empty container ID,
         // call get_container(container_id)
         // and collect them in a Vector
        },
      };

     ```
3. **Update container information retrieval**:

   - Update `get_container(container_id: &String)` to invoke `crictl inspect $container_id` instead of `crictl ps`. 

4. **Logging**:
   - Log at info level: "Scanning X containers"

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
│  2. Validate containerIds array                                             │
│  3. Send comma-separated containerIds to extractor_server via TCP           │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                    TCP: abc123,def456
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                    EXTRACTOR_SERVER (Rust) - Port 3000                      │
│                                                                             │
│  1. Read String from TCP stream                                             │
│  2. Execute: /coordinator --log-level info "abc123,def456"                  │
└─────────────────────────────────┬───────────────────────────────────────────┘
                                  │
                       Subprocess with args
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         COORDINATOR (Rust)                                  │
│                                                                             │
│  1. Parse executable argument                                               │
│  2. Get all running containers via crictl                                   │
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
| containerIds | string[] | Yes | List of container IDs to scan. Empty array scans all containers. |

#### Response

Same as existing GET endpoint response:

```json
{
  "node-name": {
    "namespace-1": {
      "pod-name-1": {
        "container-name-1": {
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
}
```

#### Error Responses

| Status | Condition | Response Body |
|--------|-----------|---------------|
| 400 | Invalid JSON | `{"error": "Invalid JSON in request body"}` |
| 400 | Missing containerIds | `{"error": "containerIds field is required"}` |
| 400 | Invalid container ID format | `{"error": "Container ID at index N is invalid"}` |
| 500 | Extraction failed | `{"error": "Runtime extraction failed: <details>"}` |

---

## Testing Requirements

### Unit Tests

1. **Exporter**:
   - JSON parsing of valid request body
   - Validation of container ID formats
   - Empty array handling
   - Malformed JSON rejection

2. **Extractor Server**:
   - JSON deserialization from TCP stream
   - Container ID validation
   - Command argument construction

3. **Coordinator**:
   - Comma-separated parsing of the main argument
   - Container filtering logic

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
   - Should scan existing ones, log warnings for missing

### E2E Tests

Add test cases to existing e2e test suite in `runtime-samples/`:
- Test POST endpoint with subset of containers
- Verify filtering works correctly in Kubernetes environment

---

## Implementation Phases

### Phase 1: Coordinator Changes
1. Change main argument `container_ids` to `container_id`
2. Implement container filtering logic
3. Add unit tests
4. Manual testing with direct coordinator invocation

### Phase 2: Extractor Server Changes
1. Update coordinator invocation
2. Add error handling
3. Add unit tests

### Phase 3: Exporter Changes
1. Add POST endpoint handler
2. Implement JSON validation
3. Update TCP communication
4. Add unit tests

### Phase 4: Integration & E2E
1. Integration tests
2. E2E tests in runtime-samples
3. Documentation updates

---

## Backward Compatibility

| Component | Backward Compatible | Notes |
|-----------|---------------------|-------|
| Exporter | Yes | GET endpoint unchanged |
| Extractor Server | Yes | Empty payload treated as empty containerIds |
| Coordinator | Yes | passing a singular container ID still works |

---

## Security Considerations

- No new privileges required
- No shell injection risk (IDs passed as single argument, not executed)
- Logging does not expose sensitive data

---

## Open Questions

1. **Container ID format**: Should we support both short (12-char) and full (64-char) container IDs?
   - Recommendation: Yes, support both with prefix matching

2. **Maximum number of container IDs**: Should there be a limit?
   - Recommendation: No hard limit, but document performance implications

3. **Response for non-existent IDs**: Should we return an error or just skip missing containers?
   - Recommendation: Skip missing, return results for found containers, log warnings

---

## Files to Modify

| File | Changes |
|------|---------|
| `exporter/cmd/exporter/main.go` | Add POST handler, JSON parsing, update TCP payload |
| `extractor/src/bin/extractor_server.rs` | Add JSON deserialization, update coordinator args |
| `extractor/src/bin/coordinator.rs` | Change main argument, filtering logic |
