## status-download topic and cleanup policy
The ```status-download``` topic is structured using a partition key, which in this case is the omashuId. This key ensures that all messages related to a specific ```omashuId``` are directed to the same partition. By using an appropriate cleanup policy, we can overwrite the previous message for each key rather than appending new ones. This approach allows us to continuously update the status of the download while maintaining only the latest message for each ```omashuId```, ensuring efficient storage and retrieval of the most up-to-date status.


## init-kafka and topics_and_schemas.txt

The init-kafka application allows to create the topics and store schemas in the schema registry. To do it, we must see the ```topics_and_schemas.txt``` as a space separated values (ssv) data storage with the folowing header.
- topic_name: name of the topic to create to kafka
- partitions: how many partitions will the topic have
- replication_factor: 
- schema_file: which avro is associated with the topic.
- has_key_schema: specify if this schema must also registrer a key
- cleanup_policy: what's the cleanup policy for the messages.
- make_it_topic: all of the schemas will be registered in the schema registry, and if this is true, they will also be made a topic in kafka as well.


## Schema Registry and AVRO confluentic

When a service consumes/produces to a Kafka topic, needs to know what type of schema is encoding/decoding. To do that, it uses the _schema-registry_ to retrieve the schema for the data, which has been stored the first time a service has produced something in the topic. How it does that is very well explained in the following link <link=https://thangvynam1808.medium.com/kafka-avro-serialization-and-schema-44e0017423b9>, and this behaviour is replicated. The implementation takes into account the structure from a message, which contains **1 magic byte which needs to be ignored** and **four bytes for the schemaID** for the _schema-registry_ to know the registry.


## Documentation

APACHE AVRO and CONSOLE PRODUCER/CONSUMER
- Simple Example: https://developer.confluent.io/tutorials/kafka-console-consumer-producer-avro/kafka.html
- Logical Converter expressions (--property avro.use.logical.type.converters=true) Page: https://docs.confluent.io/platform/current/schema-registry/fundamentals/serdes-develop/index.html 
- Specification: https://avro.apache.org/docs/1.11.1/specification/_print/
- How GO works: https://thangvynam1808.medium.com/kafka-avro-serialization-and-schema-44e0017423b9 
- confluentic-go: https://pkg.go.dev/github.com/confluentinc/confluent-kafka-go@v1.9.2
- flink + python + kafka: https://towardsdatascience.com/how-i-dockerized-apache-flink-kafka-and-postgresql-for-real-time-data-streaming-c4ce38598336
- https://github.com/tiangolo/full-stack-fastapi-template
- https://fastapi.tiangolo.com/deployment/server-workers/
- Set up Apache Kafka (i sobretot com fer-lo escalar per si fes falta) https://www.baeldung.com/ops/kafka-docker-setup
- Kafka consumer golang https://medium.com/@moabbas.ch/effective-kafka-consumption-in-golang-a-comprehensive-guide-aac54b5b79f0
- Topics and Aggregation: https://chatgpt.com/share/4b04526e-5b53-415a-b74f-b6e96199aa14
- kafka super design https://stackoverflow.blog/2022/07/21/event-driven-topic-design-using-kafka/ 
- DynamoDB as a consumer: https://docs.confluent.io/kafka-connectors/aws-dynamodb/current/overview.html