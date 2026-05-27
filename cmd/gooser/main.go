// Gooser gooses apps, baby.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ianeff/gooser/internal/gooser"
	"github.com/ianeff/gooser/internal/tui"
	"k8s.io/client-go/util/homedir"
)

// Stamped at link time via -ldflags "-X main.version=… -X main.commit=… -X main.date=…".
// GoReleaser and the Makefile both set these; the defaults apply to `go build`/`go run`.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage: gooser [appname] [flags]")
		fmt.Println("Leave empty to list all applications")
		return
	}

	appName := ""
	if len(os.Args) == 2 {
		appName = os.Args[1]
	}

	switch appName {
	case "version", "--version", "-v":
		fmt.Printf("gooser %s (commit %s, built %s)\n", version, commit, date)
		return
	}

	defaultKubeconfig := os.Getenv("KUBECONFIG")
	if defaultKubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			defaultKubeconfig = filepath.Join(home, ".kube", "config")
		}
	}
	kubeconfig := flag.String("kubeconfig", defaultKubeconfig, "path to the kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	flag.Parse()

	client, err := gooser.NewClient(*kubeconfig)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ctx := context.TODO()
	apps, err := client.List(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	found := false
	for _, app := range apps {
		if app.Name == appName {
			found = true
			break
		}
	}
	if appName != "" && !found {
		fmt.Fprintf(os.Stderr, "Application %q not found in Namespace argocd\n", appName)
		os.Exit(1)
	}

	result := tui.Result{Application: appName, Action: tui.ActionGoose}
	if appName == "" {
		result, err = tui.Run(apps)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	switch result.Action {
	case tui.ActionGoose:
		err = client.Goose(ctx, result.Application)
		fmt.Printf("Goosed %s. 🪿\n", result.Application)
	case tui.ActionTwiddle:
		var enabled bool
		enabled, err = client.Twiddle(ctx, result.Application)
		verb := "Disabled"
		if enabled {
			verb = "Enabled"
		}
		fmt.Printf("%s auto-sync for %s.\n", verb, result.Application)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
