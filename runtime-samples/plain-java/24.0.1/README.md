# plain-java

# Runtime sample for the insights-runtime-extractor

To build the container image and make it available to the OpenShift internal image
registry, run:

```shell script
make
```

# Workload Runtime Information:


```json
"runtimeInfo": {
  "os": "ubuntu",
  "osVersion": "24.04",
  "kind": "Java",
  "kindVersion": "24.0.1",
  "kindImplementer": "Eclipse Adoptium",
  "runtimes": [
  ]
}
```