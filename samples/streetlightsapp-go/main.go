//go:generate go run $GOPATH/src/github.com/TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH
package main

import (
	"context"
	"io/ioutil"
	"os"
	"strconv"

	rt "github.com/TIBCOSoftware/flogo-contrib/trigger/rest"
	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/engine"
	"github.com/TIBCOSoftware/flogo-lib/flogo"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/ghodss/yaml"
)

var (
	httpport = getEnvKey("HTTPPORT", "8090")
)

func main() {
	// Create a new Flogo app
	app := appBuilder()

	e, err := flogo.NewEngine(app)

	if err != nil {
		logger.Error(err)
		return
	}

	engine.RunEngine(e)
}

func appBuilder() *flogo.App {
	app := flogo.NewApp()

	// Convert the HTTPPort to an integer
	port, err := strconv.Atoi(httpport)
	if err != nil {
		logger.Error(err)
	}

	// Register the HTTP trigger
	trg := app.NewTrigger(&rt.RestTrigger{}, map[string]interface{}{"port": port})
	trg.NewFuncHandler(map[string]interface{}{"method": "GET", "path": "/asyncapispec"}, YamlSpec)

	return app
}

// getEnvKey tries to get the specified key from the OS environment and returns either the
// value or the fallback that was provided
func getEnvKey(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// YamlSpec is the function that gets executedto retrieve the SwaggerSpec
func YamlSpec(ctx context.Context, inputs map[string]*data.Attribute) (map[string]*data.Attribute, error) {
	// The return message is a map[string]*data.Attribute which we'll have to construct
	response := make(map[string]interface{})
	ret := make(map[string]*data.Attribute)

	fileData, err := ioutil.ReadFile("streetlightsapi.yaml")
	if err != nil {
		ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 500)
		response["msg"] = err.Error()
	} else {
		ret["code"], _ = data.NewAttribute("code", data.TypeInteger, 200)
		var data map[string]interface{}
		if err := yaml.Unmarshal(fileData, &data); err != nil {
			panic(err)
		}
		response = data
	}

	ret["data"], _ = data.NewAttribute("data", data.TypeAny, response)

	return ret, nil

}
