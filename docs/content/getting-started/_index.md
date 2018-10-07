---
title: Getting Started
weight: 2000
---

# Workflow

To use API Scout with any of your microservices, existing or new, you'll need to follow three easy steps

# Step 1: Build a microservice

The first step is to build your microservice or identify an existing microservice. The microservice that you want to index with API Scout needs to have an endpoint available that returns the OpenAPI specification

![picture1](../images/getting-started/Picture1.png)

# Step 2: Annotate it

The second step is to annotate your K8s service definition. The Kubernetes deployment file needs two specific annotations for API Scout to work:

* `apiscout/index: 'true'`: This annotation ensures that apiscout indexes the service
* `apiscout/swaggerUrl: '/swaggerspec'`: This is the URL from where apiscout will read the OpenAPI document

![picture2](../images/getting-started/Picture2.png)

# Step 3: Access the portal

The third step is easy! Access your microservices documentation through the portal...

![picture3](../images/getting-started/Picture3.png)

# In the background

While you're deploying microservices to Kubernetes, API Scout takes care of a few things. API Scout registers a watcher for Service event updates. Once an event comes in, API Scout will update its registry to reflect the change in the API.

Let's say that a developer is working on an update to a microservice. Once the developer is done and deploys the changes, the developer also makes sure the service definition has the annotations for API Scout included.

The Kubernetes API Server receives the updates made to the service by the the developer and dispatches an event to all watchers. API Scout, being one of those watchers updates its registry to reflect the change in the API.

The developer goes to the documentation portal and sees the updated API documentation.