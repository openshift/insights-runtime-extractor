#! /bin/sh
if [[ -z "${TARGETARCH}"  ]]; then
  TARGETARCH="arm64"
fi
# Runing a root with privileged to get all the required capabilities

podman run -it --rm -u 0 --privileged -e TARGETARCH=${TARGETARCH}  -v `pwd`/insights-operator-runtime:/home/dev/insights-operator-runtime/  -w /home/dev/insights-operator-runtime/ rust-dev