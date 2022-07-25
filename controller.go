package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"log"
	"strings"

	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Read user config file
	var configFile = ReadConfig()

	// Read in and parse Kubernetes config
	kubeconfig := flag.String("kubeconfig", os.Getenv("HOME")+"/.kube/config", "kubeconfig file")

	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("erorr %s building config from flags\n", err.Error())

		config, err = rest.InClusterConfig()
		if err != nil {
			fmt.Printf("error %s, getting inclusterconfig", err.Error())
		}
	}

	// Generate a new 'Clientset' which is used to interact with the Kubernetes API
	clientset, createClientErr := kubernetes.NewForConfig(config)

	if createClientErr != nil {
		panic(createClientErr)
	}

	// Generate an Informer factory so that we can listen in to built-in resources
	informerFactory := informers.NewSharedInformerFactory(clientset, time.Second*20)

	namespaceInformer := informerFactory.Core().V1().Namespaces()

	// Add event handlers!
	// This constitutes the "Read" phase of the control loop

	namespaceInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			namespaceObj := obj.(*corev1.Namespace)

			// check if the namespace contains "dev"
			if strings.Contains(namespaceObj.Name, configFile.NamespaceShouldContain) {
				fmt.Printf("Found namespace that contains %s: %s \n", configFile.NamespaceShouldContain, namespaceObj.Name)

				// find existing Resource Quotas
				existingResourceQuotas, _ := GetResourceQuotas(clientset, context.Background(), namespaceObj.Name)

				// check if there is any ResourceQuota in namespace
				if len(existingResourceQuotas) == 0 {
					fmt.Printf("did not found any resource quotas in namespace %s\n", namespaceObj.Name)
					CreateCustomResourceQuota(namespaceObj.Name, clientset, configFile)
				} else {
					// if there is any, check if the mem-cpu-dev-quota existing
					for _, quota := range existingResourceQuotas {
						if quota == configFile.ResourceQuotaName {
							fmt.Printf("ResourceQuota: %s already exists in namespace %s. skipping\n", quota, namespaceObj.Name)
						} else {
							CreateCustomResourceQuota(namespaceObj.Name, clientset, configFile)
						}

					}
				}

			}
		},
		DeleteFunc: func(obj interface{}) {
		},
	})

	stop := make(chan struct{})
	defer close(stop)

	// Start our informers
	informerFactory.Start(wait.NeverStop)
	informerFactory.WaitForCacheSync(wait.NeverStop)

	// Wait on the 'stop' channel. This let's this process continue on running.
	// You can stop it with 'CTRL-c' when running directly from a terminal
	select {}
}

func GetResourceQuotas(clientset *kubernetes.Clientset, ctx context.Context,
	namespace string) ([]string, error) {

	list, err := clientset.CoreV1().ResourceQuotas(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var listOfQuota []string

	for _, item := range list.Items {
		listOfQuota = append(listOfQuota, item.Name)
	}
	return listOfQuota, nil
}

func CreateCustomResourceQuota(namespace string, clientset *kubernetes.Clientset, configFile Config) {
	fmt.Printf("creating ResourceQuota in namespace: %s \n", namespace)
	quotaLimit := corev1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configFile.ResourceQuotaName,
			Namespace: namespace,
		},
		Spec: corev1.ResourceQuotaSpec{
			Hard: corev1.ResourceList{
				corev1.ResourceName(corev1.ResourcePods):           resource.MustParse(configFile.ResourcePods),
				corev1.ResourceName(corev1.ResourceRequestsMemory): resource.MustParse(configFile.ResourceRequestsMemory),
				corev1.ResourceName(corev1.ResourceLimitsMemory):   resource.MustParse(configFile.ResourceLimitsMemory),
				corev1.ResourceName(corev1.ResourceLimitsCPU):      resource.MustParse(configFile.ResourceLimitsCPU),
			},
		},
	}

	createdResourceQuota, createErr := clientset.CoreV1().ResourceQuotas(namespace).Create(context.TODO(), &quotaLimit, metav1.CreateOptions{})

	if createErr != nil {
		fmt.Printf("Error creating quota: %s \n", createErr)
		panic(createErr)
	} else {
		fmt.Printf("quota created %s \n", createdResourceQuota.Name)
	}

}

// Config Info from config file
type Config struct {
	NamespaceShouldContain string
	ResourceQuotaName      string
	ResourcePods           string
	ResourceRequestsMemory string
	ResourceLimitsMemory   string
	ResourceLimitsCPU      string
}

func ReadConfig() Config {
	var configfile = "./configFile"
	_, err := os.Stat(configfile)
	if err != nil {
		log.Fatal("Config file is missing: ", configfile)
	}

	var config Config
	if _, err := toml.DecodeFile(configfile, &config); err != nil {
		log.Fatal(err)
	}
	//log.Print(config.Index)
	return config
}
