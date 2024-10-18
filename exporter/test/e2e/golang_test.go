package e2e

import (
	"context"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestGolang(t *testing.T) {

	appName := "golang-app"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/golang-app:1.22.6"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Golang application "+image).
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
		Assess("runtime info extracted", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			g := Ω.NewWithT(t)

			cid, nodeName := getContainerIDAndWorkerNode(ctx, c, g, namespace, "app="+appName, containerName)
			result := extractRuntimeInfoFromContainer(ctx, g, c, cid, nodeName)
			g.Expect(result).ShouldNot(Ω.BeNil())

			g.Expect(result.Os).Should(Ω.Equal("debian"))
			g.Expect(result.OsVersion).Should(Ω.Equal("12"))
			g.Expect(result.Kind).Should(Ω.Equal("Golang"))
			g.Expect(result.KindVersion).Should(Ω.Equal("go1.22.6"))
			g.Expect(result.KindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(result.Runtimes)).To(Ω.Equal(0))

			return ctx
		})
	_ = testenv.Test(t, feature.Feature())
}
