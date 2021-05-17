# kubedock

Kubedock is an experimental implementation of the docker api that will orchestrate containers into a kubernetes cluster, rather than running containers locally. The main driver for this project is to be able running [testcontainers-java](https://www.testcontainers.org) enabled unit-tests in k8s, without the requirement of running docker-in-docker within resource heavy containers.

The current implementation is limited, but able to run containers that just expose ports, copy resources towards the container, or mount volumes. Containers that 'just' expose ports, require logging and copy resources to running containers will probably work. Volume mounting is implemented by copying the local volume towards the container, changes made by the container to this volume are not synced back. All data is considered emphemeral.

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

## Resource reaping

Kubedock will dynamically create deployments in the configured namespace. If kubedock is requested to delete a container, it will remove the deployment. However, if e.g. a test fails and didn't clean up its started containers, deployments will remain in the namespace. To prevent unused deployments lingering around, kubedock will automatically delete deployments that are older than 5 minutes (default) if it's owned by the current process. If the deployment is not owned by the running process, it will delete it after 10 minutes if the deployment has the label `kubedock=true`.

# See also

* https://hub.docker.com/r/joyrex2001/kubedock