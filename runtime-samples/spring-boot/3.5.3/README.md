# spring-boot

# Runtime sample for the insights-runtime-extractor

To build the container image and make it available to the OpenShift internal image
registry, run:

```shell script
make
```

# Workload Runtime Information:


```json
"runtimeInfo": {
  "os": "rhle",
  "osVersion": "8.10",
  "kind": "Java",
  "kindVersion": "21.0.6",
  "kindImplementer": "Red Hat, Inc.",
  "runtimes": [
    {
      "name": "Spring Boot",
      "version": "3.5.3"
    }
  ]
}
```