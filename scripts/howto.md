# Scripts

This folder contains scripts to simulate interactions with other systems and to exectue certain common commands more easliy. All the interaction with kafka is achieved with the _confluentinc/cp-schema-registry_ to ensure the minimitzation of errors coming from code that is not meant to go into production.

## How to run the scripts

To call any script, the file *run_docker.sh* must be called like this:

``` bash
./run_docker.sh <path/to/Dockerfile> <path/to/avro-schemas> <path/to/sample-data> <path/to/script-to-execute.sh> <args_to_the_script>
```

Which the path/to/Dockerfile is always the same file, but must be provided for the executing outside the folder case. Tha args to the script will go directly to the script for it to catch. For example, if you want to consume from a topic:

``` bash
./scripts/run_docker.sh ./scripts/Dockerfile ./avro-schemas ./sample-data/ ./scripts/consume_from.sh download-user-exists
```

All the stats data, regardless of the schema it follows, is in the stats-player folder for it to be iterated upon with a simple bash script not provided here. For each file, there is data of the same type properly specified in the name

# How to run produce_to and design

The idea is to run the *produce_to.sh* script

``` bash
./scripts/run_docker.sh ./scripts/Dockerfile ./avro-schemas ./sample-data ./scripts/produce_to.sh <TOPIC_NAME> <SCHEMA_ID> </PATH/TO/DATAFILE.txt>
```

The schema_id needs to be found in the schema registry. In general, you can check them easily with this chain of commands.

``` bash
# find all the topics
docker compose exec -it <CONTAINTER_NAME> curl -X GET http://schema-registry:8081/subjects

# call the topic exact name with the latest version
docker compose exec -it <CONTAINTER_NAME> curl -X GET http://schema-registry:8081/subjects/{TOPIC_NAME}/versions/latest
```

The response of the latter command will have the _id_ parameter of the JSON.

The schemas are also set in order, therefore, as they appear in the topics_and_semantics.txt will probably cooincide with the id value.

# How to consume_from

When providing the topic name, the -value must be always provided:

``` bash
./scripts/run_docker.sh ./scripts/Dockerfile ./avro-schemas ./sample-data ./scripts/produce_to.sh <TOPIC_NAME>-value
```
