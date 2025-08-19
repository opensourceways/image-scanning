#!/bin/bash
set -e

arch=$1
image_path=$2
local_image_path=$3

docker pull --platform $arch $image_path

docker save -o $local_image_path $image_path

docker rmi $image_path