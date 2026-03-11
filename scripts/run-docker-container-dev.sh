#! /bin/sh
IMAGE_NAME=fiagram_account_service
IMAGE_VERSION=$(cat ./VERSION)

docker run --rm -it \
    -v "$PWD/deployments/configs:/home:ro" \
    -p 11000:8080 \
    $IMAGE_NAME:$IMAGE_VERSION \
    /account_service -c /home/test.yaml