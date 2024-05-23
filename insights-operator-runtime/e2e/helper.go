package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"

	apimachinerywait "k8s.io/apimachinery/pkg/util/wait"

	Ω "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/e2e-framework/klient"
	"sigs.k8s.io/e2e-framework/klient/k8s"
	"sigs.k8s.io/e2e-framework/klient/k8s/resources"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/pkg/env"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

func newAppDeployment(namespace string, name string, replicas int32, containerName string, image string) *appsv1.Deployment {
	labels := map[string]string{"app": name}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: containerName, Image: image}}},
			},
		},
	}
}

func getContainerIDAndWorkerNode(ctx context.Context, c *envconf.Config, g *Ω.WithT, namespace string, selector string, containerName string) (string, string) {
	pod := getPod(ctx, c, g, namespace, selector)
	g.Expect(len(pod.Status.ContainerStatuses)).Should(Ω.Equal(1))
	container := pod.Status.ContainerStatuses[0]
	g.Expect(container.Name).Should(Ω.Equal(containerName))

	return container.ContainerID, pod.Spec.NodeName
}

func getPod(ctx context.Context, c *envconf.Config, g *Ω.WithT, namespace string, selector string) corev1.Pod {
	client, err := c.NewClient()
	g.Expect(err).ShouldNot(Ω.HaveOccurred())

	var pods corev1.PodList

	err = client.Resources(namespace).List(ctx, &pods, resources.WithLabelSelector(selector))
	g.Expect(err).ShouldNot(Ω.HaveOccurred())

	g.Expect(len(pods.Items)).Should(Ω.Equal(1))

	pod := pods.Items[0]

	g.Expect(len(pod.Status.ContainerStatuses)).Should(Ω.Equal(1))

	return pod
}

func deployAndWaitForReadiness(obj k8s.Object, selector string) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		client, err := c.NewClient()
		if err != nil {
			return nil, err
		}
		if err = client.Resources().Create(ctx, obj); err != nil {
			return nil, err
		}
		err = wait.For(podsReady(client, selector), wait.WithTimeout(time.Minute*5))
		if err != nil {
			return nil, err
		}
		return ctx, nil
	}
}

func undeploy(obj k8s.Object) env.Func {
	return func(ctx context.Context, c *envconf.Config) (context.Context, error) {
		client, err := c.NewClient()
		if err != nil {
			return nil, err
		}
		if err = client.Resources().Delete(ctx, obj); err != nil {
			return nil, err
		}

		return ctx, nil
	}
}

// PodsReady is a helper function that can be used to check that the selected pods are ready
func podsReady(client klient.Client, selector string) apimachinerywait.ConditionWithContextFunc {
	return func(ctx context.Context) (done bool, err error) {
		watchOptions := resources.WithLabelSelector(selector)

		var pods corev1.PodList

		if err := client.Resources().List(ctx, &pods, watchOptions); err != nil {
			return false, err
		}
		for _, pod := range pods.Items {
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					done = true
				}
			}
		}
		return
	}
}

func getInsightsOperatorRuntimePods(
	ctx context.Context,
	c *envconf.Config,
	csNamespace string,
	selector string) (map[string]string, error) {

	client, err := c.NewClient()
	if err != nil {
		return nil, err
	}

	pods := make(map[string]string)

	watchOptions := resources.WithLabelSelector(selector)

	var podList corev1.PodList

	if err := client.Resources(csNamespace).List(ctx, &podList, watchOptions); err != nil {
		return nil, err
	}
	for _, pod := range podList.Items {
		pods[pod.Spec.NodeName] = pod.ObjectMeta.Name
	}
	return pods, nil
}

// workloadNamespacePods tracks the identified pod shapes within a namespace.
type runtimeWorkloadInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type workloadInfo struct {
	OsReleaseId            string                `json:"os-release-id,omitempty"`
	OsReleaseVersionId     string                `json:"os-release-version-id,omitempty"`
	RuntimeKind            string                `json:"runtime-kind,omitempty"`
	RuntimeKindVersion     string                `json:"runtime-kind-version,omitempty"`
	RuntimeKindImplementer string                `json:"runtime-kind-implementer,omitempty"`
	Runtimes               []runtimeWorkloadInfo `json:"runtimes,omitempty"`
}

func scanContainer(ctx context.Context, g *Ω.WithT, c *envconf.Config, cid string, nodeName string) (workloadInfo, error) {
	client, err := c.NewClient()

	g.Expect(err).ShouldNot(Ω.HaveOccurred())

	insightsOperatorRuntimePods, err := getInsightsOperatorRuntimePods(ctx, c, insightsOperatorRuntimeNamespace, "app.kubernetes.io/name=insights-operator-runtime-e2e")
	g.Expect(err).ShouldNot(Ω.HaveOccurred())

	pod := insightsOperatorRuntimePods[nodeName]
	execCommand := []string{"/scan-container", "--log-level", "trace", cid}

	var stdout, stderr bytes.Buffer

	if err := client.Resources().ExecInPod(context.TODO(), insightsOperatorRuntimeNamespace, pod, "insights-operator-runtime", execCommand, &stdout, &stderr); err != nil {
		g.Expect(err).ShouldNot(Ω.HaveOccurred())
	}

	result := stdout.String()

	fmt.Printf("Scanned container %s:\n%s\n", cid, result)

	var scanOutput map[string]map[string]map[string]workloadInfo
	json.Unmarshal([]byte(result), &scanOutput)

	g.Expect(scanOutput).To(Ω.HaveKey(namespace))

	scanNamespace, _ := scanOutput[namespace]
	g.Expect(len(scanNamespace)).Should(Ω.Equal(1))
	for _, scanPod := range scanNamespace {
		g.Expect(scanPod).To(Ω.HaveKey(cid))
		containerScan := scanPod[cid]
		return containerScan, nil
	}

	return workloadInfo{}, fmt.Errorf("unable to scan container %s", cid)
}
