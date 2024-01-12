#!/bin/bash

echo "[TAG] ------------------------------------------------------"

# shellcheck disable=SC2162
read -p "[TAG] enter new tag(eg. 1.3.12):" number

echo ""


echo "[TAG ${number}] cherry"
git tag -a "${number}" -m "auto tag"


echo "[TAG ${number}] components/cron"
git tag -a "components/cron/v${number}" -m "auto tag"


echo "[TAG ${number}] components/data-config"
git tag -a "components/data-config/v${number}" -m "auto tag"


echo "[TAG ${number}] components/etcd"
git tag -a "components/etcd/v${number}" -m "auto tag"


echo "[TAG ${number}] components/gin"
git tag -a "components/gin/v${number}" -m "auto tag"


echo "[TAG ${number}] components/gops"
git tag -a "components/gops/v${number}" -m "auto tag"


echo "[TAG ${number}] components/gorm"
git tag -a "components/gorm/v${number}" -m "auto tag"


echo "[TAG ${number}] components/mongo"
git tag -a "components/mongo/v${number}" -m "auto tag"

echo "[TAG ${number}] examples"
git tag -a "examples/v${number}" -m "auto tag"

echo "[TAG] ------------------------------------------------------"