package gooser

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"errors"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

var errInjected = errors.New("injected API error")

// makeApp builds a minimal ArgoCD Application as an Unstructured object.
func makeApp(name, sync, health string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "argoproj.io/v1alpha1",
			"kind":       "Application",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": "argocd",
			},
			"status": map[string]interface{}{
				"sync":   map[string]interface{}{"status": sync},
				"health": map[string]interface{}{"status": health},
			},
		},
	}
}

// firstPatch returns the first PatchAction recorded by a fake dynamic client, or nil.
func firstPatch(fd *fake.FakeDynamicClient) k8stesting.PatchAction {
	for _, action := range fd.Actions() {
		if pa, ok := action.(k8stesting.PatchAction); ok {
			return pa
		}
	}
	return nil
}

// appWithAutoSync builds an Application whose spec.syncPolicy.automated is set,
// i.e. auto-sync is currently enabled.
func appWithAutoSync(name string) *unstructured.Unstructured {
	u := makeApp(name, "Synced", "Healthy")
	_ = unstructured.SetNestedMap(u.Object, map[string]interface{}{
		"selfHeal": true,
		"prune":    true,
	}, "spec", "syncPolicy", "automated")
	return u
}

// fakeClient returns a *Client backed by a fake dynamic client pre-loaded with apps.
func fakeClient(apps ...*unstructured.Unstructured) *Client {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(
		applicationGVR.GroupVersion().WithKind("Application"),
		&unstructured.Unstructured{},
	)
	scheme.AddKnownTypeWithName(
		applicationGVR.GroupVersion().WithKind("ApplicationList"),
		&unstructured.UnstructuredList{},
	)

	gvrToListKind := map[schema.GroupVersionResource]string{
		applicationGVR: "ApplicationList",
	}

	objs := make([]runtime.Object, len(apps))
	for i, a := range apps {
		objs[i] = a
	}

	dyn := fake.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToListKind, objs...)
	return &Client{dyn: dyn, ns: "argocd"}
}

