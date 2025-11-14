# nodejs-app
# Runtime sample for the insights-runtime-extractor

To build the container image and make it available to the OpenShift internal image
registry, run:

```shell script
make
```

# Workload Runtime Information:

```json
"runtimeInfo": {
  "os": "alpine",
  "osVersion": "3.22.1",
  "kind": "Node.js",
  "kindVersion": "v25.1.0",
}
```