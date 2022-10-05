# Example: quarkus-devservices

This folder contains an example which is using [Quarkus dev-services](https://quarkus.io/guides/dev-services). To run this locally, make sure kubedock is running with port-forwarding enabled (`kubedock server --port-forward`). 

```bash
export TESTCONTAINERS_RYUK_DISABLED=true
export TESTCONTAINERS_CHECKS_DISABLE=true
export DOCKER_HOST=tcp://127.0.0.1:2475
mvn test
```
