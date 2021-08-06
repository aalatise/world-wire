FROM apk-go-ww as builder
ARG BUILD_VERSION
LABEL maintainer=kingaj@us.ibm.com

RUN echo $BUILD_VERSION

# Vars
ENV PARENT_REPO=github.com/IBM
ENV PARENT_PACKAGE=$PARENT_REPO/world-wire
ENV SERVICE_NAME=crypto-service

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
RUN addgroup -S wwgroup && adduser -S wwuser -G wwgroup

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