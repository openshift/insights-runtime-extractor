IMG=rust-dev

echo "=========================================================================="
echo "Build multi arch image $IMG"
echo "=========================================================================="

podman build --platform linux/arm64,linux/amd64 -f Containerfile-rust-dev -t $IMG .
