####################
## Build kubedock ## ----------------------------------------------------------
####################

FROM docker.io/golang:1.25 AS build

ARG CODE=github.com/joyrex2001/kubedock

WORKDIR /go/src/${CODE}

# Updates ca-certificates for the `inspector` command to connect to dockerhub etc.
RUN update-ca-certificates

# Create a cached layer for the dependencies
COPY go.* .
RUN go mod download

# Copy all other files
COPY . .
RUN make test build \
    && mkdir /app \
    && cp kubedock /app/

#################
## Final image ## ------------------------------------------------------------
#################

FROM docker.io/busybox:latest
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the compiled binary from the builder stage
COPY --from=build /app/kubedock /usr/local/bin/kubedock

ENTRYPOINT ["/usr/local/bin/kubedock"]
CMD [ "server" ]
