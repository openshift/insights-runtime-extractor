package e2e

import (
	"context"
	"testing"

	Ω "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestPython3(t *testing.T) {

	appName := "python3-app"
	containerName := "python"
	// corresponded to python:3.9.19-slim
	image := "python@sha256:85c7a2a383a01e0b77b5f9c97d8b1eef70409a99552fde03c518a98dfa19609c"
	deployment := newPython3AppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Python3 from base image "+image).
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
			g.Expect(result.OsVersion).Should(Ω.Equal("12"))
			g.Expect(result.Kind).Should(Ω.Equal("Python"))
			g.Expect(result.KindVersion).Should(Ω.Equal("Python 3.9.19"))
			g.Expect(result.KindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(result.Runtimes)).To(Ω.Equal(0))

			return ctx
		})
	_ = testEnv.Test(t, feature.Feature())
}

func newPython3AppDeployment(namespace string, name string, replicas int32, containerName string, image string) *appsv1.Deployment {
	deployment := newAppDeployment(namespace, name, replicas, containerName, image)

	deployment.Spec.Template.Spec.Containers[0].Command = []string{"python3"}
	deployment.Spec.Template.Spec.Containers[0].Args = []string{
		"-m",
		"http.server"}

	return deployment
}
