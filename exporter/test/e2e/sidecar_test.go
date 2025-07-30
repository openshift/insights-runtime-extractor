package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestSidecarContainer(t *testing.T) {

	appName := "sidecar-container"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/rhbq-app:3.15.3"
	initContainerName := "logshipper"
	initImage := "quay.io/insights-runtime-extractor-samples/centos:7"
	initCommand := []string{"sh", "-c", "tail -F /opt/logs.txt"}

	deployment := newSidecarContainerDeployment(namespace, appName, 1, containerName, image, initContainerName, initImage, initCommand)

	featureMain := features.New("Sidecar container deployment. Main image: "+image+" Init image: "+initImage).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted from main container", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "17.0.13",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Quarkus",
					Version: "3.15.3.redhat-00002",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		})).
		Assess("runtime info extracted from init container", checkExtractedRuntimeInfo(namespace, "app="+appName, initContainerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "centos",
				OsVersion:       "7",
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, featureMain.Feature())
}
