package e2e

import (
	"context"
	"exporter/pkg/types"
	"fmt"
	"testing"
	"time"

	Ω "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"sigs.k8s.io/e2e-framework/third_party/helm"
)

func TestJBossEAP_7_2_9(t *testing.T) {

	appName := "jboss-eap-7-2-9-app"
	containerName := "main"
	// corresponded to registry.redhat.io/jboss-eap-7/eap72-openjdk11-openshift-rhel8:1.2
	image := "registry.redhat.io/jboss-eap-7/eap72-openjdk11-openshift-rhel8@sha256:8e87b56c7d4d0b31839cff135b05d7b6198c7884fd3c6831b2db6e9e8736560b"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("JBoss EAP 7.2 from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.2",
				Kind:            "Java",
				KindVersion:     "11.0.8",
				KindImplementer: "N/A",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Red Hat JBoss Enterprise Application Platform",
					Version: "7.2.9.GA",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestJBossEAP_7_3_10(t *testing.T) {

	appName := "jboss-eap-7-3-10-app"
	containerName := "main"
	// corresponded to registry.redhat.io/jboss-eap-7/eap73-openjdk11-openshift-rhel8:7.3.10
	image := "registry.redhat.io/jboss-eap-7/eap73-openjdk11-openshift-rhel8@sha256:b57e133fef5f2eb38b037e663704f11fc28ce683665e50023ddd998b9e33238a"
	deployment := newEAP7AppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("JBoss EAP 7.3.10 from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.5",
				Kind:            "Java",
				KindVersion:     "11.0.13",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Red Hat JBoss Enterprise Application Platform",
					Version: "7.3.10.GA",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestJBossEAP_7_4_19(t *testing.T) {

	appName := "jboss-eap-7-4-19-app"
	containerName := "main"
	// corresponded to registry.redhat.io/jboss-eap-7/eap74-openjdk8-openshift-rhel8:7.4.19
	image := "registry.redhat.io/jboss-eap-7/eap74-openjdk8-openshift-rhel8@sha256:a2a2db8ec901d0e871291b991f9eb90e06ac2f92fa17fdd0f1bdc8f8c333f4de"
	deployment := newEAP7AppDeployment(namespace, appName, 1, containerName, image)

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
		Setup(deployTestHelmRelease_8_0(appName)).
		Teardown(undeployTestHelmRelease(appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app.kubernetes.io/name="+appName, appName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "17.0.14",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Red Hat JBoss Enterprise Application Platform",
					Version: "8.0 Update 6.1",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestJBossEAP_8_1(t *testing.T) {
	appName := "jboss-eap-8-1-app"

	feature := features.New("JBoss EAP 8.1 from Helm release").
	Setup(deployTestHelmRelease_8_1(appName)).
	Teardown(undeployTestHelmRelease(appName)).
	Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app.kubernetes.io/name="+appName, appName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
		expected := types.ContainerRuntimeInfo{
			Os:              "rhel",
			OsVersion:       "9.6",
			Kind:            "Java",
			KindVersion:     "17.0.14",
			KindImplementer: "Red Hat, Inc.",
			Runtimes: []types.RuntimeComponent{{
				Name:    "Red Hat JBoss Enterprise Application Platform",
				Version: "8.1",
			}},
		}
		g.Expect(runtimeInfo).Should(Ω.Equal(expected))
	}))
	_ = testenv.Test(t, feature.Feature())
}

func deployTestHelmRelease_8_0(appName string) func(context.Context, *testing.T, *envconf.Config) context.Context {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		g := Ω.NewWithT(t)

		manager := helm.New(c.KubeconfigFile())
		if err := manager.RunInstall(helm.WithChart("https://github.com/jbossas/eap-charts/releases/download/eap8-1.1.2/eap8-1.1.2.tgz"),
			helm.WithNamespace(namespace),
			helm.WithName(appName),
			helm.WithArgs("--set", "build.uri=https://github.com/openshift/insights-runtime-extractor"),
			helm.WithArgs("--set", "build.contextDir=runtime-samples/jboss-eap/8.0"),
			helm.WithArgs("--set", "build.s2i.jdk17.runtimeImage=registry.redhat.io/jboss-eap-8/eap8-openjdk17-runtime-openshift-rhel8@sha256:d16346922c00c0430b3a3a6b3a47c910a23063a8a679a5c1b2d32fd96b92a225"),
		); err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}

		client, err := c.NewClient()
		if err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}
		err = wait.For(podsReady(client, namespace, "app.kubernetes.io/name="+appName), wait.WithTimeout(time.Minute*10))
		if err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}
		fmt.Printf("Helm release ready: %s/%s\n", namespace, appName)

		return ctx
	}
}

func deployTestHelmRelease_8_1(appName string) func(context.Context, *testing.T, *envconf.Config) context.Context {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		g := Ω.NewWithT(t)

		manager := helm.New(c.KubeconfigFile())
		if err := manager.RunInstall(helm.WithChart("https://github.com/jbossas/eap-charts/releases/download/eap81-1.0.2/eap81-1.0.2.tgz"),
			helm.WithNamespace(namespace),
					     helm.WithName(appName),
					     helm.WithArgs("--set", "build.uri=https://github.com/openshift/insights-runtime-extractor"),
					     helm.WithArgs("--set", "build.contextDir=runtime-samples/jboss-eap/8.1"),
					     helm.WithArgs("--set", "build.s2i.jdk17.runtimeImage=registry.redhat.io/jboss-eap-8/eap81-openjdk17-runtime-openshift-rhel9@sha256:5eaae1e1f9137b2e5970b22efc56f5908753f9f70938a6f9fffd3fc993e0464e"),
		); err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}

		client, err := c.NewClient()
		if err != nil {
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
		}
		err = wait.For(podsReady(client, namespace, "app.kubernetes.io/name="+appName), wait.WithTimeout(time.Minute*10))
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

// Add a readiness probe to the deployment to ensure that the EAP 7.x server is up and ready
// before the test can proceed
func newEAP7AppDeployment(namespace string, name string, replicas int32, containerName string, image string) *appsv1.Deployment {
	deployment := newAppDeployment(namespace, name, replicas, containerName, image)

	deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			Exec: &v1.ExecAction{
				Command: []string{"/bin/bash",
					"-c",
					"/opt/eap/bin/readinessProbe.sh"},
			},
		},
	}

	return deployment
}
