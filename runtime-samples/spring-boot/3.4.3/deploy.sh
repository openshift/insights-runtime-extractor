#!/bin/bash

IMAGE=quay.io/insights-runtime-extractor-samples/spring-boot:3.4.3

echo "Building & pushing ${IMAGE}"
podman build --platform linux/amd64 -t ${IMAGE} .
podman push ${IMAGE}