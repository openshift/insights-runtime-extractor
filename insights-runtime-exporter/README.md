# Insights Runtime Exporter

The `exporter` is a HTTP server that acts as the façade for the insights-runtime-extractor.

Upon request of a `GET /gather-runtime-info`, it opens a TCP connection to the `extractor` (running on `127.0.0.1:3000`)
The `extractor` replies with a directory path that contains the raw runtime information.
The `exporter` reads files in that directory and generates a JSON payload that is sent back with the HTTP response.
It then deletes the directory that it read from.

# Build

Run `make build` to build the `extractor` executable

# Test

Run `make test` to run the tests

# Design

```mermaid
sequenceDiagram
  participant http-client
  box insights-runtime-extractor
  participant exporter
  participant V as volume
  participant extractor
  end

Note over exporter: HTTP server listening on :8000
Note over extractor: TCP server bound to 127.0.0.1:3000
http-client->>exporter: GET /gather-runtime-info
activate exporter
exporter->>extractor: TCP connection to trigger the extractor
activate extractor
extractor->>extractor: extract runtime info
extractor->>V: store the info in the timestamped dir
extractor->>exporter: path of the timestamped dir
deactivate extractor
exporter->>V: read content of the dir
exporter->>exporter: generate JSON payload
exporter->>V: rmdir timestamped dir
exporter->>http-client: JSON HTTP response
deactivate exporter
```