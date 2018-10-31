# API Scout :: Finding Your APIs In Kubernetes

## What is API Scout

API Scout, helps you get up-to-date API docs to your developers by simply annotating your services in K8s.

## What are the components of API Scout

The docker image that is deployed to Kubernetes has several components:

* The container itself is based on [nginx:1.15-alpine](https://hub.docker.com/_/nginx/)
* The webapp is a staticly generated site by [Hugo](https://github.com/gohugoio/hugo) using the [Learn](https://themes.gohugo.io/hugo-theme-learn/) theme and [Swagger UI](https://github.com/swagger-api/swagger-ui/releases)
* A server app that connects to the Kubernetes cluster using a default role to watch for services that need to be indexed

_Hugo is downloaded and embedded during the build of the container_

## Build and run

API Scout has a _Makefile_ that can be used for most of the operations. Make sure you have installed Go Programming Language, set GOPATH variable and added $GOPATH/bin in your PATH

```bash
usage: make [target]
```

| Makefile targets | Description                                                           |
|------------------|-----------------------------------------------------------------------|
| build-all        | Performs clean-all and executes all build targets                     |
| build-docker     | Builds a docker image from the dist directory                         |
| build-server     | Builds the server app in dist                                         |
| build-site       | Builds the Hugo distribution in dist                                  |
| clean-all        | Removes the dist directory                                            |
| clean-docker     | Stops and removes all containers and images for apiscout              |
| clean-kube       | Removes the apiscout service and deployment from Kubernetes           |
| deps             | Get dependencies to build the server                                  |
| minikube-delete  | Delete the Minikube installation                                      |
| minikube-install | Install Minikube on this machine                                      |
| minikube-show    | Show the API Scout UI that is deployed to Minikube                    |
| minikube-start   | Start Minikube with default configuration                             |
| minikube-stop    | Stop Minikube                                                         |
| run-docker       | Runs a docker container with default settings                         |
| run-docs         | Runs the embedded Hugo server on port 1313 for the documentation      |
| run-hugo         | Runs the embedded Hugo server on port 1313                            |
| run-kube         | Deploys apiscout to Kubernetes                                        |
| run-server       | Builds the  in the server directory and runs it with default settings |
| stop-docker      | Stop and remove the running apiscout container                        |

## Requirements for Kubernetes

To be able to view the services, apiscout needs access to the Kubernetes cluster using the default service account. By default (pun intended) this account doesn't have access to view services so during deployment a new _rolebinding_ is created. After starting, it will register a watcher with the Kubernetes API Server, so that it receives events when something occurs in the cluster.

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

## Getting started

This section provides minimal steps to get `apiscout` running inside a kubernetes cluster on local machine / VM of your choice.

### Prerequisites

* Docker 
* Kubernetes environment (for example [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/))

### Steps to follow

1. Start minikube
```bash
$ make minikube-start
```

2. Build and deploy apiscout

```bash
# Install dependencies
$ make deps

# Build apiscout docker image
$ make build-all
```
* Update `image` with apiscout docker image name built in the previous step and `EXTERNALIP` value with minikube IP in apiscout.yml 

```bash
# Deploy apiscout to Kubernetes
$ make run-kube
```

3. Build and deploy sample micro service

```bash
# Navigate to samples/invoiceservice-go folder
$ cd samples/invoiceservice-go

# Install dependencies
$ make deps

# Build sample microservice application
$ make build-app

# Build dcoker image with the sample microservice application
$ make build-docker

# Deploy sample application to Kubernetes
$ make run-kube

```

### Testing


```bash
# Navigate back to apiscout directory
$ cd ../..

# Open kubernetes service url in a web browser to see sample application api specification in swagger format
$ make minikube-show

```

### Cleanup

```bash
# Delete sample application from Kubernetes
$ cd samples/invoiceservice-go
$ make clean-kube

# Delete apiscout from Kubernetes
$ cd ../..
$ make clean-kube

# Stop minikube
$ make minikube-stop
```

## License
See the [LICENSE](./LICENSE) file

_The logo made by [Freepik](http://www.freepik.com) from [www.flaticon.com](https://www.flaticon.com/) is licensed by [CC 3.0 BY](http://creativecommons.org/licenses/by/3.0/)_