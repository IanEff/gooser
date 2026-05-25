// Gooser gooses apps, baby.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

var applicationGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage: gooser [appname]")
		fmt.Println("Leave empty to list all applications")
		return
	}

	appName := ""
	if len(os.Args) == 2 {
		appName = os.Args[1]
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	appList, err := dynClient.Resource(applicationGVR).Namespace("argocd").List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Found %d Applications:\n", len(appList.Items))
	for _, app := range appList.Items {
		syncStatus, _, _ := unstructured.NestedString(app.Object, "status", "sync", "status")
		healthStatus, _, _ := unstructured.NestedString(app.Object, "status", "health", "status")
		fmt.Printf("\t- %-40s sync=%-12s health=%s\n", app.GetName(), syncStatus, healthStatus)
	}

	if appName == "" {
		fmt.Println("No patch requested.")
		return
	}

	patchPayload := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				"argocd.argoproj.io/refresh": "hard",
			},
		},
	}
	patchBytes, err := json.Marshal(patchPayload)
	if err != nil {
		panic(err.Error())
	}

	result, err := dynClient.Resource(applicationGVR).Namespace("argocd").Patch(
		context.TODO(),
		appName,
		types.MergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Patched %s - annotations: %v\n", result.GetName(), result.GetAnnotations())
}
