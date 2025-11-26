.DEFAULT_GOAL = push

build:
	./mvnw install

# corresponded to quay.io/quarkus/ubi-quarkus-mandrel-builder-image:23.1.9.0-Final-java21-amd64
build-native:
	./mvnw clean package -DskipTests \
	-Dnative \
	-Dquarkus.native.container-build=true \
	-Dquarkus.native.container-runtime=podman \
	-Dquarkus.native.march=x86-64-v3 \
	-Dquarkus.native.builder-image=quay.io/quarkus/ubi-quarkus-mandrel-builder-image@sha256:39f81544bb9f1c6b663fc73e93400887f5b084ebf03c12e50785b5883a444fca \
	-Dquarkus.container-image.image=quay.io/insights-runtime-extractor-samples/native-quarkus:3.29.2


push: build build-native