package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/ianeff/gooser/internal/gooser"
)

func printApps(apps []gooser.Application) {
	for _, app := range apps {
		fmt.Printf("\t- %-40s sync=%-12s health=%s\n", app.Name, app.Sync, app.Health)
	}
}

func selectApp(apps []gooser.Application) (string, error) {
	input := bufio.NewReader(os.Stdin)
	fmt.Print("Select an app: ")
	appNum, err := strconv.Atoi(input.ReadString('\n'))
	if err != nil {
		return "", err
	}
	if appNum < 1 || appNum > len(apps) {
		return "", fmt.Errorf("invalid app number: %d", appNum)
	}

	return apps[appNum-1].Name, nil
}
