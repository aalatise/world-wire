FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: swaggergen docker docs


docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t quotes-service .

swaggergen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	swagger flatten api-definitions/client-callback.yaml  > api-definitions/quotes-client-callback-api.json

swaggergen-model:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	swagger generate model --with-flatten=full --spec=models-swag/model-definitions/gftn-models.yaml --model-package=model --target=.
docs:
	cd api-definitions \
		&& npm install \
		&& npm run md-docs-quotes-callback-api


dep:
	go mod tidy

build:
	go build 

lint:
	gometalinter --vendor ./... --disable-all --enable=golint

install:
	go install

test:
	go test -run TestAuthForExternalEndpoint
