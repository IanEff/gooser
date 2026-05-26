# gooser 🪿

A little cli app to goose ArgoCD to refresh Applications when your gRPC connection's
acting flaky.  Pulls your default local kubeconfig.

For ephemeral clusters and impatient people.

```
gooser              # pick an app from the TUI, goose it
gooser my-app       # skip the TUI, goose it directly
```

## How it works

ArgoCD applications are just Kubernetes custom resources. Rather than pulling in
`github.com/argoproj/argo-cd/v2` (and its several hundred transitive dependencies),
gooser talks to the Kubernetes API directly using a **dynamic client**.

### The GVR

Every resource in Kubernetes is identified by three coordinates:

| Field | Value |
|-------|-------|
| **Group** | `argoproj.io` |
| **Version** | `v1alpha1` |
| **Resource** | `applications` |

Together these form a `schema.GroupVersionResource` — the GVR. Given one, the dynamic
client can list, get, patch, or delete that resource without needing to know anything
about its schema at compile time.

```go
var applicationGVR = schema.GroupVersionResource{
    Group:    "argoproj.io",
    Version:  "v1alpha1",
    Resource: "applications",
}
```

### Listing apps

The dynamic client returns `unstructured.Unstructured` objects — essentially
`map[string]interface{}` wrappers around the raw JSON that came back from the API server.
To pull specific fields out without hand-rolling map traversal, the `unstructured` package
provides helpers:

```go
sync, _, _ := unstructured.NestedString(u.Object, "status", "sync", "status")
```

That walks `u.Object["status"]["sync"]["status"]` and returns the string value, a found
bool, and an error — all without a generated type in sight.

### Triggering the refresh

A hard refresh in ArgoCD is just a metadata annotation:

```
argocd.argoproj.io/refresh: hard
```

When the ArgoCD application controller sees that annotation on an `Application` CR, it
drops its cache and re-pulls from the Git source. gooser applies it with a
`MergePatchType` patch — the lightest-weight write operation the API server supports,
which only touches the fields you specify:

```go
appResource.Patch(ctx, app, types.MergePatchType, patchBytes, metav1.PatchOptions{})
```

ArgoCD removes the annotation itself once the refresh completes.

## Dependencies

| Package | Why |
|---------|-----|
| `k8s.io/client-go/dynamic` | Dynamic Kubernetes client — no generated types needed |
| `k8s.io/apimachinery` | GVR, `Unstructured`, patch types, list options |
| `charm.land/bubbletea/v2` | Terminal UI for app selection |

## Build & run

```sh
make run      # go run ./cmd/gooser
make build    # produces bin/gooser
make ci       # fmt-check → vet → lint → test → build
```
