package e2e

import (
	"context"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestTomcat(t *testing.T) {

	appName := "upstream-tomcat"
	containerName := "c1"
	image := "tomcat:11.0-jre21"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Tomcat from base image "+image).
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			g := Ω.NewWithT(t)

			ctx, err := deployAndWaitForReadiness(deployment, "app="+appName)(ctx, c)
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			g := Ω.NewWithT(t)

			ctx, err := undeploy(deployment)(ctx, c)
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
			return ctx
		}).
		Assess("is scanned", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			g := Ω.NewWithT(t)

			cid, nodeName := getContainerIDAndWorkerNode(ctx, c, g, namespace, "app="+appName, containerName)
			result, err := scanContainer(ctx, g, c, cid, nodeName)
			g.Expect(err).ShouldNot(Ω.HaveOccurred())

			g.Expect(result.OsReleaseId).Should(Ω.Equal("ubuntu"))
			g.Expect(result.OsReleaseVersionId).Should(Ω.Equal("22.04"))
			g.Expect(result.RuntimeKind).Should(Ω.Equal("Java"))
			g.Expect(result.RuntimeKindVersion).Should(Ω.Equal("21.0.3"))
			g.Expect(result.RuntimeKindImplementer).Should(Ω.Equal("Eclipse Adoptium"))

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Apache Tomcat"))
			g.Expect(runtime.Version).To(Ω.Equal("11.0.0-M21"))

			return ctx
		})
	_ = testEnv.Test(t, feature.Feature())
}

func TestJBossWebServer(t *testing.T) {

	appName := "jboss-webserver"
	containerName := "c1"
	image := "registry.redhat.io/jboss-webserver-5/jws58-openjdk17-openshift-rhel8"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("JBoss WebServer from base image "+image).
		Setup(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			g := Ω.NewWithT(t)

			ctx, err := deployAndWaitForReadiness(deployment, "app="+appName)(ctx, c)
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			g := Ω.NewWithT(t)

			ctx, err := undeploy(deployment)(ctx, c)
			g.Expect(err).ShouldNot(Ω.HaveOccurred())
			return ctx
		}).
		Assess("is scanned", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			g := Ω.NewWithT(t)

			cid, nodeName := getContainerIDAndWorkerNode(ctx, c, g, namespace, "app="+appName, containerName)
			result, err := scanContainer(ctx, g, c, cid, nodeName)
			g.Expect(err).ShouldNot(Ω.HaveOccurred())

			g.Expect(result.OsReleaseId).Should(Ω.Equal("rhel"))
			g.Expect(result.OsReleaseVersionId).Should(Ω.Equal("8.10"))
			g.Expect(result.RuntimeKind).Should(Ω.Equal("Java"))
			g.Expect(result.RuntimeKindVersion).Should(Ω.Equal("17.0.11"))
			g.Expect(result.RuntimeKindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Apache Tomcat"))
			g.Expect(runtime.Version).To(Ω.Equal("9.0.87.redhat-00003"))

			return ctx
		})
	_ = testEnv.Test(t, feature.Feature())
}
