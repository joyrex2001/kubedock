# Example: Tekton

This folder contains an example tekton task (and a pipeline using this task) that will use kubedock to run the tests of the testcontainers-java example.

Apply the resources: 

```bash
kustomize build . | kubectl apply -f -
```
Start a pipelinerun via cmd:

```bash
tkn pipeline start kubedock-example
        -p git-url=https://github.com/joyrex2001/kubedock.git \
        -p context-dir=examples/testcontainers-java \
        -p git-revision=master
```

Or start a pipelinerun via the provided yaml-file:

```bash
kubectl create -f ./resources/example/pplr_kubedock.yaml
```

The task is using a sidecar container in which kubedock is running. Note that this sidecar container is also mounting the workspace volume. This is required when volumemounts or file copies are used in the tests. If the sidecar is not able to access the workspace, kubedock will not be able to access these files.