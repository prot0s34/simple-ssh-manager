# simple ssh manager 💻

![example workflow](https://github.com/prot0s34/simple-ssh-manager/actions/workflows/on_commit.yml/badge.svg/)
![GitHub tag (with filter)](https://img.shields.io/github/v/tag/prot0s34/simple-ssh-manager)

Lightweight SSH connections manager, support several jumphost options and multiply host-groups.
Written in Go with [tview](https://github.com/rivo/tview) and [kubernetes/client-go](https://github.com/kubernetes/client-go).

### Preview:
<p align="left">
    <img src="https://github.com/prot0s34/common-repo-stuff/blob/main/sshmanager-preview.gif" alt="Example">
</p>


### Inventory:
- support multiply lists in one inventory file
- support per-list jumphost and kubejumphost configs
- inventory should be in /home/$user/inventory.json or defined in ENV SSHMANAGER_INVENTORY=/path/to/inventory.json
- regular host, kubernetes pod or both (kubernetes -> jumphost -> targethost) can be used as jump option

### Jumphost options:
- **None** - localhost -> targethost
- **Kube❯Jump** - localhost -> kubernetes pod -> jumphost -> targethost
- **Kube** - localhost -> kubernetes pod -> targethost
- **Jump** - localhost -> jumphost -> targethost

### Kubernetes pod as jumphost - config:
- kubeJumpHostConfig.kubeconfigPath - path to kubeconfig file (default: ~/.kube/config)
- kubeJumpHostConfig.namespace - namespace of pod (used "default" if not defined)
- kubeJumpHostConfig.podName - name of pod. If podName not defined, podNameTemplate will be used for pod search (for generic pod name)
- kubeJumpHostConfig.podNameTemplate - template for pod name search

### Jumphost - config:
- JumpHostConfig.username - username for jumphost
- JumpHostConfig.password - password for jumphost
- JumpHostConfig.hostname - list of jumphosts

### How-To Use:
```
go build -o ./sshmanager .
cp sshmanager /usr/local/bin/
chmod +x /usr/local/bin/sshmanager
```

### TODO - Features:
- [x] kubectl jumphost functional
- [x] kubectl+bastion jumphost functional
- [x] bastion(single regular host) jumphost functional
- [x] multiply lists support
- [x] ~use 1 inventory with two lists intead of separate inventory files~
- [ ] cover code with more error handling
- [ ] use only fist pod name if search with template
- [ ] add ssh key-based auth support
- [ ] exclude "legend" information to bottom panel
- [ ] use tmux inside of app window instead of current behavior (close app->exec ssh in default terminal)
- [ ] use crypto/ssh for connection instead of exec ssh
- [ ] refac exec ssh commands (use ssh config file instead of command line args?)
- [ ] ssh command builder?
- [x] make release?
- [x] make CI/Actions?
- [x] add binary release to CI/Actions
- [ ] add tagging at pull requests to CI/Actions
- [ ] refac Hosts struct and optimize struct pass and use
- [ ] add echo "connected to $hostname" on each jumphost on the way to target host
- [x] add 'no strict host checking' for kube+jump option
- [ ] add "kubectl run debug --rm -i --tty \ --image=... \ --overrides='{"spec": { "nodeSelector": {"kubernetes.io/hostname": "some-node"}}}' -- bash" option?
- [ ] add kube context to inventory and kube functions 
- [x] cleanup binary from git history
- [x] wtf 50M binart, need to shrink binary size

### TO FIX:
- "Recovered from panic: runtime error: index out of range [n] with length n" after quit app with 'q'

### Changelog:
- 2023.10.29 add binary release to CI/Actions
- 2023.10.23 fix bug with selecting host for connect (affect lists that different from first list)
- 2023.10.22: added nested (kubernetes->jumphost) jump option, add regular jumphost option, back to single-list draw with ability to switch between lists, allow multiply lists in one inventory file, add separate jump configs per host, and so on (minor changes)
- 2023.10.21: added kubernetes jumphost support and modal dialog for jump options, fixed minor bugs

### Releases:
- v0.1.11 - init version
