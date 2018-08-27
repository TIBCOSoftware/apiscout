//go:generate go run $GOPATH/src/github.com/TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH

// Package main
package main

// The imports
import (
	"log"
	"os"
	"strings"

	"github.com/TIBCOSoftware/flogo-contrib/trigger/timer"
	"github.com/TIBCOSoftware/flogo-lib/engine"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"
)

const (
	// The annotation for apiscout to index a service
	apiscoutAnnotation = "apiscout/index"
	// The annotation for apiscout to get the OpenAPI doc from
	apiscoutSwaggerURL = "apiscout/swaggerUrl"
	// A template for the Markdown file for Hugo
	markdown = `---
title: {{.title}}
weight: 1000
---

{{.description}}

{{.json}}`
)

var (
	// A map[string]string of all services that have been indexed by apiscout
	serviceMap = make(map[string]string)
	// The location where to store the swaggerdocs
	swaggerStore = getEnv("SWAGGERSTORE", "/tmp/static/swaggerdocs")
	// The location where to store content for Hugo
	hugoStore = getEnv("HUGOSTORE", "/tmp/content/apis")
	// The mode in which apiscout is running (can be either KUBE or LOCAL)
	runMode = getEnv("MODE", "LOCAL")
	// The external IP address of the Kubernetes cluster in case of LOCAL mode
	externalIP = getEnv("EXTERNALIP", "")
	// The base directory for Hugo
	hugoDir = getEnv("HUGODIR", "")
)

func main() {
	// Create a new Flogo app
	app := appBuilder()

	// Create the Flogo engine
	e, err := flogo.NewEngine(app)
	if err != nil {
		logger.Error(err)
		return
	}

	// Run the engine!
	engine.RunEngine(e)
}

func appBuilder() *flogo.App {
	// Create a new app instance
	app := flogo.NewApp()

	// Get the timer interval for the server
	timerInterval := getEnv("INTERVAL", "10")

	// Print config
	log.Printf("------------------------------------------------------------")
	log.Printf("CONFIG\n")
	log.Printf("The timer interval has been set to : %s seconds", timerInterval)
	log.Printf("apiscout run mode has been set to   : %s", runMode)
	log.Printf("swagger store has been set to      : %s", swaggerStore)
	log.Printf("hugo store has been set to         : %s", hugoStore)
	if strings.ToUpper(runMode) != "KUBE" {
		log.Printf("external ip has been set to        : %s", externalIP)
	}
	if len(hugoDir) > 0 {
		log.Printf("hugo dir has been set to           : %s", hugoDir)
	}
	log.Printf("------------------------------------------------------------")

	// Register the HTTP trigger
	trg := app.NewTrigger(&timer.TimerTrigger{}, nil)

	// Add handlers
	trg.NewFuncHandler(map[string]interface{}{"repeating": "true", "notImmediate": "false", "seconds": timerInterval}, timerHandler)

	return app
}

// getEnv gets an environment variable and returns the value if the variable was set or a default value if the variable was not set
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
