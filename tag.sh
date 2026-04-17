#!/bin/bash

CURRENT_VERSION=$(sed -n 's/.*version = "\([^"]*\)".*/\1/p' const/const.go)
echo "[TAG] current version: ${CURRENT_VERSION}"
echo "[TAG] ------------------------------------------------------"

PREFIX=$(echo "$CURRENT_VERSION" | cut -d'.' -f1,2)
LAST=$(echo "$CURRENT_VERSION" | cut -d'.' -f3)
DEFAULT_VERSION="${PREFIX}.$((LAST + 1))"

# shellcheck disable=SC2162
read -p "[TAG] enter new tag(default: ${DEFAULT_VERSION}):" number
number=${number:-$DEFAULT_VERSION}

echo ""

echo "[TAG ${number}] cherry"
git tag -a "v${number}" -m "auto tag"

echo "[TAG] ------------------------------------------------------"
