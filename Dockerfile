# We use multistage build, to keep the runtime binary as small as possible, but still be able to add ca-certs to the
# runtime. We need the certs for the `inspector` command to connect to dockerhub etc.
FROM docker.io/alpine:latest AS certs
RUN apk --update add ca-certificates

FROM docker.io/busybox:latest
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/kubedock /usr/local/bin/kubedock

COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/usr/local/bin/kubedock"]
CMD [ "server" ]