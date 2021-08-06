FROM apk-go-ww as builder
ARG BUILD_VERSION
LABEL maintainer=kingaj@us.ibm.com

RUN echo $BUILD_VERSION

# Vars
ENV PARENT_REPO=github.com/IBM
ENV PARENT_PACKAGE=$PARENT_REPO/world-wire
ENV SERVICE_NAME=automation-service

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
RUN cd /go/src/$PARENT_PACKAGE/$SERVICE_NAME \
        && pwd \
        && go mod tidy \
        && go build -tags dynamic \
        && mv $SERVICE_NAME /tmp/$SERVICE_NAME

FROM alpine:latest
ARG BUILD_VERSION
RUN apk --no-cache add ca-certificates
RUN apk update
RUN apk add --no-cache bash \
    && apk add --no-cache curl \
    && apk add --no-cache jq \
    && apk add --no-cache python3 \
    && apk add --no-cache py-pip \
    && apk add --no-cache sed \
    && set -ex && apk --no-cache add sudo

# TLS & logs
RUN mkdir -p /etc/tls
ENV ENV_BUILD_VERSION=$BUILD_VERSION
RUN curl -fsSL https://clis.cloud.ibm.com/install/linux | sh

RUN curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl \
    && chmod +x ./kubectl \
    && sudo mv ./kubectl /usr/local/bin/kubectl

# Create a group and user
RUN addgroup -S wwgroup && adduser -S wwuser -G wwgroup

# Setup automate deployment script and files
RUN mkdir -p /var/k8s
RUN mkdir -p /var/ibm-es
RUN mkdir -p /var/files
RUN mkdir -p /var/logs
RUN mkdir -p /opt/error-codes/

COPY --chown=wwuser:wwgroup k8s /var/k8s
COPY --chown=wwuser:wwgroup ibm-es /var/ibm-es
COPY --chown=wwuser:wwgroup error-codes /opt/error-codes/

# Change owner for log file
RUN chown -R wwuser /var/k8s
RUN chown -R wwuser /var/ibm-es
RUN chown -R wwuser /var/files
RUN touch /var/logs/logs.txt
RUN chown wwuser /var/logs/logs.txt

# Copy executable and chown
COPY --chown=wwuser:wwgroup --from=builder /tmp/$SERVICE_NAME .

# Tell docker that all future commands should run as the appuser wwuser
USER wwuser

# Install IBM Cloud Plugins and CLI
RUN ibmcloud plugin install event-streams
RUN ibmcloud plugin install kubernetes-service

# RUN pip3 install awscli --upgrade --user
# RUN chmod +x /home/wwuser/.local/bin/aws
ENV PATH=/home/wwuser/.local/bin:$PATH

EXPOSE 5566
CMD ["./automation-service"]