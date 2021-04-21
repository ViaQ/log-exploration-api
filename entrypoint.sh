#!/bin/sh
set -eou pipefail

: ${ES_ADDR:="https://localhost:9200"}
: ${ES_CERT:="admin-cert"}
: ${ES_KEY:="admin-key"}
: ${ES_TLS:= "false"}

if [ "$1" = "log-exploration-api" ]; then
	exec log-exploration-api \
		-es-addr=${ES_ADDR} \
		-es-cert=${ES_CERT} \
		-es-key=${ES_KEY} \
		-es-tls=${ES_TLS}
fi

exec "$@"

