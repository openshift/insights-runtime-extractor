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
  "os": "ubuntu",
  "osVersion": "20.04",
  "kind": "Java",
  "kindVersion": "17.0.12",
  "kindImplementer": "Eclipse Adoptium",
  "runtimes": [
    {
      "name": "Spring Boot",
      "version": "3.1.4"
    }
  ]
}
```