package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestJBossEAP_(t *testing.T) {

	appName := "jboss-eap-7-4-19-app"
	containerName := "main"
	// corresponded to registry.redhat.io/jboss-eap-7/eap74-openjdk8-openshift-rhel8:7.4.19
	image := "registry.redhat.io/jboss-eap-7/eap74-openjdk8-openshift-rhel8@sha256:a2a2db8ec901d0e871291b991f9eb90e06ac2f92fa17fdd0f1bdc8f8c333f4de"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("JBoss EAP 7.4.19 from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:        "rhel",
				OsVersion: "8.10",
				Kind:      "Java",
				// older versions of Java do not have the $JAVA_HOME/release to identify their versions and implementers
				Runtimes: []types.RuntimeComponent{
					types.RuntimeComponent{
						Name:    "Red Hat JBoss Enterprise Application Platform",
						Version: "7.4.19.GA",
					},
				},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}
