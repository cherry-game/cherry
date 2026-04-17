#!/bin/bash

CURRENT_VERSION=$(sed -n 's/.*version = "\([^"]*\)".*/\1/p' const/const.go)
echo "[TAG] current version: ${CURRENT_VERSION}"
echo "[TAG] ------------------------------------------------------"

# shellcheck disable=SC2162
read -p "[TAG] enter new tag(current: ${CURRENT_VERSION}):" number
number=${number:-$CURRENT_VERSION}

echo ""

echo "[TAG] create ${number}"
git tag -a "v${number}" -m "auto tag"

echo "[TAG] ------------------------------------------------------"
