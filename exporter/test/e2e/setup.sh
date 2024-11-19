go install  -mod=readonly  github.com/jstemmer/go-junit-report/v2@latest

oc delete project $TEST_NAMESPACE
kubectl wait --for=delete namespace/$TEST_NAMESPACE --timeout=60s

oc delete project e2e-insights-runtime-extractor
kubectl wait --for=delete namespace/e2e-insights-runtime-extractor --timeout=60s

echo Using namespace for the insights-runtime-extractor: $TEST_NAMESPACE

oc new-project $TEST_NAMESPACE
oc apply -f test/e2e/insights-runtime-extractor-scc.yaml -n $TEST_NAMESPACE

