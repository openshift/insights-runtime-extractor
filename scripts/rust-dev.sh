#! /bin/sh
if [[ -z "${TARGETARCH}"  ]]; then
  TARGETARCH="arm64"
fi
# Runing a root with privileged to get all the required capabilities

podman run -it --rm -u 0 --privileged -e TARGETARCH=${TARGETARCH}  -v `pwd`/extractor:/home/dev/extractor/  -w /home/dev/extractor/ rust-dev