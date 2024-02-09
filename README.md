# simple ssh manager 📟

<h1>
  <a href="#--------">
    <img alt="" align="right" src="https://img.shields.io/github/v/tag/prot0s34/simple-ssh-manager"/>
  </a>
  <a href="#--------">
    <img alt="" align="left" src="https://github.com/prot0s34/simple-ssh-manager/actions/workflows/on_commit.yml/badge.svg/"/>
  </a>
</h1>


<p>Terminal based SSH connections manager</p>
<p>Allow you create multiply "inventory" host lists and connect into target host with few key motions</p>
<p>You can connect to target host using several "hop" options - via regular jumphost, kubernetes service (with SOCKS5 proxy), kubernetes service(with SOCKS5 proxy) + jumphost as proxy, direct connection</p>
<p>For using k8s service as proxy - you need to install SOCKS5 pod in your cluster and create service for it. See, for example, [dante](https://www.inet.no/dante/)</p>
<p>Currently it support only password-based auth (btw crypto/ssh used, BUT passwords in inventory stored as plaintext, keep in mind)</p>
<p>Written in Go with <a href=https://github.com/rivo/tview> rivo/tview</a> </p>

### Preview:
<p align="left">
    <img src="https://github.com/prot0s34/common-repo-stuff/blob/main/sshmanager-preview.gif" alt="Example">
</p>


### 📓 Inventory:
- support multiply hosts groups (lists) in one inventory file
- support separate jumphost and kubejumphost configs for each hosts group
- inventory should be in /home/$user/inventory.json or defined in ENV SSHMANAGER_INVENTORY=/path/to/inventory.json
- regular host, kubernetes SOCKS5 proxy service or both (kubernetes -> jumphost -> targethost) can be used as jump option

### 🔌 Jumphost options:
- **None** - localhost -> targethost
- **Kube❯Jump** - localhost -> kubernetes SOCKS5 service -> jumphost -> targethost
- **Kube** - localhost -> kubernetes SOCKS5 service -> targethost
- **Jump** - localhost -> jumphost -> targethost

### 🔧 Configuration:
```
🚢 Kubernetes SOCKS5 service as Jumphost:
- kubeJumpHostConfig.kubeconfigPath - path to kubeconfig file (default: ~/.kube/config)
- kubeJumpHostConfig.namespace - namespace of service
- kubeJumpHostConfig.service - name of service with SOCKS5 proxy. 

🔗 Jumphost - config:
- JumpHostConfig.username - username for jumphost
- JumpHostConfig.password - password for jumphost
- JumpHostConfig.hostname - hostname for jumphost


```
### 🚥 How-To Use:
```
go build -o sshmanager ./ssm
cp sshmanager /usr/local/bin/
chmod +x /usr/local/bin/sshmanager
```

### ✅ TODO - Features:
- [x] kubectl jumphost functional
- [x] kubectl+bastion jumphost functional
- [x] bastion(single regular host) jumphost functional
- [x] multiply lists support
- [x] ~use 1 inventory with two lists intead of separate inventory files~
- [x] use crypto/ssh for connection instead of exec ssh
- [x] refac exec ssh commands (use ssh config file instead of command line args?)
- [x] ssh command builder?
- [x] make release?
- [x] make CI/Actions?
- [x] add binary release to CI/Actions
- [x] add echo "connected to $hostname" on each jumphost on the way to target host
- [x] add 'no strict host checking' for kube+jump option
- [x] cleanup binary from git history
- [x] ~wtf 50M binary~, shrinked to 31MB, need to drop/replace go-client for kubernetes for more lightweight binary :(
- [x] reuse socks5 for multiply connections
- [ ] add option for creating local proxy for :DistantConnect sessions?
- [ ] additional packaging?
- [ ] cover code with more error handling
- [ ] add ssh key-based auth support
- [ ] exclude "legend" information to bottom panel
- [ ] use tmux inside of app window instead of current behavior (close app->exec ssh in default terminal)
- [ ] add tagging at pull requests to CI/Actions
- [ ] refac Hosts struct and optimize struct pass and use
- [ ] proxy via <a href=https://github.com/kubernetes/client-go>kubernetes/client-go</a> instead kubectl?
- [ ] add kube context to inventory and kube functions 
- [ ] yaml inventory?
- [ ] vim-like command mode for :q and :/ ? 
- [ ] encrypt inventory? fetch passwords from bitwarden?
- [ ] make default service namespace fallback
- [ ] make logging more cute and compact 

### ⚠️ TO FIX:
- "Recovered from panic: runtime error: index out of range [n] with length n" after quit app with 'q' (meanwhile, signal from ctrl+c handled correctly)

### ⛽ Changelog:
- 2024.02.07 refactoring connection func, change pod port-forwarding to service port-forwarding
- 2024.02.06 huge refactoring of ssh connections (sshpass bye-bye, welcome crypto/ssh lib) and implementing SOCKS5 k8s proxy with port-forwarding
- 2024.01.28 v0.1.12 add minor improvments (as print connstring), refactoring, ssh args structure, binary size optimization, and so on.
- 2023.10.29 add binary release to CI/Actions
- 2023.10.23 fix bug with selecting host for connect (affect lists that different from first list)
- 2023.10.22: added nested (kubernetes->jumphost) jump option, add regular jumphost option, back to single-list draw with ability to switch between lists, allow multiply lists in one inventory file, add separate jump configs per host, and so on (minor changes)
- 2023.10.21: added kubernetes jumphost support and modal dialog for jump options, fixed minor bugs

### 🏁 Releases:
- v0.1.12 - minor fixes & improvments
- v0.1.11 - init version
