FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: docker


docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t portal-api-service .

dep:
	go mod tidy

build:
	go build

install:
	go install
