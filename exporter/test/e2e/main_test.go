package e2e

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apimachinerywait "k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/e2e-framework/klient/conf"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/envfuncs"
)

var (
	testenv env.Environment
	// insightsOperatorRuntimeNamespace is the namespace where the insights runtime extractor is deployed
	insightsRuntimeExtractorNamespace string
	// namespace is the namespace where workloads are deployed before their runtime info are extracted
	namespace                    string
	insightsOperatorRuntimeLabel = "app.kubernetes.io/name=insights-runtime-extractor-e2e"
)

func TestMain(m *testing.M) {
	path := conf.ResolveKubeConfigFile()
	cfg := envconf.NewWithKubeConfig(path)
	testenv = env.NewWithConfig(cfg)
	namespace = "e2e-insights-runtime-extractor"
	insightsRuntimeExtractorNamespace = os.Getenv("TEST_NAMESPACE")
	testedExtractorImage := "quay.io/openshift/origin-insights-runtime-extractor:latest"
	testedExporterImage := "quay.io/openshift/origin-insights-runtime-exporter:latest"
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
		createHeadlessService(),
		deployAndWaitForReadiness(curl, "app.kubernetes.io/name=curl-e2e"),
		deployAndWaitForReadiness(insightsOperatorRuntime, insightsOperatorRuntimeLabel),
	)

	testenv.Finish(
		fetchExtractorLogs(),
		undeploy(insightsOperatorRuntime),
		undeploy(curl),
		deleteHeadlessService(),
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
						}, {
							MountPath: "/etc/tls/private",
							Name:      "insights-runtime-extractor-tls",
							ReadOnly:  true,
						}},
						Command: []string{"/extractor_server", "--log-level", "trace", "--tls-cert", "/etc/tls/private/tls.crt", "--tls-key", "/etc/tls/private/tls.key"},
					}, {
						Name:            "exporter",
						Image:           testedExporterImage,
						ImagePullPolicy: corev1.PullAlways,
						VolumeMounts: []corev1.VolumeMount{{
							MountPath: "/data",
							Name:      "data-volume",
						}, {
							MountPath: "/etc/tls/private",
							Name:      "insights-runtime-extractor-tls",
							ReadOnly:  true,
						}},
						Command: []string{"/exporter", "-bind", "0.0.0.0", "-tls-cert", "/etc/tls/private/tls.crt", "-tls-key", "/tls/tls.key", "-tls-ca", "/etc/tls/private/tls.crt", "-tls-server-name", "localhost"},
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
					}, {
						Name: "insights-runtime-extractor-tls",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "insights-runtime-extractor-tls",
							},
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

// createHeadlessService creates a headless service with the serving-cert annotation,
// waits for the OpenShift service cert controller to generate the TLS secret
func createHeadlessService() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		client, err := c.NewClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %v", err)
		}

		// Create the headless service with the serving-cert annotation
		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "exporter",
				Namespace: insightsRuntimeExtractorNamespace,
				Annotations: map[string]string{
					"service.beta.openshift.io/serving-cert-secret-name": "insights-runtime-extractor-tls",
				},
				Labels: map[string]string{
					"app.kubernetes.io/name": "insights-runtime-extractor-e2e",
				},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{
					"app.kubernetes.io/name": "insights-runtime-extractor-e2e",
				},
				Ports: []corev1.ServicePort{{
					Name:     "https",
					Protocol: corev1.ProtocolTCP,
					Port:     8000,
				}},
				Type:      corev1.ServiceTypeClusterIP,
				ClusterIP: corev1.ClusterIPNone,
			},
		}

		if err = client.Resources(insightsRuntimeExtractorNamespace).Create(ctx, service); err != nil {
			return nil, fmt.Errorf("failed to create service: %v", err)
		}
		fmt.Println("Service created: exporter")

		// Wait for the OpenShift service cert controller to create the secret
		clientset, err := kubernetes.NewForConfig(client.RESTConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to create Kubernetes clientset: %v", err)
		}

		err = apimachinerywait.PollUntilContextTimeout(ctx, 5*time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
			_, err := clientset.CoreV1().Secrets(insightsRuntimeExtractorNamespace).Get(ctx, "insights-runtime-extractor-tls", metav1.GetOptions{})
			if err != nil {
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			return nil, fmt.Errorf("timed out waiting for secret insights-runtime-extractor-tls: %v", err)
		}
		fmt.Println("Secret insights-runtime-extractor-tls found")

		return ctx, nil
	}
}

// deleteHeadlessService cleans up the headless service.
func deleteHeadlessService() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		client, err := c.NewClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create client: %v", err)
		}

		service := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "exporter",
				Namespace: insightsRuntimeExtractorNamespace,
			},
		}
		if err = client.Resources(insightsRuntimeExtractorNamespace).Delete(ctx, service); err != nil {
			fmt.Printf("failed to delete service: %v\n", err)
		}

		return ctx, nil
	}
}

// After the tests are run, store the logs from the extractor in $ARTIFACT_DIR
func fetchExtractorLogs() env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		client, err := c.NewClient()
		if err != nil {
			return ctx, fmt.Errorf("failed to create client: %v", err)
		}
		clientset, err := kubernetes.NewForConfig(client.RESTConfig())
		if err != nil {
			return ctx, fmt.Errorf("failed to create Kubernetes client: %v", err)
		}

		artifactDir := os.Getenv("ARTIFACT_DIR")
		if artifactDir == "" {
			artifactDir = "."
		}

		// find all the pods of the daemon set
		var pods corev1.PodList
		err = client.Resources(insightsRuntimeExtractorNamespace).List(ctx, &pods, resources.WithLabelSelector(insightsOperatorRuntimeLabel))
		if err != nil {
			return nil, err
		}

		// for each pods, fetch the logs of its extractor container
		// and store them in a $ARTIFACT_DIR/$pod.log file
		for _, pod := range pods.Items {
			namespace := pod.ObjectMeta.Namespace
			podName := pod.ObjectMeta.Name
			containerName := "extractor"
			fmt.Printf("Fetching logs from %s\n", podName)

			podLogOptions := &corev1.PodLogOptions{
				Container: containerName,
				Follow:    false,
			}

			logFilePath := filepath.Join(artifactDir, fmt.Sprintf("%s.log", podName))
			logFile, err := os.Create(logFilePath)
			if err != nil {
				panic(fmt.Errorf("failed to create log file: %v", err))
			}
			defer logFile.Close()

			req := clientset.CoreV1().Pods(namespace).GetLogs(podName, podLogOptions)
			logStream, err := req.Stream(ctx)
			if err != nil {
				return nil, fmt.Errorf("error opening log stream: %v", err)
			}
			defer logStream.Close()

			_, err = io.Copy(logFile, logStream)
			if err != nil {
				fmt.Printf("failed to write logs to file: %v", err)
			}
		}

		return ctx, nil
	}
}
