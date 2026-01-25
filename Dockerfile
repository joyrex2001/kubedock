
FROM docker.io/busybox:latest
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/kubedock /usr/local/bin/kubedock

# Updates ca-certificates for the `inspector` command to connect to dockerhub etc.
RUN update-ca-certificates

ENTRYPOINT ["/usr/local/bin/kubedock"]
CMD [ "server" ]