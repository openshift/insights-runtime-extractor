package e2e

import (
	"os"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var (
	testEnv env.Environment
	// insightsOperatorRuntimeNamespace is the namespace where the container scanner is deployed
	insightsOperatorRuntimeNamespace string
	// namespace is the namespace where workloads are deployed before they are scanned
	namespace string
	// testedExtractorImage is the URL of the image for the insights-runtime-extractor's extractor
	testedExtractorImage string
	// testedExporterImage is the URL of the image for the insights-runtime-extractor's exporter
	testedExporterImage string
)

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	namespace = envconf.RandomName("e2e", 10)
	insightsOperatorRuntimeNamespace = os.Getenv("TEST_NAMESPACE")
	testedExtractorImage = "quay.io/jmesnil/insights-runtime-extractor:latest"
	if value, ok := os.LookupEnv("TESTED_EXTRACTOR_IMAGE"); ok {
		testedExtractorImage = value
	}
	testedExporterImage = "quay.io/jmesnil/insights-runtime-exporter:latest"
	if value, ok := os.LookupEnv("TESTED_EXPORTER_IMAGE"); ok {
		testedExporterImage = value
	}

	insightsOperatorRuntime := newContainerScannerDaemonSet()
	curl := newCurlDeployment()

	testEnv.Setup(
		envfuncs.CreateNamespace(namespace),
		deployAndWaitForReadiness(curl, "app.kubernetes.io/name=curl-e2e"),
		deployAndWaitForReadiness(insightsOperatorRuntime, "app.kubernetes.io/name=insights-runtime-extractor-e2e"),
	)

	testEnv.Finish(
		undeploy(insightsOperatorRuntime),
		undeploy(curl),
		envfuncs.DeleteNamespace(namespace),
	)

	os.Exit(testEnv.Run(m))
}

func newContainerScannerDaemonSet() *appsv1.DaemonSet {
	securityContextPrivileged := true
	hostPathSocket := corev1.HostPathSocket
	labels := map[string]string{"app.kubernetes.io/name": "insights-runtime-extractor-e2e"}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "insights-runtime-extractor-e2e", Namespace: insightsOperatorRuntimeNamespace},
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
		ObjectMeta: metav1.ObjectMeta{Name: "curl-e2e", Namespace: insightsOperatorRuntimeNamespace},
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
