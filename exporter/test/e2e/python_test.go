package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestPython_2_7_18(t *testing.T) {

	appName := "python-2-7-18-app"
	containerName := "python-2-7-18"
	// corresponded to python:2.7.18-slim
	image := "quay.io/insights-runtime-extractor-samples/python@sha256:1cbf6240295dc583767410f74a0ec8f0fd80f640cfb0b44921fa6712ac7ccda0"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Python2 from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:          "debian",
				OsVersion:   "10",
				Kind:        "Python",
				KindVersion: "Python 2.7.18",
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestPython3_3_9_19(t *testing.T) {

	appName := "python3-3-9-19-app"
	containerName := "python-3-9-19"
	// corresponded to python:3.9.19-slim
	image := "quay.io/insights-runtime-extractor-samples/python@sha256:c66cd8a43c55ba2291230670f3681b5dcdcf284e7499a2819e0349a393798239"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Python3 from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:          "debian",
				OsVersion:   "12",
				Kind:        "Python",
				KindVersion: "Python 3.9.19",
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestPython3_3_13_2(t *testing.T) {

	appName := "python3-3-13-2-app"
	containerName := "python-3-13-2"
	// corresponded to python:3.13.2-slim
	image := "quay.io/insights-runtime-extractor-samples/python@sha256:09ae40a77fda2b4150bc9ffa99e8fcb10e935aa752209c9a7cd674ded8bd8d30"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Python3 from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:          "debian",
				OsVersion:   "12",
				Kind:        "Python",
				KindVersion: "Python 3.13.2",
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

