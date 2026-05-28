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
	defaultKubeconfig := os.Getenv("KUBECONFIG")
	if defaultKubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			defaultKubeconfig = filepath.Join(home, ".kube", "config")
		}
	}
	kubeconfig := flag.String("kubeconfig", defaultKubeconfig, "path to the kubeconfig file (default: $KUBECONFIG or ~/.kube/config)")
	flag.Parse()

	args := flag.Args()
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "Usage: gooser [appname] [--kubeconfig path]\n")
		os.Exit(1)
	}

	appName := ""
	if len(args) == 1 {
		appName = args[0]
	}

	switch appName {
	case "version", "--version", "-v":
		fmt.Printf("gooser %s (commit %s, built %s)\n", version, commit, date)
		return
	}

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
		if err = client.Goose(ctx, result.Application); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(tui.RenderGoosed(result.Application))
	case tui.ActionTwiddle:
		enabled, err := client.Twiddle(ctx, result.Application)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(tui.RenderTwiddle(result.Application, enabled))
	}
}
