package main

import (
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
