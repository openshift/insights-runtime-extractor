#!/bin/bash

IMAGE=quay.io/insights-runtime-extractor-samples/plain-java:24.0.1

echo "Building & pushing ${IMAGE}"
podman build --platform linux/amd64 -t ${IMAGE} .
podman push ${IMAGE}