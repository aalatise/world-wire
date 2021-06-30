FROM openjdk:8-jdk
MAINTAINER Michael Dowling michael.d.dowling@ibm.com

RUN mkdir -p /opt/gftn/integration-tests
COPY . /opt/gftn/integration-tests

WORKDIR /opt/gftn/integration-tests
RUN apt-get update \
      && apt-get install -y build-essential


CMD ["make", "build"]