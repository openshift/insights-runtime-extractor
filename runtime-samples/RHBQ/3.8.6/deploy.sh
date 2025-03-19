#!/bin/bash

IMAGE=quay.io/insights-runtime-extractor-samples/rhbq-app:3.8.6

echo "Building & pushing ${IMAGE}"
podman build --platform linux/amd64 -t ${IMAGE} .
podman push ${IMAGE}