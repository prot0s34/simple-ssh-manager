package main

import "errors"

func (ig *InventoryGroup) Validate() error {
	if ig.Name == "" {
		return errors.New("inventory group name is required")
	}
	if err := ig.JumpHost.Validate(); err != nil {
		return err
	}
	if err := ig.KubeJumpHost.Validate(); err != nil {
		return err
	}
	if len(ig.TargetHost) == 0 {
		return errors.New("at least one target host is required")
	}
	for _, th := range ig.TargetHost {
		if err := th.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (jh *JumpHost) Validate() error {
	if jh.Hostname == "" || jh.Username == "" || jh.Password == "" {
		return errors.New("jump host requires hostname, username, and password")
	}
	return nil
}

func (kjh *KubeJumpHost) Validate() error {
	if kjh.KubeconfigPath == "" || kjh.Namespace == "" || kjh.Service == "" || kjh.ServicePort == 0 || kjh.LocalPort == 0 {
		return errors.New("kube jump host requires kubeconfigPath, namespace, service, servicePort, and localPort")
	}
	return nil
}

func (th *TargetHost) Validate() error {
	if th.Name == "" || th.Hostname == "" || th.Username == "" || th.Password == "" {
		return errors.New("target host requires name, hostname, username, and password")
	}
	return nil
}
