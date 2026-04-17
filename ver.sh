#!/bin/bash

CURRENT_VERSION=$(sed -n 's/.*version = "\([^"]*\)".*/\1/p' const/const.go)
echo "[VER] current version: ${CURRENT_VERSION}"
echo "[VER] ------------------------------------------------------"

PREFIX=$(echo "$CURRENT_VERSION" | cut -d'.' -f1,2)
LAST=$(echo "$CURRENT_VERSION" | cut -d'.' -f3)
DEFAULT_VERSION="${PREFIX}.$((LAST + 1))"

# shellcheck disable=SC2162
read -p "[VER] enter new version(default: ${DEFAULT_VERSION}):" number
number=${number:-$DEFAULT_VERSION}

if [[ "$OSTYPE" == "linux"* ]] || [[ "$OSTYPE" == "msys"* ]]; then
	echo "[VER] use sed"
	sed -i 's/version = "[0-9.]*"/version = "'"${number}"'"/' const/const.go
elif [[ "$OSTYPE" == "darwin"* ]]; then
	echo "[VER] use gsed"
	gsed -i 's/version = "[0-9.]*"/version = "'"${number}"'"/' const/const.go
fi
