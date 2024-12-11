FROM confluentinc/cp-schema-registry:latest

USER root

RUN yum update -y && yum install -y jq && yum clean all

RUN mkdir /output

COPY ./.env /usr/local/bin/.env

RUN chown -R appuser:appuser /output

USER appuser