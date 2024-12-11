# api-data

This project aims to provide the infrastructure for data obtenction and analysis. It is based as an event driven architecture with kafka in its core, conformed by the following modules
- backend: FastAPI application which allows external petitions. Maybe are forwarded to database.
- rate-limiter: GO application which centralizes the petitions to RIOT enforcing a certain rate limiters
- DynamoDB: storing middle data for results.

The project aims to serve the OMASHU's webpage data about players and match analysis.






## Kafka Topic Design



## GO stuff
- kafka avro go: https://medium.com/@ipreferwater2/go-kafka-avro-schema-registry-b5ed185442b7
- Issues with the decoding in avro and solution: https://github.com/linkedin/goavro/issues/97









- Best explanation: https://howtodoinjava.com/kafka/kafka-with-avro-and-schema-registry/
## Peculiarities:
_rate-limiter_ uses require github.com/confluentinc/confluent-kafka-go/v2 v2.4.0 which must be higher tha 2.0.0 to have available arm64 compilation. To do that, you have to provide some explicit flags, as explained in the repository: https://github.com/confluentinc/confluent-kafka-go. This means that, if it were to compile for a x86 machine, probably something like ```docker buildxs``` should be used to guarantee execution on x86 machines.