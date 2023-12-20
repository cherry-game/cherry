#!/bin/bash

echo "[VER] ------------------------------------------------------"

# shellcheck disable=SC2162
read -p "[VER] enter new version(eg. 1.3.12):" number

echo ""

file_path='const/const.go'
sed -i 's/version\s*=\s*"[0-9.]\+"/version = "'"${number}"'"/' ${file_path}

file_path='components/**/go.mod'
sed -i 's/cherry v[0-9.]\+/cherry v'"${number}"'/' ${file_path}

file_path='examples/go.mod'
sed -i 's/cherry v[0-9.]\+/cherry v'"${number}"'/' ${file_path}
sed -i 's/components\/cron v[0-9.]\+/components\/cron v'"${number}"'/' ${file_path}
sed -i 's/components\/data-config v[0-9.]\+/components\/data-config v'"${number}"'/' ${file_path}
sed -i 's/components\/gin v[0-9.]\+/components\/gin v'"${number}"'/' ${file_path}
sed -i 's/components\/gops v[0-9.]\+/components\/gops v'"${number}"'/' ${file_path}
sed -i 's/components\/gorm v[0-9.]\+/components\/gorm v'"${number}"'/' ${file_path}