package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestSpringBoot(t *testing.T) {

	appName := "spring-boot"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/spring-boot:3.1.4"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Spring Boot from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("ubuntu"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("20.04"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("Java"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.Equal("17.0.12"))
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.Equal("Eclipse Adoptium"))

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(1))
			runtime := runtimeInfo.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Spring Boot"))
			g.Expect(runtime.Version).To(Ω.Equal("3.1.4"))
		}))
	_ = testenv.Test(t, feature.Feature())
}
