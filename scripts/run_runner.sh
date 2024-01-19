#!/usr/bin/env bash

set -euo pipefail

result_file=/data/${POD_NAMESPACE}-${POD_NAME}-${POD_UID}.json

/kperf runner run --config=/config/load_profile.yaml \
		--user-agent=${POD_NAME} \
    --result=${result_file} \
    --raw-data

# TODO(weifu): retry if it's not 409
set +e
curl -v -H "Content-Type: application/json" -XPOST -d "@${result_file}" ${TARGET_URL}
set -e
