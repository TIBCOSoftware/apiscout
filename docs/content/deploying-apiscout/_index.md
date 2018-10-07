---
title: Deploying API Scout
weight: 3000
---

# Deploy to Kubernetes

You can deploy API Scout to Kubernetes by following three easy steps

## Step 1: Create an RBAC role

Assuming you want to run API Scout inside your Kubernetes cluster, which is the recommended option, you'll need to create an _RBAC role_ so that the ServiceAccount has view access to the Kubernetes API server. This is the least privileged option for API Scout.

```yaml
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: default-view
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view
subjects:
  - kind: ServiceAccount
    name: default
    namespace: default
```

## Step 2: Create a deployment

The second step is to create a _deployment_, instructing Kubernetes to deploy API Scout. Using the template yaml file below, there are a few parameters you can update:

* **image**: The name of the docker image to deploy
* **EXTERNALIP**: The value decides if the basePath of an API specification will be overwritten with this value to enable the "try it out" option
* **SWAGGERSTORE**: The location where to store the swaggerdocs
* **HUGOSTORE**: The location where to store content for Hugo
* **MODE**: The mode in which apiscout is running (can be either KUBE or LOCAL)
* **HUGODIR**: The base directory for Hugo

```yaml
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    run: apiscout
  name: apiscout
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      run: apiscout
  template:
    metadata:
      labels:
        run: apiscout
    spec:
      containers:
      - name: apiscout
        image: <your image>
        env:
        - name: MODE
          value: "KUBE"
        - name: HUGODIR
          value: "/tmp"
        - name: EXTERNALIP
          value: "192.168.99.100"
        imagePullPolicy: Never
        ports:
        - containerPort: 80
```

## Step 3: Create a service

To allow access to the documentation portal, the final step is to create a _service_. The below template will instruct Kubernetes to make API Scout available to the outside world on port 8181.

```yaml
apiVersion: v1
kind: Service
metadata:
  labels:
    run: apiscout-svc
  name: apiscout-svc
  namespace: default
spec:
  ports:
  - port: 8181
    protocol: TCP
    targetPort: 80
  selector:
    run: apiscout
  type: LoadBalancer
```

## Complete yaml

With YAML you can combine the above steps into a single document. If you prefer that, the complete document will look like:

```yaml
---
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: default-view
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: view
subjects:
  - kind: ServiceAccount
    name: default
    namespace: default
---
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  labels:
    run: apiscout
  name: apiscout
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      run: apiscout
  template:
    metadata:
      labels:
        run: apiscout
    spec:
      containers:
      - name: apiscout
        image: retgits/apiscout:latest
        env:
        - name: MODE
          value: "KUBE"
        - name: HUGODIR
          value: "/tmp"
        - name: EXTERNALIP
          value: "192.168.99.100"
        imagePullPolicy: Never
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  labels:
    run: apiscout-svc
  name: apiscout-svc
  namespace: default
spec:
  ports:
  - port: 8181
    protocol: TCP
    targetPort: 80
  selector:
    run: apiscout
  type: LoadBalancer
```