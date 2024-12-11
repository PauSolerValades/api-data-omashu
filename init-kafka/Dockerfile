FROM confluentinc/cp-kafka:latest

WORKDIR /app

COPY create_topics_with_schemas.sh /app/create_topics_with_schemas.sh
COPY kafka-ready.properties /app/kafka-ready.properties
COPY topics_and_schemas.txt /app/topics_and_schemas.txt

USER root
RUN chmod +x /app/create_topics_with_schemas.sh
# Remove Windows-style carriage return characters (\r) from the shell script and topics file
RUN sed -i 's/\r$//' /app/create_topics_with_schemas.sh
RUN sed -i 's/\r$//' /app/topics_and_schemas.txt

USER appuser
CMD ["/bin/bash", "/app/create_topics_with_schemas.sh"]
