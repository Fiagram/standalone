#! /bin/sh
docker run \
    -e MYSQL_ROOT_PASSWORD=root \
    -e MYSQL_DATABASE=fiagram \
    -p 3306:3306 \
    -d mysql:9.5.0 \
    --character-set-server=utf8mb4 \
    --collation-server=utf8mb4_unicode_ci