package e2e

import (
	"context"
	"fmt"
	"testing"

	appsv1 "k8s.io/api/apps/v1"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestUbi9Minimal(t *testing.T) {
	testBaseImage(t, "registry.access.redhat.com/ubi9-minimal:9.4", "rhel", "9.4")
}

func TestUbi8Minimal(t *testing.T) {
	testBaseImage(t, "registry.access.redhat.com/ubi8/ubi-minimal:8.10", "rhel", "8.10")
}

func TestDebian(t *testing.T) {
	testBaseImage(t, "debian:12", "debian", "12")
}

func TestCentOs7(t *testing.T) {
	testBaseImage(t, "centos:7", "centos", "7")
}

func testBaseImage(t *testing.T, baseImage string, expectedOsReleaseId string, expectedOsReleaseVersionId string) {

	appName := envconf.RandomName("os-", 10)
	fmt.Printf("Running base image with name %s", appName)
	containerName := "main"
	deployment := newBaseImageDeployment(namespace, appName, 1, containerName, baseImage)

	feature := features.New("base image "+baseImage).
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

			g.Expect(result.OsReleaseId).Should(Ω.Equal(expectedOsReleaseId))
			g.Expect(result.OsReleaseVersionId).Should(Ω.Equal(expectedOsReleaseVersionId))
			g.Expect(result.RuntimeKind).Should(Ω.BeEmpty())
			g.Expect(result.RuntimeKindVersion).Should(Ω.BeEmpty())
			g.Expect(result.RuntimeKindImplementer).Should(Ω.BeEmpty())

			g.Expect(len(result.Runtimes)).To(Ω.Equal(0))

			return ctx
		})
	_ = testEnv.Test(t, feature.Feature())
}

func newBaseImageDeployment(namespace string, name string, replicas int32, containerName string, image string) *appsv1.Deployment {
	deployment := newAppDeployment(namespace, name, replicas, containerName, image)

	deployment.Spec.Template.Spec.Containers[0].Command = []string{"tail", "-f", "/dev/null"}

	return deployment
}
