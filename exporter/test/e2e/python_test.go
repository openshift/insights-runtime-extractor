package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestPython3(t *testing.T) {

	appName := "python3-app"
	containerName := "python"
	// corresponded to python:3.9.19-slim
	image := "python@sha256:85c7a2a383a01e0b77b5f9c97d8b1eef70409a99552fde03c518a98dfa19609c"
	deployment := newPython3AppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Python3 from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("debian"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("12"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("Python"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.Equal("Python 3.9.19"))
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(0))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func newPython3AppDeployment(namespace string, name string, replicas int32, containerName string, image string) *appsv1.Deployment {
	deployment := newAppDeployment(namespace, name, replicas, containerName, image)

	deployment.Spec.Template.Spec.Containers[0].Command = []string{"python3"}
	deployment.Spec.Template.Spec.Containers[0].Args = []string{
		"-m",
		"http.server"}

	return deployment
}
