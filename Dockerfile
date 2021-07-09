FROM golang:alpine as builder

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


