#!/bin/bash

IMAGE=quay.io/insights-runtime-extractor-samples/rhbq-app:3.27.0

echo "Building & pushing ${IMAGE}"
podman build --platform linux/amd64 -t ${IMAGE} .
podman push ${IMAGE}