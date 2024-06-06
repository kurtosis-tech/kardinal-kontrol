# Azure Voting Application + Redis Proxy Overlay

We added a Redis proxy in between the Azure vote frontend and backend. The proxy supports get, set and incrby commands with an in-memory cache. The missed gets are proxied to the Azure vote backend and their result stores in the in-memory cache.

## Deploying

Build the Redis proxy overlay and load it into Minikube.

```bash
eval $(minikube docker-env)
docker load < $(nix build ./#containers.aarch64-darwin.redis-proxy-overlay.arm64 --no-link --print-out-paths)
```

Deploy the Azure voting app and Redis proxy overlay.

```bash
kubectl apply -n <namespace> -f azure-vote-app.yaml
kubectl port-forward -n <namespace> deployment/azure-vote-front 8080:80
```
