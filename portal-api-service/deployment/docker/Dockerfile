FROM golang:alpine as builder
LABEL maintainer=lung.chen.huang@ibm.com
ARG BUILD_VERSION
RUN echo $BUILD_VERSION

# Vars

ENV PARENT_REPO=github.ibm.com/gftn
ENV PARENT_PACKAGE=$PARENT_REPO/world-wire-services
ENV SERVICE_NAME=portal-api-service

RUN apk add --update --no-cache git \
    && apk add --update --no-cache make\
    && apk add --no-cache openssh \
    && apk add --no-cache build-base

# Setup Runtime Dir Structure
RUN mkdir -p /go/src/$PARENT_PACKAGE/$SERVICE_NAME \
    && mkdir -p /go/src/$PARENT_PACKAGE/$SERVICE_NAME/vendor


COPY deployment/docker/vendor/world-wire-services /go/src/$PARENT_PACKAGE/
COPY . /go/src/$PARENT_PACKAGE/$SERVICE_NAME
RUN cd /go/src/$PARENT_PACKAGE/$SERVICE_NAME \
        && go get gopkg.in/go-playground/validator.v9 \
        && pwd \
        && make install \			
        && make build \	
        && mv portal-api-service /tmp/portal-api-service \
        && rm -Rf /go/src/*

#2nd stage
FROM alpine:latest
ARG BUILD_VERSION
RUN apk --no-cache add ca-certificates
COPY --from=builder /tmp/portal-api-service .
RUN pwd && chmod a+x portal-api-service

# Error codes and log file
RUN mkdir -p /opt/error-codes/
COPY error-codes/ /opt/error-codes/
RUN mkdir -p /opt/logs/
RUN touch /opt/logs/logs.txt

# Copy code
RUN echo $BUILD_VERSION
ENV ENV_BUILD_VERSION=$BUILD_VERSION

# TLS & logs
RUN mkdir -p /etc/tls 

# Create a group and user
RUN addgroup -S wwgroup && adduser -S wwuser -G wwgroup
# Change owner for log file
RUN chown wwuser /opt/logs/logs.txt
# Tell docker that all future commands should run as the appuser wwuser
USER wwuser

EXPOSE 8686
CMD ["./portal-api-service"]
