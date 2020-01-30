#!/bin/bash

set -e

if [ $# -lt 3 ]; then
    echo "Usage: run_in_docker.sh <path to Dockerfile> <context directory> <image tag> <cache>"
    exit 1
fi

dockerfile=$1
context=$2
destination=$3

cache="false"
if [[ ! -z "$4" ]]; then
    cache=$4
fi

if [[ $destination == *"gcr"* ]]; then
    if [[ ! -e $HOME/.config/gcloud/application_default_credentials.json ]]; then
        echo "Application Default Credentials do not exist. Run [gcloud auth application-default login] to configure them"
        exit 1
    fi
    docker run \
        -v "$HOME"/.config/gcloud:/root/.config/gcloud \
        -v "$context":/workspace \
        gcr.io/kaniko-project/executor:latest \
        --dockerfile "${dockerfile}" --destination "${destination}" --context dir:///workspace/ \
        --cache="${cache}"
else
    docker run \
        -v "$context":/workspace \
        gcr.io/kaniko-project/executor:slowjam \
        --dockerfile "${dockerfile}" --destination "${destination}" --context dir:///workspace/ \
        --cache="${cache}"
fi

cd ../pkg/slowjam/cmd/timeline
go run main.go ../../../../slowjam-test/stack.log
