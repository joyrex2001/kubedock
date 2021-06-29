# Example: docker compose

This folder contains an example docker-compose wordpress setup. Note that this is not a typical use-case for kubedock, however, it does demonstrate some of the nuances you might encounter using kubedock. To run this locally, make sure kubedock is running with port-forwarding enabled (`kubedock server --port-forward`).

```bash
docker compose up -d
docker compose ps
curl -v localhost:8000
docker compose rm -f
```

Building images is not supported, as kubedock is not able to do this.