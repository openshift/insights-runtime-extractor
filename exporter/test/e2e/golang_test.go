package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestGolang(t *testing.T) {

	appName := "golang-app"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/golang-app:1.22.6"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Golang application "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("debian"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("12"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("Golang"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.Equal("go1.22.6"))
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(0))
		}))
	_ = testenv.Test(t, feature.Feature())
}
