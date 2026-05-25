package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ianeff/gooser/internal/gooser"
)

func printApps(apps []gooser.Application) {
	for i, app := range apps {
		fmt.Printf("[%d] %-40s sync=%-12s health=%s\n", i+1, app.Name, app.Sync, app.Health)
	}
}

func selectApp(apps []gooser.Application) (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("Select app [1-%d]: ", len(apps))
		if !scanner.Scan() {
			return "", fmt.Errorf("no selection made (EOF)")
		}
		n, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil || n < 1 || n > len(apps) {
			fmt.Fprintf(os.Stderr, "enter a number between 1 and %d\n", len(apps))
			continue
		}
		return apps[n-1].Name, nil
	}
}
