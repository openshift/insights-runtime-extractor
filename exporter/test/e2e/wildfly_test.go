package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestWildFly_30_0_1_Final(t *testing.T) {

	appName := "wildfly-30-0-1-app"
	containerName := "main"
	// corresponded to quay.io/wildfly/wildfly:30.0.1.Final-jdk20
	image := "quay.io/wildfly/wildfly@sha256:4b51924efea743d48685eb2065c33d21f818fb9c5acc7139dc47e5b810610e3b"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("WildFly from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "9.2",
				Kind:            "Java",
				KindVersion:     "20.0.2",
				KindImplementer: "Eclipse Adoptium",
				Runtimes: []types.RuntimeComponent{{
					Name:    "WildFly Full",
					Version: "30.0.1.Final",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestWildFly_34_0_0_Final(t *testing.T) {

	appName := "wildfly-34-0-0-app"
	containerName := "main"
	// corresponded to quay.io/wildfly/wildfly:34.0.0.Final-jdk21
	image := "quay.io/wildfly/wildfly@sha256:a5f6f45da0ca7ca973f5a9974d56ae96d244163dc7a5c795be9e53da2a24aff0"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("WildFly from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "9.4",
				Kind:            "Java",
				KindVersion:     "21.0.4",
				KindImplementer: "Eclipse Adoptium",
				Runtimes: []types.RuntimeComponent{{
					Name:    "WildFly",
					Version: "34.0.0.Final",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}
