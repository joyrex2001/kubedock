# Configuration reference

The kubedock binary has the following commands available:
* `server` Start the kubedock api server
* `dind` Start the kubedock docker-in-docker proxy
* `readme` Display project readme
* `version`  Display kubedock version details

The `server` command is the actual kubedock server, and is the relevant command to be used to start kubedock. The table below shows all possible commands and possible arguments. Some commands are also configureable via environment variables, as shown in the environment variable column.

|command|argument|default|environment variable|description|
|---|---|---|---|---|
|server|--listen-addr|:2475|SERVER_LISTEN_ADDR|Webserver listen address|
|server|--unix-socket|||Unix socket to listen to (instead of port)|
|server|--tls-enable|false|SERVER_TLS_ENABLE|Enable TLS on api server|
|server|--tls-key-file||SERVER_TLS_CERT_FILE|TLS keyfile|
|server|--tls-cert-file||SERVER_TLS_CERT_FILE|TLS certificate file|
|server|--namespace / -n|<current namespace>|NAMESPACE|Namespace in which containers should be orchestrated|
|server|--initimage|joyrex2001/kubedock:version|INIT_IMAGE|Image to use as initcontainer for volume setup|
|server|--dindimage|joyrex2001/kubedock:version|DIND_IMAGE|Image to use as sidecar container for docker-in-docker support|
|server|--pull-policy|ifnotpresent|PULL_POLICY|Pull policy that should be applied (ifnotpresent,never,always)|
|server|--service-account|default|SERVICE_ACCOUNT|Service account that should be used for deployed pods|
|server|--image-pull-secrets||IMAGE_PULL_SECRETS|Comma separated list of image pull secrets that should be used|
|server|--pod-template||POD_TEMPLATE|Pod file that should be used as the base for creating pods|
|server|--inspector / -i|false||Enable image inspect to fetch container port config from a registry|
|server|--timeout / -t|1m|TIME_OUT|Container creating/deletion timeout|
|server|--reapmax / -r|60m|REAPER_REAPMAX|Reap all resources older than this time|
|server|--request-cpu||K8S_REQUEST_CPU|Default k8s cpu resource request (optionally add ,limit)|
|server|--request-memory||K8S_REQUEST_MEMORY|Default k8s memory resource request (optionally add ,limit)|
|server|--runas-user||K8S_RUNAS_USER|Numeric UID to run pods as (defaults to UID in image)|
|server|--lock|false||Lock namespace for this instance|
|server|--lock-timeout|15m||Max time trying to acquire namespace lock|
|server|--verbosity / -v|1|VERBOSITY|Log verbosity level|
|server|--prune-start / -P|false||Prune all existing kubedock resources before starting|
|server|--port-forward|false||Open port-forwards for all services|
|server|--reverse-proxy|false||Reverse proxy all services via 0.0.0.0 on the kubedock host as well|
|server|--pre-archive|false||Enable support for copying single files to containers without starting them|
|server|--annotation||K8S_ANNOTATION_annotation|annotation that need to be added to every k8s resource (key=value)|
|server|--label||K8S_LABEL_label|label that need to be added to every k8s resource (key=value)|
|dind|--unix-socket|/var/run/docker.sock||Unix socket to listen to|
|dind|--kubedock-url|||Kubedock url to proxy requests to|
|dind|--verbosity / -v|1|VERBOSITY|Log verbosity level|
|readme||||Display project readme|
|readme|config|||Display configuration reference|
|readme|licence|||Display project licence|
|version||||Display kubedock version details|

## Labels and annotations

Labels that are added to container images are added as annotations and labels to the created kubernetes pods. Additional labels and annotations can be added with the `--annotation` and `--label` cli argument. Environment variables that start with `K8S_ANNOTATION_` and `K8S_LABEL_` will be added as a kubernetes annotation or label as well. For example `K8S_ANNOTATION_FOO` will create an annotation `foo` with the value of the environment variable. Note that annotations and labels added via environment variables or cli will not be processed by kubedock if they have a specific control function. For these occasions specific environment variables and cli arguments are present.