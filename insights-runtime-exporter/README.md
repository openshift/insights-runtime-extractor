# Insights Runtime Exporter

The exporter is a HTTP server that acts as the facade for the insights-runtime-extractor.

Upon request of a `GET /gather-runtime-info`, it opens a TCP connection to the `extractor` (running on `localhost:3000`)
The `extractor` replies with a directory path that contains the raw runtime information.
The `exported` reads files in that directory and generates a JSON payload that is sent back with the HTTP response
It then deletes the directory that it read from.

# Build

Run `make build` to build the `extractor` executable

# Test

Run `make test` to run the tests

# Design

```mermaid
sequenceDiagram
  participant http-client
  box insights-runtime-extractor [pod]
  participant exporter
  participant V as volume
  participant extractor
  end

Note over exporter: HTTP server listening on :8000
Note over extractor: TCP server bound to 127.0.0.1:3000
http-client->>exporter: GET /gather-runtime-info
activate exporter
exporter->>extractor: TCP trigger an extraction
extractor->>exporter: path of the timestamped dir
exporter->>V: read content of the dir
exporter->>exporter: generate JSON payload
exporter->>V: rmdir timestamped dir
exporter->>insights-operator: JSON HTTP response
deactivate exporter
```