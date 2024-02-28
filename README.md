# Kubedock

Kubedock is a minimal implementation of the docker api that will orchestrate containers on a kubernetes cluster, rather than running containers locally. The main driver for this project is to run tests that require docker-containers inside a container, without the requirement of running docker-in-docker within resource heavy containers. Containers that are orchestrated by kubedock are considered short-lived and ephemeral and not intended to run production services. An example use case is running [testcontainers-java](https://www.testcontainers.org) enabled unit-tests in a tekton pipeline. In this use case, running kubedock in a sidecar can help orchestrating containers inside the kubernetes cluster instead of within the task container itself.

## Quick start

Running this locally with a testcontainers enabled unit-test requires to run kubedock with port-forwarding enabled (`kubedock server --port-forward`). After that start the unit tests in another terminal with the below environment variables set, for example:

```bash
export TESTCONTAINERS_RYUK_DISABLED=true  ## optional, can be enabled
export TESTCONTAINERS_CHECKS_DISABLE=true ## optional, can be enabled
export DOCKER_HOST=tcp://127.0.0.1:2475
mvn test
```

The default configuration for kubedock is to orchestrate in the namespace that has been set in the current context. This can be overruled with -n argument (or via the `NAMESPACE` environment variable). The service requires permissions to create pods, services and configmaps. If namespace locking is used, the service also requires permissions to create leases in the namespace.

To see a complete list of available options: `kubedock --help`.

# Implementation

When kubedock is started with `kubedock server` it will start an API server on port :2475, which can be used as a drop-in replacement for the default docker api server. Additionally, kubedock can also start listening to an unix-socket (`docker.sock`).

## Containers

Container API calls are translated towards kubernetes pods. When a container is started, it will create a kubernetes service within the cluster and maps the ports to that of the container (note that only tcp is supported). This will make it accessible for use within the cluster (e.g. within a containerized pipeline within that same cluster). It is also possible to create port-forwards for the ports that should be exposed with the `--port-forward` argument. These are however not very performant, nor stable and are intended for local debugging. If the ports should be exposed on localhost as well, but port-forwarding is not required, they can be made available via the built-in reverse-proxy. This can be enabled with the `--reverse-proxy` argument and is mutually exclusive with `--port-forward`.

Starting a container is a blocking call that will wait until it results in a running pod. By default it will wait for maximum 1 minute, but this is configurable with the `--timeout` argument. The logs API calls will always return the complete history of logs, and doesn't differentiate between stdout/stderr. All log output is send as stdout. Executions in the containers are supported.

By default, all containers will be orchestrated using kubernetes pods. If a container has been given a specific name, this will be visible in the name of the pod. If the label `com.joyrex2001.kubedock.name-prefix` has been set, this will be added as a prefix to the name.

The containers will be started with the `default` service account. This can be changed with the `--service-account`. If required, the uid of the user that runs inside the container can also be enforced with the `--runas-user` argument and the `com.joyrex2001.kubedock.runas-user` label.

## Volumes

Volumes are implemented by copying the source content to the container by means of an init-container that is started before the actual container is started. By default the kubedock image with the same version as the running kubedock is used as the init container. However, this can be any image that has tar available and can be configured with the `--initimage` argument.

Volumes are one-way copies and ephemeral. This typically means, any data that is written into the volume is not available locally. This also means that mounts to devices, or sockets are not supported (e.g. mounting a docker-socket). Volumes that point to a single file will be converted to a configmap (and is implicitly read-only always).

Copying data from a running container back to the client is supported as well, but only works if the running container has tar available. Also be aware that copying data to a container will implicitly start the container. This is different compared to a real docker api, where a container can be in an unstarted state. To 'workaround' this, use a volume instead. Alternatively kubedock can be started with `--pre-archive`, which will convert copy statements of single files to configmaps when the container is started yet. This will implicitly make the target file read-only, and may not work in all use-cases (hence it's not the default).

## Networking

Kubedock flattens all networking, which basically means that everything will run in the same namespace. This should be sufficient for most use-cases. Network aliases are supported. When a network alias is present, it will create a service exposing all ports that have been exposed by the container. If no ports are configured, kubedock is able to fetch ports that are exposed in the container image. To do this, kubedock should be started with the `--inspector` argument.

## Images

Kubedock implements the images API by tracking which images are requested. It is not able to actually build or import images. If kubedock is started with `--inspector`, kubedock will fetch configuration information about the image by calling external container registries. This configuration includes ports that are exposed by the container image itself, and increases network aliases support. The registries should be configured by the client (for example by doing a `skopeo login`). By default images that are used are deployed with a 'IfNotPresent' pull policy. This can be globally configured with the `--pull-policy` argument, and can be configured on container level by adding a label `com.joyrex2001.kubedock.pull-policy` to the container. Possible values are 'never', 'always' and 'ifnotpresent'.

## Namespace locking

If multiple kubedocks are using the namespace, it might be possible there will be collisions in network aliases. Since networks are flattened (see Networking), all network aliases will result in a Service with the name of the given network alias. To ensure tests don't fail because of these name collisions, kubedock can lock the namespace while it's running. When enabling this with the `--lock` argument, kubedock will create a lease called `kubedock-lock` in the namespace in which it tracks the current ownership.

## Resource requests and limits

By default containers are started without any resource request configuration. This can impact performance of the tests that are run in the containers. Setting resource requests (and limits) will allow better scheduling, and can improve the overall performance of the running containers. Global requests and limits can be set with `--request-cpu` and `--request-memory`, which takes regular kubernetes resource requests configurations as can be found in the [kubernetes documentation](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/). Limits are optional, and can be configured by adding it with a ,limit. If the values should be configured specifically for a container, they can be configured by adding `com.joyrex2001.kubedock.request-cpu` or `com.joyrex2001.kubedock.request-memory` labels to the container with their specific requests (and limits). The labels take precedence over the cli configuration.

## Active deadline seconds

Sometimes you may want to specify an `activeDeadlineSeconds` for the pods run by Kubedock; this is useful in multi-tenant environments if you want the pods to use resources in the `terminating` quota (if `activeDeadlineSeconds` is not set, pods will use `notterminating` quota). You can set the default value using `--active-deadline-seconds`; pod-specific values can be configured by adding `com.joyrex2001.kubedock.active-deadline-seconds` label.

## Pod template

The pods that are created by kubedock can be customized with additional configuration by providing a pod template with `--pod-template`. If this is provided, all pods that are created by kubedock will use the provided pod template as a base. If the template contains a containers definition, it will use the first entry in the list as a template for all containers kubedock adds to a pod (including sidecars and init containers). Note that volumes are ignored in these templates. Settings configured via the pod-template have the least precedence in case these can also be configured via other means (cli or labels).

## Kubernetes labels and annotations

Labels that are added to container images are added as annotations and labels to the created kubernetes pods. Additional labels and annotations can be added with the `--annotation` and `--label` cli argument. Environment variables that start with `K8S_ANNOTATION_` and `K8S_LABEL_` will be added as a kubernetes annotation or label as well. For example `K8S_ANNOTATION_FOO` will create an annotation `foo` with the value of the environment variable. Note that annotations and labels added via environment variables or cli will not be processed by kubedock if they have a specific control function. For these occasions specific environment variables and cli arguments are present.

## Resources cleanup

Kubedock will dynamically create pods and services in the configured namespace. If kubedock is requested to delete a container, it will remove the pod and related services. Kubedock will also delete all the resources (services and pods) it created in the running instance before exiting (identified with the `kubedock.id` label).

### Automatic reaping

If a test fails and didn't clean up its started containers, these resources will remain in the namespace. To prevent unused pods, configmaps and services lingering around, kubedock will automatically delete these resources. If these resources are owned by the current process, they will be removed if they are older than 60 minutes (default). If the resources have the label `kubedock=true`, but are not owned by the running process, it will delete them 15 minutes after the initial reap interval (in the default scenario; after 75 minutes).

### Forced cleaning

The reaping of resources can also be enforced at startup. When kubedock is started with the `--prune-start` argument, it will delete all resources that have the label `kubedock=true`, before starting the API server. This includes resources that are created by other instances of kubedock.

## Docker-in-docker support

Kubedock detects if a docker-socket is bound, and will add a kubedock-sidecar providing this docker-socket to support docker-in-docker use-cases. The sidecar that will be deployed for these containers, will proxy all api calls to the main kubedock.

## Service Account RBAC

As a reference, the below role can be used to manage the permissions of the service account that is used to run kubedock in a cluster. The uncommented rules are the minimal permissions. Depending on use of `--lock`, the additional (commented) rule is required as well.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: kubedock
  namespace: cicd
rules:
  - apiGroups: [""]
    resources: ["pods"]
    verbs: ["create", "get", "list", "delete", "watch"]
  - apiGroups: [""]
    resources: ["pods/log"]
    verbs: ["list", "get"]
  - apiGroups: [""]
    resources: ["pods/exec"]
    verbs: ["create"]
  - apiGroups: [""]
    resources: ["services"]
    verbs: ["create", "get", "list", "delete"]
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["create", "get", "list", "delete"]
## optional permissions (depending on kubedock use)
# - apiGroups: ["coordination.k8s.io"]
#   resources: ["leases"]
#   verbs: ["create", "get", "update"]
```

# See also

* https://github.com/joyrex2001/kubedock
* https://hub.docker.com/r/joyrex2001/kubedock
