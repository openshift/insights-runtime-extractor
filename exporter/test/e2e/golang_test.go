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
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:          "debian",
				OsVersion:   "12",
				Kind:        "Golang",
				KindVersion: "go1.22.6",
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}