func TestAppFrom(t *testing.T) {
	tests := []struct {
		name  string
		input unstructured.Unstructured
		want  Application
	}{
		{
			name:  "all fields present",
			input: *makeApp("my-app", "Synced", "Healthy"),
			want:  Application{Name: "my-app", Sync: "Synced", Health: "Healthy"},
		},
		{
			name: "missing status block",
			input: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{"name": "bare"},
				},
			},
			want: Application{Name: "bare", Sync: "", Health: ""},
		},
		{
			name: "partial status — sync only",
			input: unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{"name": "half"},
					"status": map[string]interface{}{
						"sync": map[string]interface{}{"status": "OutOfSync"},
					},
				},
			},
			want: Application{Name: "half", Sync: "OutOfSync", Health: ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appFrom(tt.input)
			if got != tt.want {
				t.Errorf("appFrom() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name string
		apps []*unstructured.Unstructured
		want []Application
	}{
		{
			name: "empty cluster",
			apps: nil,
			want: []Application{},
		},
		{
			name: "single app",
			apps: []*unstructured.Unstructured{
				makeApp("app-a", "Synced", "Healthy"),
			},
			want: []Application{
				{Name: "app-a", Sync: "Synced", Health: "Healthy"},
			},
		},
		{
			name: "multiple apps",
			apps: []*unstructured.Unstructured{
				makeApp("app-a", "Synced", "Healthy"),
				makeApp("app-b", "OutOfSync", "Degraded"),
			},
			want: []Application{
				{Name: "app-a", Sync: "Synced", Health: "Healthy"},
				{Name: "app-b", Sync: "OutOfSync", Health: "Degraded"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := fakeClient(tt.apps...)
			got, err := c.List(context.Background())
			if err != nil {
				t.Fatalf("List() unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("List() returned %d apps, want %d", len(got), len(tt.want))
			}
			// Index by name — fake client doesn't guarantee insertion order.
			byName := make(map[string]Application, len(got))
			for _, a := range got {
				byName[a.Name] = a
			}
			for _, want := range tt.want {
				got, ok := byName[want.Name]
				if !ok {
					t.Errorf("app %q missing from results", want.Name)
					continue
				}
				if got != want {
					t.Errorf("app %q: got %+v, want %+v", want.Name, got, want)
				}
			}
		})
	}
}

func TestListError(t *testing.T) {
	t.Run("propagates API errors", func(t *testing.T) {
		c := fakeClient()
		c.dyn.(*fake.FakeDynamicClient).PrependReactor("list", "applications",
			func(k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, errInjected
			},
		)
		if _, err := c.List(context.Background()); err == nil {
			t.Fatal("List() expected error, got nil")
		}
	})
}

func TestGoose(t *testing.T) {
	t.Run("sends merge patch with hard refresh annotation", func(t *testing.T) {
		c := fakeClient(makeApp("my-app", "Synced", "Healthy"))

		if err := c.Goose(context.Background(), "my-app"); err != nil {
			t.Fatalf("Goose() unexpected error: %v", err)
		}

		pa := firstPatch(c.dyn.(*fake.FakeDynamicClient))
		if pa == nil {
			t.Fatal("no patch action recorded")
		}
		if pa.GetPatchType() != types.MergePatchType {
			t.Errorf("patch type = %v, want MergePatchType", pa.GetPatchType())
		}

		var body map[string]interface{}
		if err := json.Unmarshal(pa.GetPatch(), &body); err != nil {
			t.Fatalf("unmarshal patch body: %v", err)
		}
		annotations, _, _ := unstructured.NestedStringMap(body, "metadata", "annotations")
		if got := annotations["argocd.argoproj.io/refresh"]; got != "hard" {
			t.Errorf("refresh annotation = %q, want %q", got, "hard")
		}
	})

	t.Run("propagates API errors", func(t *testing.T) {
		c := fakeClient(makeApp("my-app", "Synced", "Healthy"))
		// Inject a failure for any patch on applications.
		c.dyn.(*fake.FakeDynamicClient).PrependReactor("patch", "applications",
			func(k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, errInjected
			},
		)
		if err := c.Goose(context.Background(), "my-app"); err == nil {
			t.Fatal("Goose() expected error, got nil")
		}
	})
}

func TestTwiddle(t *testing.T) {
	t.Run("when auto-sync is off, patch enables selfHeal and prune and returns true", func(t *testing.T) {
		c := fakeClient(makeApp("my-app", "Synced", "Healthy"))

		enabled, err := c.Twiddle(context.Background(), "my-app")
		if err != nil {
			t.Fatalf("Twiddle() unexpected error: %v", err)
		}
		if !enabled {
			t.Errorf("Twiddle() returned enabled=false, want true")
		}

		pa := firstPatch(c.dyn.(*fake.FakeDynamicClient))
		if pa == nil {
			t.Fatal("no patch action recorded")
		}
		if pa.GetPatchType() != types.MergePatchType {
			t.Errorf("patch type = %v, want MergePatchType", pa.GetPatchType())
		}

		var body map[string]interface{}
		if err := json.Unmarshal(pa.GetPatch(), &body); err != nil {
			t.Fatalf("unmarshal patch body: %v", err)
		}
		for _, field := range []string{"selfHeal", "prune"} {
			got, ok, err := unstructured.NestedBool(body, "spec", "syncPolicy", "automated", field)
			if err != nil || !ok {
				t.Errorf("spec.syncPolicy.automated.%s missing from patch", field)
				continue
			}
			if !got {
				t.Errorf("spec.syncPolicy.automated.%s = false, want true", field)
			}
		}
	})

	t.Run("when auto-sync is on, patch clears automated and returns false", func(t *testing.T) {
		c := fakeClient(appWithAutoSync("my-app"))

		enabled, err := c.Twiddle(context.Background(), "my-app")
		if err != nil {
			t.Fatalf("Twiddle() unexpected error: %v", err)
		}
		if enabled {
			t.Errorf("Twiddle() returned enabled=true, want false")
		}

		pa := firstPatch(c.dyn.(*fake.FakeDynamicClient))
		if pa == nil {
			t.Fatal("no patch action recorded")
		}

		// MergePatch with a null value removes the field; assert the raw JSON
		// contains "automated":null rather than an object.
		if got := string(pa.GetPatch()); !strings.Contains(got, `"automated":null`) {
			t.Errorf("patch body = %s, want it to contain \"automated\":null", got)
		}
	})

	t.Run("propagates Get errors", func(t *testing.T) {
		c := fakeClient(makeApp("my-app", "Synced", "Healthy"))
		c.dyn.(*fake.FakeDynamicClient).PrependReactor("get", "applications",
			func(k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, errInjected
			},
		)
		if _, err := c.Twiddle(context.Background(), "my-app"); err == nil {
			t.Fatal("Twiddle() expected error, got nil")
		}
	})

	t.Run("propagates Patch errors", func(t *testing.T) {
		c := fakeClient(makeApp("my-app", "Synced", "Healthy"))
		c.dyn.(*fake.FakeDynamicClient).PrependReactor("patch", "applications",
			func(k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, errInjected
			},
		)
		if _, err := c.Twiddle(context.Background(), "my-app"); err == nil {
			t.Fatal("Twiddle() expected error, got nil")
		}
	})
}
