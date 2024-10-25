package e2e

import (
	"exporter/pkg/types"
	"testing"

	appsv1 "k8s.io/api/apps/v1"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestUbi9Minimal(t *testing.T) {
	testBaseImage(t, "registry.access.redhat.com/ubi9-minimal:9.4", "rhel", "9.4")
}

func TestUbi8Minimal(t *testing.T) {
	testBaseImage(t, "registry.access.redhat.com/ubi8/ubi-minimal:8.10", "rhel", "8.10")
}

func TestDebian(t *testing.T) {
	testBaseImage(t, "debian:12", "debian", "12")
}

func TestCentOs7(t *testing.T) {
	testBaseImage(t, "centos:7", "centos", "7")
}

func testBaseImage(t *testing.T, baseImage string, expectedOs string, expectedOsVersion string) {
	appName := envconf.RandomName("os", 10)
	containerName := "main"
	deployment := newBaseImageDeployment(namespace, appName, 1, containerName, baseImage)

	feature := features.New("base image "+baseImage).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {

			expected := types.ContainerRuntimeInfo{
				Os:        expectedOs,
				OsVersion: expectedOsVersion,
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func newBaseImageDeployment(namespace string, name string, replicas int32, containerName string, image string) *appsv1.Deployment {
	deployment := newAppDeployment(namespace, name, replicas, containerName, image)

	deployment.Spec.Template.Spec.Containers[0].Command = []string{"tail", "-f", "/dev/null"}

	return deployment
}
