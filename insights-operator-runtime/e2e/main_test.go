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
)

func TestMain(m *testing.M) {
	cfg, _ := envconf.NewFromFlags()
	testEnv = env.NewWithConfig(cfg)
	namespace = envconf.RandomName("e2e", 10)
	insightsOperatorRuntimeNamespace = os.Getenv("TEST_NAMESPACE")

	insightsOperatorRuntime := newContainerScannerDaemonSet()

	testEnv.Setup(
		envfuncs.CreateNamespace(namespace),
		deployAndWaitForReadiness(insightsOperatorRuntime, "app.kubernetes.io/name=insights-operator-runtime-e2e"),
	)

	testEnv.Finish(
		undeploy(insightsOperatorRuntime),
		envfuncs.DeleteNamespace(namespace),
	)

	os.Exit(testEnv.Run(m))
}

func newContainerScannerDaemonSet() *appsv1.DaemonSet {
	securityContextPrivileged := true
	hostPathSocket := corev1.HostPathSocket
	labels := map[string]string{"app.kubernetes.io/name": "insights-operator-runtime-e2e"}

	return &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{Name: "insights-operator-runtime-e2e", Namespace: insightsOperatorRuntimeNamespace},
		Spec: appsv1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					ServiceAccountName: "insights-operator-runtime-sa",
					HostPID:            true,
					ImagePullSecrets: []corev1.LocalObjectReference{{
						Name: "insights-operator-runtime-pull-secret",
					}},
					Containers: []corev1.Container{{
						Name:            "container-scanner",
						Image:           "ghcr.io/jmesnil/insights-operator-runtime:latest",
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
						}},
					}},
					Volumes: []corev1.Volume{{
						Name: "crio-socket",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "/run/crio/crio.sock",
								Type: &hostPathSocket,
							},
						}}},
				},
			},
		},
	}
}
