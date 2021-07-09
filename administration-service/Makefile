FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: dep swaggergen docker docs


docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t administration-service .


swaggergen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	swagger validate api-definitions/internal-api.yaml
	swagger flatten api-definitions/internal-api.yaml  > api-definitions/internal-api.json

docs:
	cd api-definitions \
		&& npm install \
		&& npm run md-docs-internal-api

dep:
	go mod tidy
	go mod download

lint:
	gometalinter --vendor ./... --disable-all --enable=golint

test:
	go test -run TestAuthForExternalEndpoint

build:
	go build -a -v -tags musl

install:
	go get -d -v
