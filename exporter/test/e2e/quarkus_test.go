package e2e

import (
	"context"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestQuarkus_3_13_0(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/quarkus:3.13.0"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Quarkus from"+image).
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

			g.Expect(result.Os).Should(Ω.Equal("rhel"))
			g.Expect(result.OsVersion).Should(Ω.Equal("8.10"))
			g.Expect(result.Kind).Should(Ω.Equal("Java"))
			g.Expect(result.KindVersion).Should(Ω.Equal("17.0.12"))
			g.Expect(result.KindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			g.Expect(runtime.Version).To(Ω.Equal("3.13.0"))

			return ctx
		})
	_ = testenv.Test(t, feature.Feature())
}

func TestNativeQuarkus_3_13_0(t *testing.T) {

	appName := "native-quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/native-quarkus:3.13.0"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Native Quarkus from "+image).
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

			g.Expect(result.Os).Should(Ω.Equal("rhel"))
			g.Expect(result.OsVersion).Should(Ω.Equal("8.9"))
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
	_ = testenv.Test(t, feature.Feature())
}

func TestQuarkus_3_15_1(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/quarkus:3.15.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Quarkus from"+image).
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

			g.Expect(result.Os).Should(Ω.Equal("rhel"))
			g.Expect(result.OsVersion).Should(Ω.Equal("8.10"))
			g.Expect(result.Kind).Should(Ω.Equal("Java"))
			g.Expect(result.KindVersion).Should(Ω.Equal("21.0.5"))
			g.Expect(result.KindImplementer).Should(Ω.Equal("Red Hat, Inc."))

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			g.Expect(runtime.Version).To(Ω.Equal("3.15.1"))

			return ctx
		})
	_ = testenv.Test(t, feature.Feature())
}

func TestNativeQuarkus_3_15_1(t *testing.T) {

	appName := "native-quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/native-quarkus:3.15.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Native Quarkus from "+image).
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

			g.Expect(result.Os).Should(Ω.Equal("rhel"))
			g.Expect(result.OsVersion).Should(Ω.Equal("8.10"))
			g.Expect(result.Kind).Should(Ω.Equal("GraalVM"))
			g.Expect(result.KindVersion).Should(Ω.BeEmpty())
			g.Expect(result.KindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(result.Runtimes)).To(Ω.Equal(1))
			runtime := result.Runtimes[0]
			g.Expect(runtime.Name).To(Ω.Equal("Quarkus"))
			// In native mode, Quarkus started reporting its version since 3.15.x
			g.Expect(runtime.Version).To(Ω.Equal("3.15.1"))

			return ctx
		})
	_ = testenv.Test(t, feature.Feature())
}
