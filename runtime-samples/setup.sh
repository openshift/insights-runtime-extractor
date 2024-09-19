oc project e2e-insights-runtime-extractor || oc new-project e2e-insights-runtime-extractor
oc patch configs.imageregistry.operator.openshift.io/cluster --type merge -p '{"spec":{"defaultRoute":true}}'
OCP_REGISTRY=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
oc registry login --registry=$OCP_REGISTRY
