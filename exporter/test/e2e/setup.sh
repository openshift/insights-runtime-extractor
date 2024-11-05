go install  -mod=readonly  github.com/jstemmer/go-junit-report/v2@latest

echo Using namespace for the insights-runtime-extractor: $TEST_NAMESPACE

oc new-project $TEST_NAMESPACE
oc apply -f test/e2e/insights-runtime-extractor-scc.yaml -n $TEST_NAMESPACE