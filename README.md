# kubedock

Kubedock is an experimental implementation of the docker api that will orchestrate containers into a kubernetes cluster, rather than running containers locally. The main driver for this project is to be able running [testcontainers-java](https://www.testcontainers.org) enabled unit-tests in k8s, without the requirement of running docker-in-docker within resource heavy containers.

The current implementation is limited, but able to run containers that just expose ports, copy resources towards the container, or mount volumes. Containers that 'just' expose ports, require logging and copy resources to running containers will probably work. Volume mounting is implemented by copying the local volume towards the container, changes made by the container to this volume are not synced back. All data is considered emphemeral. If a container has network aliases configured, it will create k8s services with the alias as name. However, if aliases are present, a port mapping should be configured as well. If no port-mapping is given, services can still be created, but require kubedock to run with the `--inspector` argument. This will let kubedock inspect the image by fetching the configuration details from the registry, and use the ports that are listed in the image itself. 

## Quick start

Running this locally with a testcontainers enabled unit-test requires to run kubedock (`make run`). After that start the unit tests in another terminal with the below environment variables set, for example:

```bash
export TESTCONTAINERS_RYUK_DISABLED=true
export DOCKER_HOST=tcp://127.0.0.1:8080
mvn test
```

The default configuration for kubedock is to orchestrate in the default kubernetes namespace, this can be configured with the `NAMESPACE` environment variable (or via the -n argument). The service requires permissions to create Deployments in the namespace.

To see a complete list of available options: `kubedock --help`.

## Compatibility

This project is mainly focussed on getting [testcontainers-java](https://www.testcontainers.org) tests running. Therefor it has limited support of the docker api, only the minimum that is typically used in testcontainers-java backed tests.

Most of the below use-cases as described by testcontainers are working:

* [generic containers](https://www.testcontainers.org/features/creating_container/)
* [networking and communicating with containers](https://www.testcontainers.org/features/networking/)
* [executing commands](https://www.testcontainers.org/features/commands/)
* [waiting for containers to start or be ready](https://www.testcontainers.org/features/startup_and_waits/)
* [files and volumes](https://www.testcontainers.org/features/files/)
* [accessing container logs](https://www.testcontainers.org/features/container_logs/)

The below use-cases are mostly not working:

* [advanced networking](https://www.testcontainers.org/features/networking/)
* [creating images on-the-fly](https://www.testcontainers.org/features/creating_images/)
* [ryuk resource reaper](https://www.testcontainers.org/features/configuration/)
* [advanced options](https://www.testcontainers.org/features/advanced_options/)

## Resource cleanup

Kubedock will dynamically create deployments and services in the configured namespace. If kubedock is requested to delete a container, it will remove the deployment and related services. Kubedock will also delete all the resources (Services and Deployments) it created in the running instance before exiting (identified with the `kubedock.id` label).

### Automatic reaping

If e.g. a test fails and didn't clean up its started containers, these resources will remain in the namespace. To prevent unused deployments and services lingering around, kubedock will automatically delete deployments and services that are older than 15 minutes (default) if it's owned by the current process. If the deployment is not owned by the running process, it will delete it after 30 minutes if the deployment or service has the label `kubedock=true`.

### Forced cleaning

The reaping of resources can also be enforced at startup. When kubedock is started with the `--prune-start` argument, it will delete all resources that have the `kubedock=true` before starting the API server. These resource includes resources created by other instances. 

# See also

* https://hub.docker.com/r/joyrex2001/kubedock