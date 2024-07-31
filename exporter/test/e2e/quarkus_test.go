package e2e

import (
	"context"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestQuarkus(t *testing.T) {

	appName := "upstream-quarkus"
	containerName := "c1"
	image := "quay.io/jmesnil/code-with-quarkus"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Upstream Quarkus").
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
			result := scanContainer(ctx, g, c, cid, nodeName)
			g.Expect(result).ShouldNot(Ω.BeNil())

			g.Expect(result.Os).Should(Ω.Equal("rhel"))
			g.Expect(result.OsVersion).Should(Ω.Equal("8.8"))
			g.Expect(result.Kind).Should(Ω.Equal("Java"))
			g.Expect(result.KindVersion).Should(Ω.Equal("17.0.8"))
			g.Expect(result.KindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			g.Expect(runtime.Version).To(Ω.Equal("3.4.1"))

			return ctx
		})
	_ = testEnv.Test(t, feature.Feature())
}

func TestQuarkusNative(t *testing.T) {

	appName := "quarkus-native"
	containerName := "c1"
	image := "quay.io/jmesnil/quarkus-native-app"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Upstream Native Quarkus").
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
			result := scanContainer(ctx, g, c, cid, nodeName)
			g.Expect(result).ShouldNot(Ω.BeNil())

			g.Expect(result.Os).Should(Ω.Equal("debian"))
			g.Expect(result.OsVersion).Should(Ω.Equal("11"))
			g.Expect(result.Kind).Should(Ω.Equal("GraalVM"))
			g.Expect(result.KindVersion).Should(Ω.BeEmpty())
			g.Expect(result.KindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			// In native mode, Quarkus does not expose its version
			g.Expect(runtime.Version).To(Ω.BeEmpty())

			return ctx
		})
	_ = testEnv.Test(t, feature.Feature())
}
