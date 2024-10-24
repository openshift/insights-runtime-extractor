package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestQuarkus_3_13_0(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/quarkus:3.13.0"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("rhel"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("8.10"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("Java"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.Equal("17.0.12"))
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(1))
			runtime := runtimeInfo.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			g.Expect(runtime.Version).To(Ω.Equal("3.13.0"))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestNativeQuarkus_3_13_0(t *testing.T) {

	appName := "native-quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/native-quarkus:3.13.0"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Native Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("rhel"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("8.9"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("GraalVM"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.BeEmpty())
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(1))
			runtime := runtimeInfo.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			// In native mode, Quarkus does not expose its version
			g.Expect(runtime.Version).To(Ω.BeEmpty())
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestQuarkus_3_15_1(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/quarkus:3.15.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("rhel"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("8.10"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("Java"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.Equal("21.0.4"))
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(1))
			runtime := runtimeInfo.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			g.Expect(runtime.Version).To(Ω.Equal("3.15.1"))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestNativeQuarkus_3_15_1(t *testing.T) {

	appName := "native-quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/native-quarkus:3.15.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Native Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("rhel"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("8.10"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("GraalVM"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.BeEmpty())
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(1))
			runtime := runtimeInfo.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			// In native mode, Quarkus started reporting its version since 3.15.x
			g.Expect(runtime.Version).To(Ω.Equal("3.15.1"))
		}))
	_ = testenv.Test(t, feature.Feature())
}
