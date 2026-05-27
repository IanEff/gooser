package gooser

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var applicationGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applications",
}

// Client wraps the Kubernetes dynamic client for ArgoCD application operations.
type Client struct {
	dyn dynamic.Interface
	ns  string
}

// NewClient constructs a Client from a kubeconfig file path.
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

// List returns all ArgoCD applications in the configured namespace.
func (c *Client) List(ctx context.Context) ([]Application, error) {
	apps := []Application{}

	appList, err := c.dyn.Resource(applicationGVR).Namespace(c.ns).List(
		ctx,
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("cannot list applications: %w", err)
	}

	for _, item := range appList.Items {
		apps = append(apps, appFrom(item))
	}

	return apps, nil
}

func appFrom(u unstructured.Unstructured) Application {
	sync, _, _ := unstructured.NestedString(u.Object, "status", "sync", "status")
	health, _, _ := unstructured.NestedString(u.Object, "status", "health", "status")
	return Application{
		Name:   u.GetName(),
		Sync:   sync,
		Health: health,
	}
}

// Goose triggers a hard refresh on the named ArgoCD application.
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
		return fmt.Errorf("cannot marshal patch: %w", err)
	}
	_, err = appResource.Patch(ctx, app, types.MergePatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("cannot patch application: %w", err)
	}
	return nil
}

// Twiddle toggles automated sync policy for the named ArgoCD application,
// reading the current state from the resource and returning the new state
// (true means automated sync is now enabled).
func (c *Client) Twiddle(ctx context.Context, app string) (bool, error) {
	appResource := c.dyn.Resource(applicationGVR).Namespace(c.ns)

	current, err := appResource.Get(ctx, app, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("cannot get application: %w", err)
	}
	_, wasEnabled, err := unstructured.NestedMap(current.Object, "spec", "syncPolicy", "automated")
	if err != nil {
		return false, fmt.Errorf("cannot read sync policy: %w", err)
	}

	var automated interface{}
	if !wasEnabled {
		automated = map[string]interface{}{"selfHeal": true, "prune": true}
	} // else: nil — MergePatch with null removes the field, disabling auto-sync.

	patchPayload := map[string]interface{}{
		"spec": map[string]interface{}{
			"syncPolicy": map[string]interface{}{
				"automated": automated,
			},
		},
	}
	patchBytes, err := json.Marshal(patchPayload)
	if err != nil {
		return false, fmt.Errorf("cannot marshal patch: %w", err)
	}
	if _, err = appResource.Patch(ctx, app, types.MergePatchType, patchBytes, metav1.PatchOptions{}); err != nil {
		return false, fmt.Errorf("cannot patch application: %w", err)
	}
	return !wasEnabled, nil
}
