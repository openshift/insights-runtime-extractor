.DEFAULT_GOAL = push

build:
	./mvnw install

# corresponded to quay.io/quarkus/ubi-quarkus-mandrel-builder-image:24.1.0.0-Final-java23-amd64
build-native:
	./mvnw clean package -DskipTests \
	-Dnative \
	-Dquarkus.native.container-build=true \
	-Dquarkus.native.container-runtime=podman \
	-Dquarkus.native.march=x86-64-v3 \
	-Dquarkus.native.builder-image=quay.io/quarkus/ubi-quarkus-mandrel-builder-image@sha256:4ed5878e0ea4e51c00fe9ab0618a9c803dc5bc582e4677664706cb750973ebd6 \
	-Dquarkus.container-image.image=quay.io/insights-runtime-extractor-samples/native-quarkus:3.24.4


push: build build-native