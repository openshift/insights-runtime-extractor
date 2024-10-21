package e2e

import (
	"context"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestSpringBoot(t *testing.T) {

	appName := "spring-boot"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/spring-boot:3.1.4"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Spring Boot from "+image).
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

			g.Expect(result.Os).Should(Ω.Equal("ubuntu"))
			g.Expect(result.OsVersion).Should(Ω.Equal("20.04"))
			g.Expect(result.Kind).Should(Ω.Equal("Java"))
			g.Expect(result.KindVersion).Should(Ω.Equal("17.0.12"))
			g.Expect(result.KindImplementer).Should(Ω.Equal("Eclipse Adoptium"))

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Spring Boot"))
			g.Expect(runtime.Version).To(Ω.Equal("3.1.4"))

			return ctx
		})
	_ = testenv.Test(t, feature.Feature())
}
