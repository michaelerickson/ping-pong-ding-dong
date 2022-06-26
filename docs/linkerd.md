# LINKERD

Notes on installing and using [LINKERD](https://linkerd.io/2.11/overview/) as a
service mesh. I'm trying LINKERD instead of Istio as it appears that it might
be a little more focused and easier to set up.

Okay, `linkerd` is super easy to set up and I'm very impressed. Just follow
the instructions here: 

1. https://linkerd.io/2.11/getting-started/
2. https://linkerd.io/2.11/debugging-an-app/
3. https://linkerd.io/2.11/tasks/debugging-your-service/
4. https://linkerd.io/2.11/tasks/adding-your-service/

To automatically mesh a deployment or pod, add an annotation to it as below:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myApp
spec:
  replicas: 1
  selector:
    matchLabels:
      app: myApp
  # ...
  template:
    metadata:
      annotations:
        linkerd.io/inject: enabled # <--- This is the annotation needed
      labels:
        app: myApp
    spec:
      containers:
        - name: ghcr.io/foo/bar
    # ...
```

Run your deployments and launch the dashboard using:

```shell
linkerd viz dashboard &
```
