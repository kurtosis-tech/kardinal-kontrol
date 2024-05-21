# Kardinal Progressive Delivery Demo

This repo contains the [Argo Rollouts](https://github.com/argoproj/argo-rollouts) demo application source code and examples. It demonstrates the
various deployment strategies and progressive delivery features of Argo Rollouts.

Before running an example:

1. Enter the dev shell:

```bash
nix develop
minikube start --driver=docker
```

2. Setup Argo Rollouts (first time)

```bash
kubectl create namespace argo-rollouts
kubectl apply -n argo-rollouts -f https://github.com/argoproj/argo-rollouts/releases/latest/download/install.yaml
```

```bash
kubectl apply -f src/bluegreen-rollout.yaml
kubectl apply -f src/service.yaml
```

```bash
minikube service bluegreen-demo --url
minikube service bluegreen-demo-preview --url
```

3. Watch the rollout or experiment using the argo rollouts kubectl plugin:

```bash
kubectl argo rollouts get rollout bluegreen-demo --watch
```

4. For rollouts, trigger an update by setting the image of a new color to run:

```bash
kubectl argo rollouts set image bluegreen-demo "*=argoproj/rollouts-demo:yellow"
```

```bash
kubectl argo rollouts dashboard
```
