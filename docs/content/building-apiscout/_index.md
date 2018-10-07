---
title: Building API Scout
weight: 4000
---

# Building API Scout

API Scout has a _Makefile_ that can be used for most of the operations. Make sure you have installed Go Programming Language, set GOPATH variable and added $GOPATH/bin in your PATH

```bash
usage: make [target]
```

# Build targets

| Target           | Description                                                           |
|------------------|-----------------------------------------------------------------------|
| deps             | Get dependencies to build the server                                  |
| build-all        | Performs clean-all and executes all build targets                     |
| build-docker     | Builds a docker image from the dist directory                         |
| build-server     | Builds the server app in dist                                         |
| build-site       | Builds the Hugo distribution in dist                                  |

# Clean targets

| Target           | Description                                                           |
|------------------|-----------------------------------------------------------------------|
| clean-all        | Removes the dist directory                                            |
| clean-docker     | Stops and removes all containers and images for apiscout              |
| clean-kube       | Removes the apiscout service and deployment from Kubernetes           |

# Minikube targets

| Target           | Description                                                           |
|------------------|-----------------------------------------------------------------------|
| minikube-delete  | Delete the Minikube installation                                      |
| minikube-install | Install Minikube on this machine                                      |
| minikube-show    | Show the API Scout UI that is deployed to Minikube                    |
| minikube-start   | Start Minikube with default configuration                             |
| minikube-stop    | Stop Minikube                                                         |

# Docker targets

| Target           | Description                                                           |
|------------------|-----------------------------------------------------------------------|
| run-docker       | Runs a docker container with default settings                         |
| run-docs         | Runs the embedded Hugo server on port 1313 for the documentation      |
| run-hugo         | Runs the embedded Hugo server on port 1313                            |
| run-kube         | Deploys apiscout to Kubernetes                                        |
| run-server       | Builds the  in the server directory and runs it with default settings |
| stop-docker      | Stop and remove the running apiscout container                        |