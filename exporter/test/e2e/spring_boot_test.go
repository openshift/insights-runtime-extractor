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
			expected := types.ContainerRuntimeInfo{
				Os:              "ubuntu",
				OsVersion:       "20.04",
				Kind:            "Java",
				KindVersion:     "17.0.12",
				KindImplementer: "Eclipse Adoptium",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Spring Boot",
					Version: "3.1.4",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}
