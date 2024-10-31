package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNodeJS(t *testing.T) {
	appName := "node-app"
	containerName := "nodejs"
	// corresponded to node:22.6.0-alpine3.20
	image := "node@sha256:4162c8a0f1fef9d3b003eb1fd3d8a26db46815288832aa453d829f4129d4dfd3"
	deployment := newNodeAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Node.js from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:          "alpine",
				OsVersion:   "3.20.2",
				Kind:        "Node.js",
				KindVersion: "v22.6.0",
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func newNodeAppDeployment(namespace string, name string, replicas int32, containerName string, image string) *appsv1.Deployment {
	deployment := newAppDeployment(namespace, name, replicas, containerName, image)

	deployment.Spec.Template.Spec.Containers[0].Command = []string{"node"}
	deployment.Spec.Template.Spec.Containers[0].Args = []string{
		"-e",
		"r=require;r(\"http\").createServer((i,o)=>r(\"stream\").pipeline(r(\"fs\").createReadStream(i.url.slice(1)),o,_=>_)).listen(8080)"}

	return deployment
}
