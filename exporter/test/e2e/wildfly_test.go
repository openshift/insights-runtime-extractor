package e2e

import (
	"context"
	"testing"

	Ω "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func deployTestResource(deployment *appsv1.Deployment, appName string) func(context.Context, *testing.T, *envconf.Config) context.Context {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		g := Ω.NewWithT(t)

		ctx, err := deployAndWaitForReadiness(deployment, "app="+appName)(ctx, c)
		g.Expect(err).ShouldNot(Ω.HaveOccurred())
		return ctx
	}
}

func undeployTestResource(deployment *appsv1.Deployment, appName string) func(context.Context, *testing.T, *envconf.Config) context.Context {
	return func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
		g := Ω.NewWithT(t)

		ctx, err := undeploy(deployment)(ctx, c)
		g.Expect(err).ShouldNot(Ω.HaveOccurred())
		return ctx
	}
}

func TestWildFly_30_0_1_Final(t *testing.T) {

	appName := "wildfly-app"
	containerName := "main"
	// corresponded to quay.io/wildfly/wildfly:30.0.1.Final-jdk20
	image := "quay.io/wildfly/wildfly@sha256:4b51924efea743d48685eb2065c33d21f818fb9c5acc7139dc47e5b810610e3b"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("WildFly 30.0.1.Final application "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", func(ctx context.Context, t *testing.T, c *envconf.Config) context.Context {
			g := Ω.NewWithT(t)

			cid, nodeName := getContainerIDAndWorkerNode(ctx, c, g, namespace, "app="+appName, containerName)
			result := extractRuntimeInfoFromContainer(ctx, g, c, cid, nodeName)
			g.Expect(result).ShouldNot(Ω.BeNil())

			g.Expect(result.Os).Should(Ω.Equal("rhel"))
			g.Expect(result.OsVersion).Should(Ω.Equal("9.2"))
			g.Expect(result.Kind).Should(Ω.Equal("Java"))
			g.Expect(result.KindVersion).Should(Ω.Equal("20.0.2"))
			g.Expect(result.KindImplementer).Should(Ω.Equal("Eclipse Adoptium"))

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("WildFly Full"))
			g.Expect(runtime.Version).To(Ω.Equal("30.0.1.Final"))

			return ctx
		})
	_ = testenv.Test(t, feature.Feature())
}
