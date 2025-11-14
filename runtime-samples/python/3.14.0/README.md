# python-app
# Runtime sample for the insights-runtime-extractor

To build the container image and make it available to the OpenShift internal image
registry, run:

```shell script
make
```

# Workload Runtime Information:

```json
"runtimeInfo": {
  "os": "debian",
  "osVersion": "13",
  "kind": "Python",
  "kindVersion": "Python 3.14.0",
}
```