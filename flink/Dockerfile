FROM flink:1.19.1

# Accept build arguments
ARG TARGETARCH
ARG USER_ID=1001
ARG GROUP_ID=1001

# Install software dependencies
RUN apt-get update && apt-get install -y \
    software-properties-common \
    wget \
    && rm -rf /var/lib/apt/lists/*

# Add deadsnakes PPA to fetch Python 3.11
RUN add-apt-repository ppa:deadsnakes/ppa
RUN apt-get update && apt-get install -y \
    python3.11 \
    python3.11-dev \
    python3.11-venv \
    python3-pip \
    openjdk-11-jdk \
    maven \
    && rm -rf /var/lib/apt/lists/*

# Install pip for Python 3.11
RUN wget https://bootstrap.pypa.io/get-pip.py && \
    python3.11 get-pip.py && \
    rm get-pip.py

RUN python3.11 -m pip install --upgrade pip

# Set Python 3.11 as the default python
RUN update-alternatives --install /usr/bin/python python /usr/bin/python3.11 1

# Set JAVA_HOME environment variable
ENV JAVA_HOME=/usr/lib/jvm/java-11-openjdk-$TARGETARCH
ENV PATH=$PATH:$JAVA_HOME/bin

# Download dependencies using Maven and copy them into the /opt/flink/lib directory
COPY ./pom.xml /opt/flink/pom.xml
RUN mvn -f /opt/flink/pom.xml dependency:copy-dependencies -DoutputDirectory=/opt/flink/lib/

# install custom libraries
COPY ./python-modules/process_lol_matches-0.2.0-py3-none-any.whl /opt/flink/python-modules/
RUN pip install /opt/flink/python-modules/process_lol_matches-0.2.0-py3-none-any.whl

# Install Python dependencies in the java dependency by flink
COPY requirements.txt /opt/flink/requirements.txt
RUN JAVA_HOME=/usr/lib/jvm/java-11-openjdk-$TARGETARCH pip install -r /opt/flink/requirements.txt

# Set the Python path where the dependencies got installed
ENV PYTHONPATH="/usr/local/lib/python3.11/dist-packages:${PYTHONPATH}"

# Create a non-root user with the specified UID and GID
RUN groupadd -g $GROUP_ID flinkgroup && \
    useradd -m -u $USER_ID -g flinkgroup -s /bin/bash flinkuser

# Ensure the non-root user has access to necessary directories
RUN mkdir -p /opt/flink/log && \
    chown -R flinkuser:flinkgroup /opt/flink/log
RUN mkdir -p /opt/flink/output && \
    chown -R flinkuser:flinkgroup /opt/flink/output

# Switch to the non-root user
#USER flinkuser



# Set the workdir to a directory owned by the non-root user
WORKDIR /opt/flink

# Add config file
ADD ./config.yaml /opt/flink/conf/config.yaml

# Set the entrypoint back to the original Flink image entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]
