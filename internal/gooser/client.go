package gooser

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var applicationGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

type Client struct {
	dyn dynamic.Interface
	ns  string
}

func NewClient(kubeconfig string) (*Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("cannot load kubeconfig: %w", err)
	}
	dyn, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("cannot create dynamic client: %w", err)
	}
	return &Client{dyn: dyn, ns: "argocd"}, nil
}

func (c *Client) List(ctx context.Context) ([]Application, error) {
	apps := []Application{}

	appList, err := c.dyn.Resource(applicationGVR).Namespace(c.ns).List(
		context.TODO(),
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list applications: %w", err)
	}

	for _, app := range appList.Items {
		apps = append(apps, Application{
			Name: app.GetName(),
			// This also works: syncStatus, _, _ := unstructured.NestedString(app.Object, "status", "sync", "status")
			// healthStatus, _, _ := unstructured.NestedString(app.Object, "status", "health", "status")
			Sync:   app.GetAnnotations()["sync-status"],
			Health: app.GetAnnotations()["health-status"],
		})
	}

	return apps, nil
}

func (c *Client) Goose(ctx context.Context, app string) error {
	appResource := c.dyn.Resource(applicationGVR).Namespace(c.ns)

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
	_, err = appResource.Patch(ctx, app, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("cannot patch application: %w", err)
	}
	return nil
}
