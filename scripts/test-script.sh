#!/usr/bin/env bash

set -e

deployment=${1:-${DEPLOYMENT_NAME}}
namespace=${2:-${NAMESPACE}}

[[ -z $namespace ]] && namespace="default"

path=$(dirname "$0")

timed() {
  end=$(date +%s)
  dt=$(($end - $1))
  dd=$(($dt / 86400))
  dt2=$(($dt - 86400 * $dd))
  dh=$(($dt2 / 3600))
  dt3=$(($dt2 - 3600 * $dh))
  dm=$(($dt3 / 60))
  ds=$(($dt3 - 60 * $dm))

  LC_NUMERIC=C printf "\nTotal runtime: %02d min %02d seconds\n" "$dm" "$ds"
}

success() {
  newman run \
    --delay-request=100 \
    --folder=success \
    --export-environment postman/environment.json \
    --environment postman/environment.json \
    postman/collection.json
}

step() {
  local step=$1
  [[ $((step % 2)) -eq 0 ]] && replicas=1 || replicas=0

  printf "=== Step %d: scale %s to %s ===\n" "$step" "$deployment" "$replicas"

  kubectl scale deployment "$deployment" -n "$namespace" --replicas "$replicas" 

  newman run \
    --delay-request=100 \
    --folder=step"$step" \
    --export-environment postman/environment.json \
    --environment postman/environment.json \
    postman/collection.json

  printf "=== Step %d completed ===\n" "$step"
}

start=$(date +%s)
trap 'timed $start' EXIT

printf "=== Start test scenario ===\n"

# success execute
success

# stop deployment
step 1

# start deployment
step 2

# stop deployment
step 3

# start deployment
step 4
