# Invoice Service

This sample Flogo application is used to demonstrate some key Flogo constructs, can be deployed to Kubernetes, and is set to be indexed by API Scout

## Files
```bash
.
├── Dockerfile          <-- A Dockerfile to build a container based on an Alpine base image
├── main.go             <-- The Go source code for the app
├── Makefile            <-- A Makefile to help build and deploy the app
├── payment-go-svc.yml  <-- The Kubernetes deployment file
├── README.md           <-- This file
└── swagger.json        <-- The OpenAPI specification for the app
```

## Make targets
The [Makefile](./Makefile) has a few targets:
* **deps**: Get all the Go dependencies for the app
* **clean**: Remove the `dist` folder for a new deployment
* **clean-kube**: Remove all the deployed artifacts from Kubernetes
* **build-app**: Build an executable (and store it in the dist folder)
* **build-docker**: Build a Docker container from the contents of the `dist` folder
* **run-docker**: Run the Docker image with default settings
* **run-kube**: Deploy the app to Kubernetes

_For more detailed information on the commands that are executed you can check out the [Makefile](./Makefile)_

## Build and deploy the app
To build and deploy the app to Kubernetes, run the make targets for _deps_, _build-app_, _build-docker_ and _run-kube_

## API
After starting the app, it will register with two endpoints:
* **/api/invoices/:id**: Get the invoice details for the invoice ID.
* **/swagger**: Get the OpenAPI specification for this app

## API Scout
As you deploy the app to Kubernetes, after a few seconds the API will be found by API Scout and indexed. The lines 36 to 38 in [invoice-go-svc.yml](./invoice-go-svc.yml) are the annotations that make sure the API is found.