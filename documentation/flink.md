# Flink

## Limitations sorted from Implementation and Developement

There are several aspects to be discussed on the Docker setup of the Flink image to use with python, specially when trying to use it with Confluentic's AVRO and Schema Registry in Kafka.

### Confluentics AVRO Schema Registry

There is **no native support in DataStreamAPI** to consume from a topic using the schema registry. The autor of this documentation has tried several instances of Custom decoders for Kafka to use in Python, but none of the tried have succeed on implement a desirable and reproducible behaviour. Therefore, two options were left:
- Implement this whole part of the project in Java: no ability to call the custom functions written for the company during one year.
- Implement the Java Class and use it as a compiled jar in the DatastreamAPI: it could have worked and it's technically the most optimal option, but its complexity and difficulty makes it an unfeasible solution in my context. Besides, no one knows Java in the company but me, and my level is pretty low.
- Use the TableAPI and translate to DataStreamAPI: It's impelemnted natively and suports Confuentic AVRO out-of-the-box, as well as kafka topic consumption.

The third option has been implemented, but not with it's concessions and limitations due to version incompatibility and tooling.

### Java Dependencies and Flink Version
To consume from a topic natively, Flink needs (at least) the following java dependencies (.jar) where JAVA_PATH is specified:

``` xml
<dependency>
  <groupId>org.apache.flink</groupId>
  <artifactId>flink-avro</artifactId>
  <version>1.19.0</version>
  <scope>compile</scope>
</dependency>
<dependency>
  <groupId>org.apache.flink</groupId>
  <artifactId>flink-avro-confluent-registry</artifactId>
  <version>1.19.0</version>
  <scope>compile</scope>
</dependency>
<dependency>
  <groupId>org.apache.flink</groupId>
  <artifactId>flink-sql-connector-kafka</artifactId>
  <version>3.2.0-1.19</version>
  <scope>compile</scope>
</dependency>
```

As of today (12/09/2024) Flink is at version 1.20, but ```flink-sql-connector-kafka``` has no compatible version with Flink 1.20. Therefore, to keep using Kafka we must downgrade Flink to version 1.19.1, the latest version which has the ```flink-sql-connector-kafka``` developed and working. The biggest implication is that Python 3.11 is the lastest version supported by Flink, and the one that gets installed with it.

To manage all the Java dependencies (and to avoid entering dependency hell) a Maven project has been set up to manage all all the needed dependencies, and install them when Docker goes up.

TODO: REDUCE THE AMOUNT OF DEPENDENCIES WHICH ARE IN THE POM.XML TO THE MINIMUM SPECIFIED BY ALL THE DOCUMENTATION.

### Python Installation

When installing ```python``` as a normal dependency in the ```Dockerfile```, 3.10 gets downloaded. To overwrite this, the ```Dockerfile``` updated the available versions of python with deadsnakes at the beginning of the dockerfile, and then ```Python3.11``` gets installed as default for the whole image. 

To install ```PyFlink``` library, it's need to install the same version as Flink is running. Right now this is located at _flink/requirements.txt_, but optimally should be installed when the docker is building, just for correctness (this is not a trivial dependency, it's a crucial one central to the system)

### Flink's AVRO types

Flink's [AVRO format](https://nightlies.apache.org/flink/flink-docs-release-1.19/docs/connectors/table/formats/avro/#data-type-mapping) support is rather limited in comparison with [confluentic specification](https://nightlies.apache.org/flink/flink-docs-release-1.19/docs/connectors/table/formats/avro-confluent/) and the complete AVRO specification. Specifically, _match-data_ AVRO schema has the ```dataType``` as _String_ instead of _Enum_, as well as having no support for ```logical_type```, therefrore the uuid and timestamps are also encoded as basic types.

### Restrictions
In the documentation for [Python Config](https://nightlies.apache.org/flink/flink-docs-master/docs/dev/python/python_config/), there are two options to configure the paths and the external libraries in _flink/config.yml_, which are the following:

``` yml
python.files: path/to/your/wheel.whl
python.pythonpath: path/to/your/path/with/libraries
```

They have been tried to provide access to the _process-lol-matches_ custom library, but the author has not managed to get it working. Desipite that, the libraries are installed coping the python-modules folder and intstalling its contents. The other two python yml for the version and executable seems to work.

To also have it written, I managed to make the python.files option work with a whl compiled with ```python setup.py bwheel``` instead of the recommended ```python -m build --wheel```.

### Volumes and permissions
For Flink to write the data to the file ```./flink/ouptut``` a volume is used (check ```docker-compose.yml```) between that folder and the container. Flink outputs generate a folder with the date, and the output at hourly basis, like the following tree as an example:

``` tree
./flink/output/
├── 2024-10-02--12
│   └── part-jobID-0
├── 2024-10-02--12
│   └── part-jobID-0
└── 2024-10-03--14
    ├── part-jobID-0
    ├── part-jobID-1
    └── part-jobID-2
```

To ensure that the container process is able to generate the directories properly, the ```USER_ID``` and ```GROUP_ID``` are specified in the container. There are three ways to specify them:
1. Globally: running
```
export UID GID
```
2. Locally: in the same shell that you'll run docker compose run the code below. This will load the users varibales values properly into the docker.
``` bash
export USER_ID=$(id -u)
export GROUP_ID=$(id -g)
```
3. Giving them value: the proper values for those groups are:
```
USER_ID=1000
GLOBAL_ID=1001
```

The Flink build process **will not take care of setting the proper permissions if the directories ```output``` and ```logs``` were already created** making errors possible in previous clones of the repository. To fix this error (_jobmanager_ falining due to not having permissions to write to the _output_ directory) the following commands must be invoqued to change the permissions of the existing folders:
```
chown $(id -u):$(id -g) ./flink/output
```

where the id should be properly specified.

