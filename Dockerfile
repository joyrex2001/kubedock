####################
## Build kubedock ## ----------------------------------------------------------
####################

FROM docker.io/golang:1.17 AS kubedock

ARG CODE=github.com/joyrex2001/kubedock

ADD . /go/src/${CODE}/
RUN cd /go/src/${CODE} \
    && make test build \
    && mkdir /app \
    && cp kubedock /app

#################
## Final image ## ------------------------------------------------------------
#################

FROM alpine:3

RUN apk add --no-cache ca-certificates \
    && update-ca-certificates

COPY --from=kubedock /app /app

WORKDIR /app

ENTRYPOINT ["/app/kubedock"]
CMD [ "server" ]
