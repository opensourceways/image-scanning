#!/bin/bash
set -e

trivy_resource_dir=trivy_resource

function init() {
  trivy=$1
  trivy_db=$2
  vuln_list=$3

  mkdir -p ${trivy_resource_dir} && cd ${trivy_resource_dir}

  git clone --depth=1 $vuln_list
  git clone --depth=1 $trivy_db
  git clone --depth=1 $trivy

  cd trivy-db
  go build -o trivy-db cmd/trivy-db/main.go
  ./trivy-db build --cache-dir ../ --only-update openeuler --output-dir ../db/

  cd ../trivy
  go build -o trivy cmd/trivy/main.go
  cp trivy ../../
}

function update() {
  cd ${trivy_resource_dir}/vuln-list
  git pull
  cd ../trivy-db
  ./trivy-db build --cache-dir ../ --only-update openeuler --output-dir ../db/
}

case $1 in
  init)
    init $2 $3 $4
    ;;
  update)
    update
    ;;
  *)
    echo "Usage: $0 {init|update}"
    exit 1
    ;;
esac



