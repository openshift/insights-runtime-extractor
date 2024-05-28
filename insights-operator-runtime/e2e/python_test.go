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
	image := "python:3.12.2"
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
			result, err := scanContainer(ctx, g, c, cid, nodeName)
			g.Expect(err).ShouldNot(Ω.HaveOccurred())

			g.Expect(result.OsReleaseId).Should(Ω.Equal("debian"))
			g.Expect(result.OsReleaseVersionId).Should(Ω.Equal("12"))
			g.Expect(result.RuntimeKind).Should(Ω.Equal("Python"))
			g.Expect(result.RuntimeKindVersion).Should(Ω.Equal("Python 3.12.2"))
			g.Expect(result.RuntimeKindImplementer).Should(Ω.BeEmpty())

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
