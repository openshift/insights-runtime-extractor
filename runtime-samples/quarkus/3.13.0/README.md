# quarkus 3.13.0

# Runtime sample for the insights-runtime-extractor

To build the container image and make it available to the OpenShift internal image
registry, run:

```shell script
make
```

# Workload Runtime Information:

For the OpenJDK image:

```json
"runtimeInfo": {
  "os": "rhel",
  "osVersion": "8.10",
  "kind": "Java",
  "kindVersion": "17.0.12",
  "kindImplementer": "Red Hat, Inc.",
  "runtimes": [ 
    {
      "name": "Quarkus",
      "version": "3.13.0"
    }
  ]
}
```

For the native image:

```json
"runtimeInfo": {
  "os": "rhel",
  "osVersion": "8.10",
  "kind": "GraalVM",
  "runtimes": [ 
    {
      "name": "Quarkus",
    }
  ]
}
```