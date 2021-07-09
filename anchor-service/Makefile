FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`
all: dep swaggergen docker docs

docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t anchor-service .

swaggergen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	swagger validate api-definitions/anchor-api.yaml
	swagger validate api-definitions/anchor-onboarding-api.yaml
	swagger flatten api-definitions/anchor-api.yaml  > api-definitions/anchor-api.json
	swagger flatten api-definitions/anchor-onboarding-api.yaml  > api-definitions/anchor-onboarding-api.json

docs:
	cd api-definitions \
		&& npm install \
		&& npm run md-docs-anchor-api \
		&& npm run md-docs-anchor-onboarding-api

dep:
	go mod tidy

lint:
	gometalinter --vendor ./... --disable-all --enable=golint

test:
	go test -run TestAuthForExternalEndpoint

build:
	go build

install:
	go install
