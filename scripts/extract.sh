if [[ -z "${NAMESPACE}"  ]]; then
  NAMESPACE="default"
fi

pods=$(kubectl get pods -n $NAMESPACE  --selector=app.kubernetes.io/name=insights-runtime-extractor --no-headers  -o custom-columns=":metadata.name")

for pod in $pods; do
  echo Extracting runtime info from $pod...
  out=$(kubectl exec -n $NAMESPACE $CS_NAMESPACE $pod -c exporter -- curl -s http://127.0.0.1:8000/gather_runtime_info?hash=false)
  echo $out | jq .
done
