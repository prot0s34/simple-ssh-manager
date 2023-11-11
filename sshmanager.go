package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var inModalDialog = false
var inventoryIndex = 0

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	app := tview.NewApplication()

	inventoryGroups, err := loadInventoryGroups()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	inventoryIndex = 0

	listHostsGroup := createHostList(app, inventoryGroups[inventoryIndex].Hosts, inventoryGroups[inventoryIndex].Name)

	setHostListSelectedFunc(listHostsGroup, inventoryGroups[inventoryIndex].Hosts, app, inventoryGroups, listHostsGroup)

	navigateBetweenInventoryGroups(app, &inventoryIndex, inventoryGroups, listHostsGroup)

	if err := app.SetRoot(listHostsGroup, true).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func loadInventoryGroups() ([]InventoryGroup, error) {
	inventoryPath := os.Getenv("SSHMANAGER_INVENTORY")

	if inventoryPath == "" {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		inventoryPath = usr.HomeDir + "/inventory.json"
	}

	file, err := ioutil.ReadFile(inventoryPath)
	if err != nil {
		return nil, err
	}

	var inventoryGroups []InventoryGroup
	if err := json.Unmarshal(file, &inventoryGroups); err != nil {
		return nil, err
	}

	return inventoryGroups, nil
}

func initKubernetesClient(kubeconfigPath string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func findPodByKeyword(clientset *kubernetes.Clientset, namespace, keyword string) (string, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		//		LabelSelector: "your-label-selector", // add functinal to match label selector if defined
	})
	if err != nil {
		return "", err
	}

	for _, pod := range pods.Items {
		if strings.Contains(pod.Name, keyword) {
			return pod.Name, nil
		}
	}

	return "", fmt.Errorf("Pod not found with keyword: %s", keyword)
}

func updateHostList(app *tview.Application, list *tview.List, hosts []Host, inventoryName string) {
	for _, host := range hosts {
		list.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil).SetBorder(true)
		list.SetTitle("[black:darkcyan]" + inventoryName + "[white:-]").SetTitleAlign(tview.AlignLeft)
		list.Box.SetBorderAttributes(tcell.RuneBoard)
	}
	list.AddItem("", "'Q' to Quit, < or > to change host list, 'Enter' to connect", 'q', func() {
		app.Stop()
	})
}

func createHostList(app *tview.Application, hosts []Host, inventoryName string) *tview.List {
	list := tview.NewList()

	for _, host := range hosts {
		list.AddItem("name:"+host.Name, "user:"+host.Username+" / hostname:"+host.Hostname, 0, nil).SetBorder(true)
		list.SetTitle("[black:darkcyan]" + inventoryName + "[white:-]").SetTitleAlign(tview.AlignLeft)
		list.Box.SetBorderAttributes(tcell.RuneBoard)
	}

	list.AddItem("", "'Q' to Quit, < or > to change host list, 'Enter' to connect", 'q', func() {
		app.Stop()
	})

	return list
}

func setHostListSelectedFunc(list *tview.List, hosts []Host, app *tview.Application, inventoryGroups []InventoryGroup, listHostsGroup *tview.List) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := inventoryGroups[inventoryIndex].Hosts[index]

		if host.Username == "" {
			fmt.Println("Error: Host username is missing in the inventory.")
			os.Exit(1)
		}

		inModalDialog = true

		dialog := tview.NewModal().
			SetText("Choose a jumphost option for host: " + host.Name).
			AddButtons([]string{"None", "Kube❯Jump", "Kube", "Jump", "Cancel"}).
			SetDoneFunc(func(buttonIndex int, buttonLabel string) {
				inModalDialog = false
				switch buttonIndex {
				case 0: // None
					app.Stop()
					cmd := exec.Command("sshpass", "-p", host.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", host.Username+"@"+host.Hostname)
					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 1: // Kube + Jump
					if inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath == "" {
						fmt.Println("Error: KubeconfigPath is missing in the inventory.")
						os.Exit(1)
					}

					clientset, err := initKubernetesClient(inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath)
					if err != nil {
						fmt.Println("Error initializing Kubernetes client:", err)
						os.Exit(1)
					}

					if inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName == "" {
						podName, err := findPodByKeyword(clientset, inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace, inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodNameTemplate)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName = podName
					}

					app.Stop()

					cmd := exec.Command("kubectl", "--kubeconfig", inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath, "exec", "-it", inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName, "--", "sshpass", "-p", inventoryGroups[inventoryIndex].JumpHostConfig.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", inventoryGroups[inventoryIndex].JumpHostConfig.Username+"@"+inventoryGroups[inventoryIndex].JumpHostConfig.Hostname, "sshpass", "-p", "'"+host.Password+"'", "ssh", "-o", "'StrictHostKeyChecking no'", host.Username+"@"+host.Hostname)

					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 2: // Kube
					if inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath == "" {
						fmt.Println("Error: KubeconfigPath is missing in the inventory.")
						os.Exit(1)
					}

					clientset, err := initKubernetesClient(inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath)
					if err != nil {
						fmt.Println("Error initializing Kubernetes client:", err)
						os.Exit(1)
					}

					if inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName == "" {
						podName, err := findPodByKeyword(clientset, inventoryGroups[inventoryIndex].KubeJumpHostConfig.Namespace, inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodNameTemplate)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName = podName
					}

					app.Stop()
					cmd := exec.Command("kubectl", "--kubeconfig", inventoryGroups[inventoryIndex].KubeJumpHostConfig.KubeconfigPath, "exec", "-it", inventoryGroups[inventoryIndex].KubeJumpHostConfig.PodName, "--", "sshpass", "-p", host.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", host.Username+"@"+host.Hostname)

					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 3: // Jump
					app.Stop()
					cmd := exec.Command("sshpass", "-p", inventoryGroups[inventoryIndex].JumpHostConfig.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", inventoryGroups[inventoryIndex].JumpHostConfig.Username+"@"+inventoryGroups[inventoryIndex].JumpHostConfig.Hostname, "sshpass", "-p", host.Password, "ssh", "-o", "'StrictHostKeyChecking no'", "-t", host.Username+"@"+host.Hostname)
					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 4: // Cancel
					inModalDialog = false
					app.SetRoot(listHostsGroup, true)
				}
			})

		app.SetRoot(dialog, true)
		navigateBetweenInventoryGroups(app, &inventoryIndex, inventoryGroups, listHostsGroup)
	})
}

func navigateBetweenInventoryGroups(app *tview.Application, inventoryIndex *int, inventoryGroups []InventoryGroup, listHostsGroup *tview.List) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if inModalDialog {
			return event
		}

		if event.Key() == tcell.KeyLeft {
			*inventoryIndex = (*inventoryIndex - 1 + len(inventoryGroups)) % len(inventoryGroups)
			listHostsGroup.Clear()
			updateHostList(app, listHostsGroup, inventoryGroups[*inventoryIndex].Hosts, inventoryGroups[*inventoryIndex].Name)
		} else if event.Key() == tcell.KeyRight {
			*inventoryIndex = (*inventoryIndex + 1) % len(inventoryGroups)
			listHostsGroup.Clear()
			updateHostList(app, listHostsGroup, inventoryGroups[*inventoryIndex].Hosts, inventoryGroups[*inventoryIndex].Name)
		}

		return event
	})
}
