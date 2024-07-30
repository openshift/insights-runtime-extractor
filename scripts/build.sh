if [[ -z "${IMAGE_REGISTRY}"  ]]; then
  IMAGE_REGISTRY="quay.io/jmesnil"
fi
if [[ -z "${TAG}"  ]]; then
  TAG=latest
fi

for component in "extractor" "exporter" ; do
  IMG=$IMAGE_REGISTRY/insights-runtime-$component:$TAG

  podman manifest rm $IMG
  podman manifest create $IMG

  podman build --platform linux/arm64,linux/amd64 --manifest $IMG --target=$component .

  if [ $? -ne 0 ]; then
    echo "❌ Building image failed"
    exit 1
  fi

  echo "=========================================================================="
  echo "✅ Building image $IMG"
  echo "=========================================================================="

  podman manifest push $IMG

  echo "=========================================================================="
  echo "✅ Pushing image $IMG"
  echo "=========================================================================="

done
