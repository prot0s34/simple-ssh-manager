package main

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func initKubernetesClient(kubeconfigPath string) (*corev1.CoreV1Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := corev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func findPodByKeyword(client *corev1.CoreV1Client, namespace, keyword string) (string, error) {
	pods, err := client.Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, keyword) {
			return pod.Name, nil
		}
	}

	return "", fmt.Errorf("pod not found with keyword: %s", keyword)
}
