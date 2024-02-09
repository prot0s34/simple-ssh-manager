package main

type AppContext struct {
	InventoryIndex *int
	Inventory      []InventoryGroup
}

type InventoryGroup struct {
	Name         string       `json:"name"`
	JumpHost     JumpHost     `json:"jumpHostConfig"`
	KubeJumpHost KubeJumpHost `json:"kubeJumpHostConfig"`
	TargetHost   []TargetHost `json:"hosts"`
}

type TargetHost struct {
	Name     string `json:"name,omitempty"`
	Hostname string `json:"hostname,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type JumpHost struct {
	Hostname string `json:"hostname,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type KubeJumpHost struct {
	KubeconfigPath string `json:"kubeconfigPath,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	Service        string `json:"service,omitempty"`
	ServicePort    int    `json:"servicePort,omitempty"`
	LocalPort      int    `json:"localPort,omitempty"`
}
