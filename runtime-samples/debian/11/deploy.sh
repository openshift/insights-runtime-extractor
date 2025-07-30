#!/bin/bash

IMAGE_NAME=debian:11
CONTAINER_REGISTRY=quay.io/insights-runtime-extractor-samples
IMAGE=${CONTAINER_REGISTRY}/${IMAGE_NAME}

echo "Building ${IMAGE}"
podman build --platform linux/amd64 -t ${IMAGE} .
echo "Pushing ${IMAGE}"
podman push  ${IMAGE}