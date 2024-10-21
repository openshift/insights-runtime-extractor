# quarkus 3.15.1

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
  "kindVersion": "21.0.5",
  "kindImplementer": "Red Hat, Inc.",
  "runtimes": [ 
    {
      "name": "Quarkus",
      "version": "3.15.1"
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
      "version": "3.15.1"
    }
  ]
}
```