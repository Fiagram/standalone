#! /bin/sh
IMAGE_NAME=fiagram_standalone
IMAGE_VERSION=$(cat ./VERSION)

docker build --target deployment -t $IMAGE_NAME:$IMAGE_VERSION .