.DEFAULT_GOAL = push

build:
	./mvnw install

# corresponded to quay.io/quarkus/ubi-quarkus-mandrel-builder-image:23.0.5.0-Final-java17-amd64
build-native:
	./mvnw clean package -DskipTests \
	-Dnative \
	-Dquarkus.native.container-build=true \
	-Dquarkus.native.container-runtime=podman \
	-Dquarkus.native.march=x86-64-v3 \
	-Dquarkus.native.builder-image=quay.io/quarkus/ubi-quarkus-mandrel-builder-image@sha256:ee78c273b9a89d8de2f57bee7cf4a57d5b654b1d2cdfe5e784d15d36ad03209e \
	-Dquarkus.container-image.name=upstream-native-quarkus

push: build build-native
	./deploy.sh
