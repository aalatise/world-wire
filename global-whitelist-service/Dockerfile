FROM apk-go-ww as builder
ARG BUILD_VERSION
LABEL maintainer=kingaj@us.ibm.com

RUN echo $BUILD_VERSION

# Vars
ENV PARENT_REPO=github.com/IBM
ENV PARENT_PACKAGE=$PARENT_REPO/world-wire
ENV SERVICE_NAME=global-whitelist-service
ENV SERVER_NAME=whitelistserver

# Setup Runtime Dir Structure
RUN mkdir -p /go/src/$PARENT_PACKAGE/$SERVICE_NAME

# Set necessary environment variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=1  \
    GOARCH="amd64" \
    GOOS=linux

# Copy code
COPY . /go/src/$PARENT_PACKAGE/$SERVICE_NAME

# Build
RUN cd /go/src/$PARENT_PACKAGE/$SERVICE_NAME/$SERVER_NAME \
        && pwd \
        && go mod tidy \
        && go build -tags dynamic \
        && mv $SERVER_NAME /tmp/$SERVER_NAME

FROM alpine:latest
ARG BUILD_VERSION
RUN apk --no-cache add ca-certificates

# Error codes and log file
RUN mkdir -p /opt/logs/
RUN touch /opt/logs/logs.txt

ENV ENV_BUILD_VERSION=$BUILD_VERSION
# TLS & logs
RUN mkdir -p /etc/tls 

# Create a group and user
RUN addgroup -S wwgroup && adduser -S wwuser -G wwgroup

# Copy executable and chown
COPY --chown=wwuser:wwgroup --from=builder /tmp/$SERVER_NAME .
RUN mkdir -p /opt/error-codes/
COPY --chown=wwuser:wwgroup whitelistserver/error-codes/ /opt/error-codes/

# Change owner for log file
RUN chown wwuser /opt/logs/logs.txt

# Tell docker that all future commands should run as the appuser wwuser
USER wwuser

EXPOSE 11234
CMD ["./whitelistserver"]
