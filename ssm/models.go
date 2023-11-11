package main

type Host struct {
	Name     string `json:"name"`
	Hostname string `json:"hostname"`
	Username string `json:"username,omitempty"`
	Password string `json:"password"`
}

type InventoryGroup struct {
	Name           string `json:"name"`
	JumpHostConfig struct {
		Hostname string `json:"hostname,omitempty"`
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
	}
	KubeJumpHostConfig struct {
		KubeconfigPath  string `json:"kubeconfigPath,omitempty"`
		PodName         string `json:"podName,omitempty"`
		Namespace       string `json:"namespace,omitempty"`
		PodNameTemplate string `json:"podNameTemplate,omitempty"`
	} `json:"kubeJumpHostConfig,omitempty"`
	Hosts []Host `json:"hosts"`
}
