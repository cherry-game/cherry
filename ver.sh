#!/bin/bash

echo "[VER] ------------------------------------------------------"

# shellcheck disable=SC2162
read -p "[VER] enter new version(eg. 1.3.14):" number

if [[ "$OSTYPE" == "linux"* ]] || [[ "$OSTYPE" == "msys"* ]] ; then
    echo "[VER] use sed"
    sed -i 's/version = "[0-9.]*"/version = "'"${number}"'"/' const/const.go
elif [[ "$OSTYPE" == "darwin"* ]]; then
    echo "[VER] use gsed"
    gsed -i 's/version = "[0-9.]*"/version = "'"${number}"'"/' const/const.go
fi