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
			result, err := scanContainer(ctx, g, c, cid, nodeName)
			g.Expect(err).ShouldNot(Ω.HaveOccurred())

			g.Expect(result.OsReleaseId).Should(Ω.Equal("rhel"))
			g.Expect(result.OsReleaseVersionId).Should(Ω.Equal("8.8"))
			g.Expect(result.RuntimeKind).Should(Ω.Equal("Java"))
			g.Expect(result.RuntimeKindVersion).Should(Ω.Equal("17.0.8"))
			g.Expect(result.RuntimeKindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			g.Expect(runtime.Version).To(Ω.Equal("3.4.1"))

			return ctx
		})
	_ = testEnv.Test(t, feature.Feature())
}
