package e2e

import (
	"exporter/pkg/types"
	"testing"

	Ω "github.com/onsi/gomega"
	"sigs.k8s.io/e2e-framework/pkg/features"
)

func TestQuarkus_3_13_0(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/quarkus:3.13.0"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "17.0.12",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Quarkus",
					Version: "3.13.0",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestNativeQuarkus_3_13_0(t *testing.T) {

	appName := "native-quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/native-quarkus:3.13.0"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Native Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:        "rhel",
				OsVersion: "8.9",
				Kind:      "GraalVM",
				Runtimes: []types.RuntimeComponent{{
					Name: "Quarkus",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestQuarkus_3_15_1(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/quarkus:3.15.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "21.0.4",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Quarkus",
					Version: "3.15.1",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestNativeQuarkus_3_15_1(t *testing.T) {

	appName := "native-quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/native-quarkus:3.15.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Native Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:        "rhel",
				OsVersion: "8.10",
				Kind:      "GraalVM",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Quarkus",
					Version: "3.15.1",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestQuarkus_3_19_3(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/quarkus:3.19.3"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "21.0.4",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Quarkus",
					Version: "3.19.3",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestNativeQuarkus_3_19_3(t *testing.T) {

	appName := "native-quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/native-quarkus:3.19.3"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Native Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:        "rhel",
				OsVersion: "8.10",
				Kind:      "GraalVM",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Quarkus",
					Version: "3.19.3",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestQuarkus_3_24_4(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/quarkus:3.24.4"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Quarkus from "+image).
	Setup(deployTestResource(deployment, appName)).
	Teardown(undeployTestResource(deployment, appName)).
	Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
		expected := types.ContainerRuntimeInfo{
			Os:              "rhel",
			OsVersion:       "8.10",
			Kind:            "Java",
			KindVersion:     "21.0.4",
			KindImplementer: "Red Hat, Inc.",
			Runtimes: []types.RuntimeComponent{{
				Name:    "Quarkus",
				Version: "3.24.4",
			}},
		}
		g.Expect(runtimeInfo).Should(Ω.Equal(expected))
	}))
	_ = testenv.Test(t, feature.Feature())
}

func TestNativeQuarkus_3_24_4(t *testing.T) {

	appName := "native-quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/native-quarkus:3.24.4"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Native Quarkus from "+image).
	Setup(deployTestResource(deployment, appName)).
	Teardown(undeployTestResource(deployment, appName)).
	Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
		expected := types.ContainerRuntimeInfo{
			Os:        "rhel",
			OsVersion: "8.10",
			Kind:      "GraalVM",
			Runtimes: []types.RuntimeComponent{{
				Name:    "Quarkus",
				Version: "3.24.4",
			}},
		}
		g.Expect(runtimeInfo).Should(Ω.Equal(expected))
	}))
	_ = testenv.Test(t, feature.Feature())
}

func TestRedHatBuildOfQuarkus_3_8_6(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/rhbq-app:3.8.6"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Red Hat Build of Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "17.0.13",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Quarkus",
					Version: "3.8.6.redhat-00004",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestRedHatBuildOfQuarkus_3_15_3(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/rhbq-app:3.15.3"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Red Hat Build of Quarkus from "+image).
		Setup(deployTestResource(deployment, appName)).
		Teardown(undeployTestResource(deployment, appName)).
		Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
			expected := types.ContainerRuntimeInfo{
				Os:              "rhel",
				OsVersion:       "8.10",
				Kind:            "Java",
				KindVersion:     "17.0.13",
				KindImplementer: "Red Hat, Inc.",
				Runtimes: []types.RuntimeComponent{{
					Name:    "Quarkus",
					Version: "3.15.3.redhat-00002",
				}},
			}
			g.Expect(runtimeInfo).Should(Ω.Equal(expected))
		}))
	_ = testenv.Test(t, feature.Feature())
}

func TestRedHatBuildOfQuarkus_3_20_1(t *testing.T) {

	appName := "quarkus"
	containerName := "main"
	image := "quay.io/insights-runtime-extractor-samples/rhbq-app:3.20.1"
	deployment := newAppDeployment(namespace, appName, 1, containerName, image)

	feature := features.New("Red Hat Build of Quarkus from "+image).
	Setup(deployTestResource(deployment, appName)).
	Teardown(undeployTestResource(deployment, appName)).
	Assess("runtime info extracted", checkExtractedRuntimeInfo(namespace, "app="+appName, containerName, func(g *Ω.WithT, runtimeInfo types.ContainerRuntimeInfo) {
		expected := types.ContainerRuntimeInfo{
			Os:              "rhel",
			OsVersion:       "8.10",
			Kind:            "Java",
			KindVersion:     "21.0.8",
			KindImplementer: "Red Hat, Inc.",
			Runtimes: []types.RuntimeComponent{{
				Name:    "Quarkus",
				Version: "3.20.1.redhat-00003",
			}},
		}
		g.Expect(runtimeInfo).Should(Ω.Equal(expected))
	}))
	_ = testenv.Test(t, feature.Feature())
}