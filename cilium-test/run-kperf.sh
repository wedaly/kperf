#!/usr/bin/env bash

set -xe

# KPERF_IMAGE="telescope.azurecr.io/oss/kperf:v0.1.6"
# Need a custom kperf image to get https://github.com/Azure/kperf/pull/145
KPERF_IMAGE="widalytest.azurecr.io/kperf:widalytest004"

../bin/kperf rg del || true
../bin/kperf rg run --runner-image=$KPERF_IMAGE --runnergroup="file://$(pwd)/../contrib/internal/manifests/loadprofile/warmup.yaml"
../bin/kperf rg result || true

../bin/kperf rg del || true
../bin/kperf rg run --runner-image=$KPERF_IMAGE --runnergroup="file://$(pwd)/runnergroup-spec.yaml"
../bin/kperf rg result || true
