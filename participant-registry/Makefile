FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: swaggergen docker docs lint test


docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t participant-registry .


swaggergen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	swagger flatten api-definitions/stellar-participant-registry-api.yaml  > api-definitions/participant-registry-api.json


docs:
	cd api-definitions \
		&& npm install \
		&& npm run md-docs-participant-registry

dep:
	go mod tidy

build:
	go build

lint:
	gometalinter --vendor ./... --disable-all --enable=golint

test:
	go test -run TestAuthForExternalEndpoint

install:
	go install
