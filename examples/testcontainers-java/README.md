# Example: testcontainers-java

This folder contains an example using the test-containers framework. To run this locally, make sure kubedock is running with port-forwarding enabled (`kubedock server --port-forward`). 

```bash
export TESTCONTAINERS_RYUK_DISABLED=true
export TESTCONTAINERS_CHECKS_DISABLE=true
export DOCKER_HOST=tcp://127.0.0.1:2475
mvn test
```

The example includes:

* `NginxTest.java` which demonstrates how to use volumes in combination with kubedock.
* `NetworkAliasesTest.java` which demonstrates how to use network aliases in combination with kubedock.

When kubedock is running in a cluster, it can return the actual cluster IPs of the services. However, the testcontainers-java framework will not use the actual cluster IP and returns the IP of the docker API instead (see [this](https://github.com/testcontainers/testcontainers-java/issues/452) issue). Therefore you either need to reverse-proxy or port-forward from inside the cluster as well. In this case you might want to use `--reverse-proxy`, which will enable local reverse proxies to the cluster IPs of the services and is more stable as `--port-forward`.