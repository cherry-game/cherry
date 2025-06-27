#!/bin/bash

echo "[TAG] ------------------------------------------------------"

# shellcheck disable=SC2162
read -p "[TAG] enter new tag(eg. 1.3.14):" number

echo ""


echo "[TAG ${number}] cherry"
git tag -a "v${number}" -m "auto tag"

echo "[TAG] ------------------------------------------------------"