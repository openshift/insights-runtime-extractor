package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestJava_17_0_13(t *testing.T) {

	appName := "plain-java"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/plain-java:0.0.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Plain Java from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "17.0.13",
				KindImplementer: "Red Hat, Inc.",
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestJava_24_0_1(t *testing.T) {

	appName := "plain-java"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/plain-java:24.0.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Plain Java from "+image).
	Setup(deployTestResource(deployment, appName)).
	Teardown(undeployTestResource(deployment, appName)).
	Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
		expected := types.ContainerRuntimeInfo{
			Os:              "alpine",
			OsVersion:       "3.21.4",
			Kind:            "Java",
			KindVersion:     "24.0.1",
			KindImplementer: "Eclipse Adoptium",
		}
		g.Expect(runtimeInfo).Should(Ω.Equal(expected))
	}))
	_ = testenv.Test(t, feature.Feature())
}

func TestJava_25_0_1(t *testing.T) {

	appName := "plain-java"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/plain-java:25.0.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Plain Java from "+image).
	Setup(deployTestResource(deployment, appName)).
	Teardown(undeployTestResource(deployment, appName)).
	Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
		expected := types.ContainerRuntimeInfo{
			Os:              "rhel",
			OsVersion:       "10.0",
			Kind:            "Java",
			KindVersion:     "25.0.1",
			KindImplementer: "Eclipse Adoptium",
		}
		g.Expect(runtimeInfo).Should(Ω.Equal(expected))
	}))
	_ = testenv.Test(t, feature.Feature())
}
