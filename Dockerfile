####################
## Build kubedock ## ----------------------------------------------------------
####################

FROM docker.io/golang:1.16 AS kubedock

ARG CODE=github.com/joyrex2001/kubedock

ADD . /go/src/${CODE}/
RUN cd /go/src/${CODE} \
 && go test ./... \
 && make build \
 && mkdir /app \
 && cp kubedock /app

#################
## Final image ## ------------------------------------------------------------
#################

FROM docker.io/busybox:latest

COPY --from=kubedock /app /app

WORKDIR /app

ENTRYPOINT ["/app/kubedock"]
CMD []
