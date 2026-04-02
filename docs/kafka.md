# Refer: https://hub.docker.com/r/apache/kafka
```
docker run -d  -p 9092:9092 apache/kafka:4.2.0

docker exec --workdir /opt/kafka/bin/ -it broker sh

./kafka-topics.sh --bootstrap-server localhost:9092 --create --topic torch

./kafka-console-producer.sh --bootstrap-server localhost:9092 --topic torch

./kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic torch --from-beginning
```

# On docker container

```
docker exec -it fiagram_kafka-0 bash

# List topics
/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --list

# Create a topic
/opt/kafka/bin/kafka-topics.sh --bootstrap-server localhost:9092 --create --topic torch --partitions 3

# Produce messages (type messages, Enter to send, Ctrl+C to exit)
/opt/kafka/bin/kafka-console-producer.sh --bootstrap-server localhost:9092 --topic torch

# Consume messages
/opt/kafka/bin/kafka-console-consumer.sh --bootstrap-server localhost:9092 --topic torch --from-beginning
```

# Test Messages
```
{"of_strategy_id":119992, "symbol":"VCB", "strategy":"BB20", "type":"SELL"}
```