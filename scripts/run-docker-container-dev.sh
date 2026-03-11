#! /bin/sh
IMAGE_NAME=fiagram_standalone
IMAGE_VERSION=$(cat ./VERSION)

docker run --rm -it \
    -v "$PWD/deployments/configs/test.yaml:/home/test.yaml:ro" \
    -p 11000:8080 \
    $IMAGE_NAME:$IMAGE_VERSION \
    /standalone -c /home/test.yaml