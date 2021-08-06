FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: docker

docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t automation-service .
	rm -rf deployment/docker/vendor/

dep:
	go mod tidy

swaggergen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	swagger validate api-definition/internal-api.yaml
	swagger flatten api-definition/internal-api.yaml  > api-definition/internal-api.json

build:
	go build

install:
	go install

test:
	go test -run TestAuthForExternalEndpoint
