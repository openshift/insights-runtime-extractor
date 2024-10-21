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

	"exporter/pkg/types"
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

func getContainerIDAndWorkerNode(ctx context.Context, c *envconf.Config, g *Ω.WithT, namespace string, selector string, containerName string) (namespacedContainerId, string) {
	pod := getPod(ctx, c, g, namespace, selector)
	g.Expect(len(pod.Status.ContainerStatuses)).Should(Ω.Equal(1))
	container := pod.Status.ContainerStatuses[0]
	g.Expect(container.Name).Should(Ω.Equal(containerName))

	return namespacedContainerId{
		namespace:   namespace,
		podName:     pod.ObjectMeta.Name,
		containerId: container.ContainerID,
	}, pod.Spec.NodeName
}

func getPod(ctx context.Context, c *envconf.Config, g *Ω.WithT, namespace string, selector string) corev1.Pod {
	client, err := c.NewClient()
	g.Expect(err).ShouldNot(Ω.HaveOccurred())

	var pods corev1.PodList

	err = client.Resources(namespace).List(ctx, &pods, resources.WithLabelSelector(selector))
	g.Expect(err).ShouldNot(Ω.HaveOccurred())

	if len(pods.Items) > 1 {
		fmt.Printf("Got more pods than expected:%#v\n", &pods.Items)
	}
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
		fmt.Printf("Resource ready: %s/%s\n", obj.GetNamespace(), obj.GetName())
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

		if len(pods.Items) == 0 {
			return false, nil
		}

		for _, pod := range pods.Items {
			podReady := false
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					podReady = true
					break
				}
			}
			if !podReady {
				return false, nil
			}
		}

		return true, nil
	}
}

func getInsightsRuntimePodIPs(
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
		pods[pod.Spec.NodeName] = pod.Status.PodIP
	}
	return pods, nil
}

type namespacedContainerId struct {
	namespace   string
	podName     string
	containerId string
}

func extractRuntimeInfoFromContainer(ctx context.Context, g *Ω.WithT, c *envconf.Config, cid namespacedContainerId, nodeName string) types.ContainerRuntimeInfo {
	client, err := c.NewClient()
	g.Expect(err).ShouldNot(Ω.HaveOccurred())

	curlPodName := getPod(ctx, c, g, insightsRuntimeExtractorNamespace, "app.kubernetes.io/name=curl-e2e").Name

	insightsRuntimePodIPs, err := getInsightsRuntimePodIPs(ctx, c, insightsRuntimeExtractorNamespace, "app.kubernetes.io/name=insights-runtime-extractor-e2e")
	g.Expect(err).ShouldNot(Ω.HaveOccurred())

	var stdout, stderr bytes.Buffer
	command := []string{"curl", "-s", "http://" + insightsRuntimePodIPs[nodeName] + ":8000/gather_runtime_info?hash=false"}

	err = client.Resources().ExecInPod(ctx, insightsRuntimeExtractorNamespace, curlPodName, "curl", command, &stdout, &stderr)
	g.Expect(err).ShouldNot(Ω.HaveOccurred())
	g.Expect(stderr.String()).Should(Ω.BeEmpty())

	output := stdout.String()
	g.Expect(output).Should(Ω.Not(Ω.BeEmpty()))

	var infoOutput map[string]map[string]map[string]types.ContainerRuntimeInfo
	json.Unmarshal([]byte(output), &infoOutput)

	container := infoOutput[cid.namespace][cid.podName][cid.containerId]
	g.Expect(container).ShouldNot(Ω.BeNil())
	fmt.Println("Extracted runtime info:", container)

	return container
}
