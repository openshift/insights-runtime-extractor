package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestNodeJS_22_6_0(t *testing.T) {
	appName := "node-app-22-6-0"
	containerName := "nodejs-22-6-0"
	// corresponded to node:22.6.0-alpine3.20
	image := "quay.io/insights-runtime-extractor-samples/node@sha256:7cbffc9cf3886cea69479e3ea3a57034896264bf0e263944465437cdcf13b345"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

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

func TestNodeJS_22_14_0(t *testing.T) {
	appName := "node-app-22-14-0"
	containerName := "nodejs-22-14-0"
	// corresponded to node:22.14.0-slim
	image := "quay.io/insights-runtime-extractor-samples/node@sha256:9a81af4036bda12b35b8be6a55586d0a088dfff0cbaa51f326301623a1b17814"	
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Node.js from base image "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:          "debian",
				OsVersion:   "12",
				Kind:        "Node.js",
				KindVersion: "v22.14.0",
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestNodeJS_24_4_1(t *testing.T) {
	appName := "node-app-24-4-1"
	containerName := "nodejs-24-4-1"
	// corresponded to node:24.4.1-alpine3.22
	image := "quay.io/insights-runtime-extractor-samples/node:24.4.1-alpine3.22"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Node.js from base image "+image).
	Setup(deployTestResource(deployment, appName)).
	Teardown(undeployTestResource(deployment, appName)).
	Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
		expected := types.ContainerRuntimeInfo{
			Os:          "alpine",
			OsVersion:   "3.22.1",
			Kind:        "Node.js",
			KindVersion: "v24.4.1",
		}
		g.Expect(runtimeInfo).Should(Ω.Equal(expected))
	}))
	_ = testenv.Test(t, feature.Feature())
}

