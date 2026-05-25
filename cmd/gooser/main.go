// Gooser gooses apps, baby.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ianeff/gooser/internal/gooser"
	"k8s.io/client-go/util/homedir"
)

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

	printApps(apps)

	name := appName
	if name == "" {
		return
	}

	if err := client.Goose(ctx, name); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
