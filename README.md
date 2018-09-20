# API Scout :: Finding Your APIs In Kubernetes

## What is API Scout
API Scout, helps you get up-to-date API docs to your developers by simply annotating your services in K8s. 

## What are the components of API Scout
The docker image that is deployed to Kubernetes has several components:
* The container itself is based on [nginx:1.15-alpine](https://hub.docker.com/_/nginx/)
* The webapp is a staticly generated site by [Hugo](https://github.com/gohugoio/hugo) using the [Learn](https://themes.gohugo.io/hugo-theme-learn/) theme and an additonal [shortcode for OpenAPI](https://github.com/tenfourty/hugo-openapispec-shortcode)
* A server app that connects to the Kubernetes cluster using a default role to watch for services that need to be indexed

_Hugo is downloaded and embedded during the build of the container_

## Build and run
apiscout has a _Makefile_ that can be used for most of the operations. Make sure you have installed Go Programming Language, set GOPATH variable and added $GOPATH/bin in your PATH

```
usage: make [target]
```

### Cleaning targets:
* **clean-all** : Removes the dist directory
* **clean-docker** : Stops and removes all containers and images for apiscout
* **clean-kube** : Removes the apiscout service and deployment from Kubernetes

### Cleaning targets:
* **deps** : Gets required dependencies. Run before the build-server target

### Build targets:
* **build-site** : Builds the Hugo distribution in dist
* **build-server** : Builds the Flogo server in dist
* **build-docker** : Builds a docker image from the dist directory
* **build-all** : Performs clean-all and executes all build targets

### Run targets
* **run-server** : Builds the in the server directory and runs it with default settings
* **run-docker** : Runs a docker container with default settings
* **run-hugo** : Runs the embedded Hugo server on port 1313
* **run-kube** : Deploys apiscout to Kubernetes

### Stop targets
* **stop-docker** : Stop and remove the running apiscout container

## Requirements for Kubernetes
To be able to view the services, apiscout needs access to the Kubernetes cluster using the default service account. By default (pun intended) this account doesn't have access to view services so during deployment a new _rolebinding_ is created. After starting, it will poll the Kubernetes API server every 10 seconds, or at the timeinterval specified by the environment variable `INTERVAL`.

apiscout looks for two annotations to be able to index a service:
* `apiscout/index: 'true'` This annotation ensures that apiscout indexes the service
* `apiscout/swaggerUrl: '/swaggerspec'` This is the URL from where apiscout will read the OpenAPI document

## Environment variables for the docker container
apiscout has a few environment variables that the docker container (and thus the deployment to Kubernetes) can use:
* **SWAGGERSTORE**: The location where to store the swaggerdocs
* **HUGOSTORE**: The location where to store content for Hugo
* **MODE**: The mode in which apiscout is running (can be either KUBE or LOCAL)
* **EXTERNALIP**: The external IP address of the Kubernetes cluster in case of LOCAL mode
* **HUGODIR**: The base directory for Hugo

## License
See the [LICENSE](./LICENSE) file
