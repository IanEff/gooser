// Package gooser provides a client for interacting with ArgoCD applications.
package gooser

// Application is the projection of an ArgoCD Application we care about.
type Application struct {
	Name   string
	Sync   string
	Health string
}
