#!/bin/bash

echo "[GIT] ------------------------------------------------------"

# shellcheck disable=SC2162
read -p "[GIT] enter new tag:" number

echo ""


echo "[GIT ${number}] cherry"
git tag -a "${number}" -m "auto tag"


echo "[GIT ${number}] components/cron"
git tag -a "components/cron/${number}" -m "auto tag"


echo "[GIT ${number}] components/data-config"
git tag -a "components/data-config/${number}" -m "auto tag"


echo "[GIT ${number}] components/etcd"
git tag -a "components/etcd/${number}" -m "auto tag"


echo "[GIT ${number}] components/gin"
git tag -a "components/gin/${number}" -m "auto tag"


echo "[GIT ${number}] components/gops"
git tag -a "components/gops/${number}" -m "auto tag"


echo "[GIT ${number}] components/gorm"
git tag -a "components/gorm/${number}" -m "auto tag"


echo "[GIT ${number}] components/mongo"
git tag -a "components/mongo/${number}" -m "auto tag"

echo "[GIT ${number}] examples"
git tag -a "examples/${number}" -m "auto tag"

echo "[GIT] ------------------------------------------------------"