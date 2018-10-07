---
title: Dependencies
weight: 6000
---

## What are the components of API Scout

The docker image that is deployed to Kubernetes has several components:

* The container itself is based on [nginx:1.15-alpine](https://hub.docker.com/_/nginx/)
* The webapp is a staticly generated site by [Hugo](https://github.com/gohugoio/hugo) using the [Learn](https://themes.gohugo.io/hugo-theme-learn/) theme and [Swagger UI](https://github.com/swagger-api/swagger-ui/releases)
* A server app that connects to the Kubernetes cluster using a default role to watch for services that need to be indexed

_Hugo is downloaded and embedded during the build of the container_

## Logo

The logo made by [Freepik](http://www.freepik.com) from [www.flaticon.com](https://www.flaticon.com/) is licensed by [CC 3.0 BY](http://creativecommons.org/licenses/by/3.0/)