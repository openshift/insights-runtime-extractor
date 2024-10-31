echo Using namespace for the insights-runtime-extractor: $TEST_NAMESPACE

oc new-project $TEST_NAMESPACE
oc apply -f insights-runtime-extractor-scc.yaml -n $TEST_NAMESPACE