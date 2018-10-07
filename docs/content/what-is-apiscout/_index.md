---
title: What is API Scout?
weight: 1000
---

# What is API Scout

As we're all building and deploying microservices, there are a few common concerns that every developer and Ops has:

* _“where did I deploy that microservice?”_ 😩
* _“what is the API definition of that microservice again?“_ 😟

When your deployment footprint grows, keeping track of all those deployed microservices on Kubernetes can become quite a challenge. Keeping the API documentation updated for developers, could become even more challenging. API Scout is an attempt to solve that challenge. API Scout helps you get up-to-date API docs to your developers by simply annotating your services in Kubernetes.

# How it works

API Scout catalogs and documents your Kubernetes microservices to, ultimately, productize them as APIs using an API management platform. To do that it:

* Automatically discover microservices with annotations
* Generates beautiful pixel-perfect OAS/Swagger-based API Docs
* Has first-class support for Kubernetes, PKS & OpenShift
* Is 100% Open Source, so free to use & build on