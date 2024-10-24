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
	// Tomcat image that corresponded to 11.0-jre21
	image := "tomcat@sha256:3b353d4a30c315ae1177b04642c932e07fcd87155452629d4809c38d9854de72"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Tomcat from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("ubuntu"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("24.04"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("Java"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.Equal("21.0.4"))
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.Equal("Eclipse Adoptium"))

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(1))
			runtime := runtimeInfo.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Apache Tomcat"))
			g.Expect(runtime.Version).To(Ω.Equal("11.0.0-M22"))
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
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("rhel"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("8.10"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("Java"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.Equal("17.0.12"))
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(runtimeInfo.Runtimes)).To(Ω.Equal(1))
			runtime := runtimeInfo.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Apache Tomcat"))
			g.Expect(runtime.Version).To(Ω.Equal("9.0.87.redhat-00003"))
		}))
	_ = testenv.Test(t, feature.Feature())
}
