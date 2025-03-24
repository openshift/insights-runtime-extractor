#!/bin/bash

IMAGE_NAME=python:3.9.19-slim
CONTAINER_REGISTRY=quay.io/insights-runtime-extractor-samples
IMAGE=${CONTAINER_REGISTRY}/${IMAGE_NAME}

echo "Building ${IMAGE}"
podman build --platform linux/amd64 -t ${IMAGE} .
echo "Pushing ${IMAGE}"
podman push  ${IMAGE}