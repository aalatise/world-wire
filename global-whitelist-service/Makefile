FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: swaggergen docker docs

docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t global-whitelist-service .

swaggergen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger

docs:
	cd api-definitions \
		&& npm install


dep:
	go mod tidy

build:
	cd ./whitelistserver && go build 

lint:
	gometalinter --vendor ./... --disable-all --enable=golint

install:
	cd ./whitelistserver && go install

test:
	cd ./whitelistserver && go test -run TestAuthForExternalEndpoint
