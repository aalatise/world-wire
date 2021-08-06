FROM golang:alpine3.13
LABEL maintainer=kingaj@us.ibm.com

# Downgrade to alpine3.13 recommended by
# https://stackoverflow.com/questions/68013058/alpine3-14-docker-libtls-so-20-conflict
# Note that this limits go version to 1.16

RUN apk update \
     && apk add --no-cache bash\
     && apk add --no-cache build-base\
     && apk add --no-cache git\
     && apk add --update --no-cache pkgconfig\
     && apk add --update --no-cache libxml2-dev\
     && apk add --no-cache openssh \
     && apk add --update --no-cache libc-dev

RUN git clone https://github.com/edenhill/librdkafka.git \ 
	&& cd librdkafka \
	&& ./configure \
	&& make \
	&& make install


