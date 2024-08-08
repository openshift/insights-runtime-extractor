#!/bin/bash

IMAGE=upstream-quarkus:1.0.0-SNAPSHOT
OCP_IMAGE_REGISTRY=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
OCP_NAMESPACE=e2e-insights-runtime-extractor
OCP_IMAGE=${OCP_IMAGE_REGISTRY}/${OCP_NAMESPACE}/${IMAGE}

echo "Pushing ${OCP_IMAGE}"
podman tag ${IMAGE} ${OCP_IMAGE}
podman push --tls-verify=false ${OCP_IMAGE}
oc set image-lookup -n ${OCP_NAMESPACE} upstream-quarkus

IMAGE=upstream-native-quarkus:1.0.0-SNAPSHOT
OCP_IMAGE=${OCP_IMAGE_REGISTRY}/${OCP_NAMESPACE}/${IMAGE}
echo "Pushing ${OCP_IMAGE}"
podman tag ${IMAGE} ${OCP_IMAGE}
podman push --tls-verify=false ${OCP_IMAGE}
oc set image-lookup -n ${OCP_NAMESPACE} upstream-native-quarkus

echo "OCP Image stream"
oc get -n ${OCP_NAMESPACE} imagestream
