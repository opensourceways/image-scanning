#!/bin/bash
set -e

function init() {
  trivy_resource_dir=$1
  trivy=$2
  trivy_db=$3
  vuln_list=$4

  mkdir -p ${trivy_resource_dir} && cd ${trivy_resource_dir}

  git clone --depth=1 $vuln_list
  git clone --depth=1 $trivy_db
  git clone --depth=1 $trivy

  cd trivy-db
  go build -o trivy-db cmd/trivy-db/main.go
  build_db

  cd ../trivy
  go build -o trivy cmd/trivy/main.go
}

function build_db() {
  ./trivy-db build --cache-dir ../ --only-update openeuler --output-dir ../db/
  ./trivy-db build --cache-dir ../ --only-update ubuntu --output-dir ../db/
}

function update() {
  trivy_resource_dir=$1
  cd ${trivy_resource_dir}/vuln-list
  git pull
  cd ../trivy-db
  build_db
}

case $1 in
  init)
    init $2 $3 $4 $5
    ;;
  update)
    update $2
    ;;
  *)
    echo "Usage: $0 {init|update}"
    exit 1
    ;;
esac



