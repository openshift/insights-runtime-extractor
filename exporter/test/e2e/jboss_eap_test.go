package e2e

import (
	"context"
	"exporter/pkg/types"
	"fmt"
	"testing"
	"time"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

func TestJBossEAP_7_4_19(t *testing.T) {

	appName := "jboss-eap-7-4-19-app"
	containerName := "main"
	// corresponded to registry.redhat.io/jboss-eap-7/eap74-openjdk8-openshift-rhel8:7.4.19
	image := "registry.redhat.io/jboss-eap-7/eap74-openjdk8-openshift-rhel8@sha256:a2a2db8ec901d0e871291b991f9eb90e06ac2f92fa17fdd0f1bdc8f8c333f4de"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("JBoss EAP 7.4.19 from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:        "rhel",
				OsVersion: "8.10",
				Kind:      "Java",
				// older versions of Java do not have the $JAVA_HOME/release to identify their versions and implementers
				Runtimes: []types.RuntimeComponent{{
					Name:    "Red Hat JBoss Enterprise Application Platform",
					Version: "7.4.19.GA",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestJBossEAP_8_0(t *testing.T) {
	appName := "jboss-eap-8-0-app"

	feature := features.New("JBoss EAP 8.0 from Helm release").
		Setup(deployTestHelmRelease(appName)).
		Teardown(undeployTestHelmRelease(appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app.kubernetes.io/name="+appName, appName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			g.Expect(runtimeInfo.Os).Should(Ω.Equal("rhel"))
			g.Expect(runtimeInfo.OsVersion).Should(Ω.Equal("8.10"))
			g.Expect(runtimeInfo.Kind).Should(Ω.Equal("Java"))
			g.Expect(runtimeInfo.KindVersion).Should(Ω.Equal("17.0.12"))
			g.Expect(runtimeInfo.KindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(runtimeInfo.Runtimes)).Should(Ω.Equal(1))
			runtime := runtimeInfo.Runtimes[0]
			g.Expect(runtime.Name).Should(Ω.Equal("Red Hat JBoss Enterprise Application Platform"))
			g.Expect(runtime.Version).Should(Ω.HavePrefix("8.0 Update"))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func deployTestHelmRelease(appName string) func(context.Context, *testing.T, *envconf.Config) context.Context {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		g := Ω.NewWithT(t)

		manager := helm.New(c.KubeconfigFile())
		if err := manager.RunInstall(helm.WithChart("https://github.com/jbossas/eap-charts/releases/download/eap8-1.1.2/eap8-1.1.2.tgz"),
			helm.WithNamespace(namespace),
			helm.WithName(appName),
		); err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}

		client, err := c.NewClient()
		if err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}
		err = wait.For(podsReady(client, "app.kubernetes.io/name="+appName), wait.WithTimeout(time.Minute*10))
		if err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}
		fmt.Printf("Helm release ready: %s/%s\n", namespace, appName)

		return ctx
	}
}

func undeployTestHelmRelease(appName string) func(context.Context, *testing.T, *envconf.Config) context.Context {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		g := Ω.NewWithT(t)

		manager := helm.New(c.KubeconfigFile())

		if err := manager.RunUninstall(helm.WithNamespace(namespace),
			helm.WithName(appName),
		); err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}
		fmt.Printf("Helm release %s uninstalled successfully\n", appName)

		return ctx
	}
}
