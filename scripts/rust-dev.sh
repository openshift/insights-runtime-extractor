#! /bin/sh
if [[ -z "${TARGETARCH}"  ]]; then
  TARGETARCH="arm64"
fi
# Runing a root with privileged to get all the required capabilities

podman run -it --rm -u 0 --privileged \
  -e TARGETARCH=${TARGETARCH} \
  -v `pwd`:/opt/app-root/src/insights-runtime-extractor \
  -w /opt/app-root/src/insights-runtime-extractor/extractor \
  rust-dev /bin/bash