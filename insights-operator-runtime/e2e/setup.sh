echo Using namespace for the container scanner: $TEST_NAMESPACE

oc new-project $TEST_NAMESPACE
oc apply -f /tmp/insights-operator-runtime-pull-secret.yaml -n $TEST_NAMESPACE
oc apply -f insights-operator-runtime-scc.yaml -n $TEST_NAMESPACE

