package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestTomcat(t *testing.T) {

	appName := "tomcat"
	containerName := "main"
	// Tomcat image that corresponded to tomcat:11.0-jre21
	image := "quay.io/insights-runtime-extractor-samples/tomcat@sha256:505c9fcae748e2e0488c8aa13794f041088b4771739aa906ef761b124474c874"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Tomcat from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "ubuntu",
				OsVersion:       "24.04",
				Kind:            "Java",
				KindVersion:     "21.0.5",
				KindImplementer: "Eclipse Adoptium",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Apache Tomcat",
					Version: "11.0.1",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestJBossWebServer(t *testing.T) {

	appName := "jboss-webserver"
	containerName := "main"
	image := "registry.redhat.io/jboss-webserver-5/jws58-openjdk17-openshift-rhel8@sha256:fd96fec3eb328060fe7cc7d4f23ad27976103d16b98932e999a6103caa42d0a7"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("JBoss WebServer 5.8 from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "17.0.12",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Apache Tomcat",
					Version: "9.0.87.redhat-00003",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}
