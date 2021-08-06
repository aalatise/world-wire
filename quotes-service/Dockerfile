FROM apk-go-ww as builder
LABEL maintainer=kingaj@us.ibm.com
ARG BUILD_VERSION

RUN echo $BUILD_VERSION

# Vars
ENV PARENT_REPO=github.com/IBM
ENV PARENT_PACKAGE=$PARENT_REPO/world-wire
ENV SERVICE_NAME=quotes-service

# Setup Runtime Dir Structure
RUN mkdir -p /go/src/$PARENT_PACKAGE/$SERVICE_NAME

# Set necessary environment variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=1  \
    GOARCH="amd64" \
    GOOS=linux

# Copy code
COPY . /go/src/$PARENT_PACKAGE/$SERVICE_NAME
RUN cd /go/src/$PARENT_PACKAGE/$SERVICE_NAME \
        && pwd \
        && go mod tidy \
        && go build -tags dynamic \
        && mv $SERVICE_NAME /tmp/$SERVICE_NAME

# Wrapper image	
FROM alpine:latest	
ARG BUILD_VERSION
ENV ENV_BUILD_VERSION=$BUILD_VERSION

RUN apk --no-cache add ca-certificates

# TLS & logs
RUN mkdir -p /etc/tls
RUN mkdir -p /opt/error-codes/
COPY error-codes/ /opt/error-codes/
RUN mkdir -p /opt/logs/
RUN touch /opt/logs/logs.txt

# Create a group and user
RUN addgroup -S wwgroup && adduser -S wwuser -G wwgroup

# Copy executable and chown
COPY --chown=wwuser:wwgroup --from=builder /tmp/$SERVICE_NAME .

# Change owner for log file
RUN chown wwuser /opt/logs/logs.txt

# Tell docker that all future commands should run as the appuser wwuser
USER wwuser

EXPOSE 11234
CMD ["./quotes-service"]
