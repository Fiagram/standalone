# Refer: https://hub.docker.com/r/apache/kafka
```
docker run -d  -p 9092:9092 apache/kafka:4.2.0

docker exec --workdir /opt/kafka/bin/ -it broker sh

./kafka-topics.sh --bootstrap-server localhost:9092 --create --topic torch

./kafka-console-producer.sh --bootstrap-server localhost:9092 --topic torch

./kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic torch --from-beginning
```