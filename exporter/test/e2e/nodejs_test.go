package e2e

import (
	"context"
	"testing"

	Ω "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNodeJS(t *testing.T) {

	appName := "node-app"
	containerName := "nodejs"
	// corresponded to node:22.6.0-alpine3.20
	image := "node@sha256:4162c8a0f1fef9d3b003eb1fd3d8a26db46815288832aa453d829f4129d4dfd3"
	deployment := newNodeAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Node.js from base image "+image).
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

			g.Expect(result.Os).Should(Ω.Equal("alpine"))
			g.Expect(result.OsVersion).Should(Ω.Equal("3.20.2"))
			g.Expect(result.Kind).Should(Ω.Equal("Node.js"))
			g.Expect(result.KindVersion).Should(Ω.Equal("v22.6.0"))
			g.Expect(result.KindImplementer).Should(Ω.BeEmpty())

			return ctx
		})
	_ = testEnv.Test(t, feature.Feature())
}

func newNodeAppDeployment(namespace string, name string, replicas int32, containerName string, image string) *appsv1.Deployment {
	deployment := newAppDeployment(namespace, name, replicas, containerName, image)

	deployment.Spec.Template.Spec.Containers[0].Command = []string{"node"}
	deployment.Spec.Template.Spec.Containers[0].Args = []string{
		"-e",
		"r=require;r(\"http\").createServer((i,o)=>r(\"stream\").pipeline(r(\"fs\").createReadStream(i.url.slice(1)),o,_=>_)).listen(8080)"}

	return deployment
}
