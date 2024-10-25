package e2e

import (
	"fmt"
	"os"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var (
	testenv env.Environment
	// insightsOperatorRuntimeNamespace is the namespace where the insights runtime extractor is deployed
	insightsRuntimeExtractorNamespace string
	// namespace is the namespace where workloads are deployed before their runtime info are extracted
	namespace string
)

func TestMain(m *testing.M) {
	path := conf.ResolveKubeConfigFile()
	cfg := envconf.NewWithKubeConfig(path)
	testenv = env.NewWithConfig(cfg)
	namespace = "e2e-insights-runtime-extractor"
	insightsRuntimeExtractorNamespace = os.Getenv("TEST_NAMESPACE")
	testedExtractorImage := "insights-runtime-extractor:latest"
	testedExporterImage := "insights-runtime-exporter:latest"
	if imageRegistry, ok := os.LookupEnv("IMAGE_REGISTRY"); ok {
		testedExtractorImage = imageRegistry + "/insights-runtime-extractor:latest"
		testedExporterImage = imageRegistry + "/insights-runtime-exporter:latest"
	}
	if testingExtractorImage, ok := os.LookupEnv("INSIGHTS_RUNTIME_EXTRACTOR"); ok {
		testedExtractorImage = testingExtractorImage
	}
	if testingExporterImage, ok := os.LookupEnv("INSIGHTS_RUNTIME_EXPORTER"); ok {
		testedExporterImage = testingExporterImage
	}

	fmt.Printf("#### Tested images:\n- %s\n- %s\n", testedExtractorImage, testedExporterImage)
	insightsOperatorRuntime := newInsightsRuntimeExtractorDaemonSet(testedExtractorImage, testedExporterImage)
	curl := newCurlDeployment()

	testenv.Setup(
		envfuncs.CreateNamespace(namespace),
		deployAndWaitForReadiness(curl, "app.kubernetes.io/name=curl-e2e"),
		deployAndWaitForReadiness(insightsOperatorRuntime, "app.kubernetes.io/name=insights-runtime-extractor-e2e"),
	)

	testenv.Finish(
		undeploy(insightsOperatorRuntime),
		undeploy(curl),
		envfuncs.DeleteNamespace(namespace),
	)

	os.Exit(testenv.Run(m))
}

func newInsightsRuntimeExtractorDaemonSet(testedExtractorImage string, testedExporterImage string) *appsv1.DaemonSet {
	securityContextPrivileged := true
	hostPathSocket := corev1.HostPathSocket
	labels := map[string]string{"app.kubernetes.io/name": "insights-runtime-extractor-e2e"}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "insights-runtime-extractor-e2e", Namespace: insightsRuntimeExtractorNamespace},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					ServiceAccountName: "insights-runtime-extractor-sa",
					HostPID:            true,
					Containers: []corev1.Container{{
						Name:            "extractor",
						Image:           testedExtractorImage,
						ImagePullPolicy: corev1.PullAlways,
						Env: []corev1.EnvVar{{
							Name:  "CONTAINER_RUNTIME_ENDPOINT",
							Value: "unix:///crio.sock",
						}},
						SecurityContext: &corev1.SecurityContext{
							Privileged: &securityContextPrivileged,
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{"ALL"},
								Add:  []corev1.Capability{"CAP_SYS_ADMIN"},
							}},
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/crio.sock",
							Name:      "crio-socket",
						}, {
							MountPath: "/data",
							Name:      "data-volume",
						}},
					}, {
						Name:            "exporter",
						Image:           testedExporterImage,
						ImagePullPolicy: corev1.PullAlways,
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/data",
							Name:      "data-volume",
						}},
						Command: []string{"/exporter", "-bind", "0.0.0.0"},
					}},
					Volumes: []corev1.Volume{{
						Name: "crio-socket",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/run/crio/crio.sock",
								Type: &hostPathSocket,
							}},
					}, {
						Name: "data-volume",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
				},
			},
		},
	}
}

func newCurlDeployment() *appsv1.Deployment {
	labels := map[string]string{"app.kubernetes.io/name": "curl-e2e"}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "curl-e2e", Namespace: insightsRuntimeExtractorNamespace},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "curl",
						Image:   "quay.io/curl/curl",
						Command: []string{"tail", "-f", "/dev/null"},
					}},
				},
			},
		},
	}
}
