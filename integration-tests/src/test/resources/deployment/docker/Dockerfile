FROM openjdk:8
LABEL maintainer=seema.s.phalke@ibm.com

RUN mkdir -p /opt/integrationtests
WORKDIR /opt/integrationtests
COPY . /opt/integrationtests
RUN ./gradlew compileJava \
        && ./gradlew compileGroovy \
        && ./gradlew compileTestJava \
        && ./gradlew compileTestGroovy



