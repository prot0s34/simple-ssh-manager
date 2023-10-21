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

type Host struct {
	Name               string `json:"name"`
	Hostname           string `json:"hostname"`
	Username           string `json:"username,omitempty"`
	Password           string `json:"password"`
	JumpHost           string `json:"jumpHost,omitempty"`
	KubeJumpHostConfig struct {
		KubeconfigPath  string `json:"kubeconfigPath,omitempty"`
		PodName         string `json:"podName,omitempty"`
		Namespace       string `json:"namespace,omitempty"`
		PodNameTemplate string `json:"podNameTemplate,omitempty"`
	} `json:"kubeJumpHostConfig,omitempty"`
}

type Inventory struct {
	InventoryName1  string `json:"inventory_name1"`
	InventoryGroup1 []Host `json:"inventory_group1"`
	InventoryName2  string `json:"inventory_name2"`
	InventoryGroup2 []Host `json:"inventory_group2"`
}

var inModalDialog = false

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()

	inventory, err := loadInventory()
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	app := tview.NewApplication()

	listHostsGroupFirst := createHostList(app, inventory.InventoryGroup1, inventory.InventoryName1)
	listHostsGroupSecond := createHostList(app, inventory.InventoryGroup2, inventory.InventoryName2)

	flex := tview.NewFlex().
		AddItem(listHostsGroupFirst, 0, 1, true).
		AddItem(listHostsGroupSecond, 0, 1, true)

	setHostListSelectedFunc(listHostsGroupFirst, inventory.InventoryGroup1, app, listHostsGroupFirst, listHostsGroupSecond)
	setHostListSelectedFunc(listHostsGroupSecond, inventory.InventoryGroup2, app, listHostsGroupFirst, listHostsGroupSecond)
	navigateBetweenFlexLists(app, listHostsGroupFirst, listHostsGroupSecond)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

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

	// If no matching pod was found, return an error
	return "", fmt.Errorf("Pod not found with keyword: %s", keyword)
}

func navigateBetweenFlexLists(app *tview.Application, list1, list2 *tview.List) {
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if inModalDialog {
			return event
		}

		inactiveColor := tcell.ColorBlack
		activeColor := tcell.ColorBlue

		if event.Key() == tcell.KeyLeft {
			app.SetFocus(list1)
			list1.SetBorderColor(activeColor).SetBorderAttributes(tcell.AttrBold)
			list2.SetBorderColor(inactiveColor)
		} else if event.Key() == tcell.KeyRight {
			app.SetFocus(list2)
			list1.SetBorderColor(inactiveColor)
			list2.SetBorderColor(activeColor).SetBorderAttributes(tcell.AttrBold)
		}
		return event
	})
}

func loadInventory() (*Inventory, error) {
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

	var inventory Inventory
	if err := json.Unmarshal(file, &inventory); err != nil {
		return nil, err
	}

	currentUser, _ := user.Current()
	for i := range inventory.InventoryGroup1 {
		if inventory.InventoryGroup1[i].Username == "" {
			inventory.InventoryGroup1[i].Username = currentUser.Username
		}
	}
	for i := range inventory.InventoryGroup2 {
		if inventory.InventoryGroup2[i].Username == "" {
			inventory.InventoryGroup2[i].Username = currentUser.Username
		}
	}

	return &inventory, nil
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

func setHostListSelectedFunc(list *tview.List, hosts []Host, app *tview.Application, listHostsGroupFirst, listHostsGroupSecond *tview.List) {
	list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		host := hosts[index]

		if host.Username == "" {
			fmt.Println("Error: Host username is missing in the inventory.")
			os.Exit(1)
		}

		inModalDialog = true

		dialog := tview.NewModal().
			SetText("Choose a jump option for host:" + host.Name).
			AddButtons([]string{"None", "Kube", "Cancel"}).
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

				case 1: // Kube
					if host.KubeJumpHostConfig.KubeconfigPath == "" {
						fmt.Println("Error: KubeconfigPath is missing in the inventory.")
						os.Exit(1)
					}

					clientset, err := initKubernetesClient(host.KubeJumpHostConfig.KubeconfigPath)
					if err != nil {
						fmt.Println("Error initializing Kubernetes client:", err)
						os.Exit(1)
					}

					if host.KubeJumpHostConfig.PodName == "" {
						podName, err := findPodByKeyword(clientset, host.KubeJumpHostConfig.Namespace, host.KubeJumpHostConfig.PodNameTemplate)
						if err != nil {
							fmt.Println("Error:", err)
							os.Exit(1)
						}
						host.KubeJumpHostConfig.PodName = podName
					}

					app.Stop()
					cmd := exec.Command("kubectl", "--kubeconfig", host.KubeJumpHostConfig.KubeconfigPath, "exec", "-it", host.KubeJumpHostConfig.PodName, "--", "sshpass", "-p", host.Password, "ssh", "-o", "StrictHostKeyChecking no", "-t", host.Username+"@"+host.Hostname)

					cmd.Stdout = os.Stdout
					cmd.Stdin = os.Stdin
					cmd.Stderr = os.Stderr

					if err := cmd.Run(); err != nil {
						fmt.Println("Error:", err)
						os.Exit(1)
					}

				case 2: // Cancel
					app.SetRoot(tview.NewFlex().
						AddItem(listHostsGroupFirst, 0, 1, true).
						AddItem(listHostsGroupSecond, 0, 1, true), true)
				}
				inModalDialog = false
			})

		app.SetRoot(dialog, true)
		navigateBetweenFlexLists(app, listHostsGroupFirst, listHostsGroupSecond)
	})
}
