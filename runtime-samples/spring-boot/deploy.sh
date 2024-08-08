#!/bin/bash

IMAGE=spring-boot:0.0.1-SNAPSHOT
OCP_IMAGE_REGISTRY=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
OCP_NAMESPACE=e2e-insights-runtime-extractor
OCP_IMAGE=${OCP_IMAGE_REGISTRY}/${OCP_NAMESPACE}/${IMAGE}

echo "Pushing ${OCP_IMAGE}"
podman build --platform linux/amd64 -t ${OCP_IMAGE} .
podman push --tls-verify=false ${OCP_IMAGE}
oc set image-lookup -n ${OCP_NAMESPACE} spring-boot

echo "OCP Image stream"
oc get -n ${OCP_NAMESPACE} imagestream
