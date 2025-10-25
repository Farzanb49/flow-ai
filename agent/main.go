package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	ns := getenv("NAMESPACE", "default")
	client, _, err := getClient(ns)
	if err != nil { log.Fatal(err) }
	ctx := context.Background()

	pods, err := client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: "serving.knative.dev/service"})
	if err != nil { log.Fatal(err) }
	for _, p := range pods.Items {
		for _, cs := range p.Status.ContainerStatuses {
			if cs.State.Waiting != nil || cs.RestartCount > 0 {
				logs, _ := client.CoreV1().Pods(ns).GetLogs(p.Name, &corev1.PodLogOptions{}).DoRaw(ctx)
				advice := analyze(string(logs))
				fmt.Printf("Pod %s issue: %s\nFix: %s\n", p.Name, stateOf(cs), advice)
			}
		}
	}
}

func analyze(logs string) string {
	l := strings.ToLower(logs)
	switch {
	case strings.Contains(l, "imagepullbackoff"), strings.Contains(l, "manifest unknown"):
		return "Check ECR image, tag, and registry auth; run docker login to ECR."
	case strings.Contains(l, "crashloopbackoff"):
		return "Review CMD/ENTRYPOINT and env; add readiness/liveness; adjust minScale."
	case strings.Contains(l, "permission"), strings.Contains(l, "forbidden"):
		return "Bind IAM roles via IRSA to ServiceAccount; adjust RBAC."
	default:
		return "Inspect env, ConfigMaps/Secrets, resource limits; check Knative revision conditions."
	}
}

func stateOf(cs corev1.ContainerStatus) string {
	if cs.State.Waiting != nil { return cs.State.Waiting.Reason }
	if cs.State.Terminated != nil { return cs.State.Terminated.Reason }
	return "Unknown"
}

func getClient(namespace string) (*kubernetes.Clientset, string, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" { kubeconfig = clientcmd.RecommendedHomeFile }
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil { return nil, "", err }
	}
	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil { return nil, "", err }
	if namespace == "" { namespace = "default" }
	return clientset, namespace, nil
}

func getenv(k, d string) string { if v := os.Getenv(k); v != "" { return v }; return d }
