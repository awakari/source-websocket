#!/bin/bash

export SLUG=ghcr.io/awakari/source-websocket
export VERSION=latest
docker tag awakari/source-websocket "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
