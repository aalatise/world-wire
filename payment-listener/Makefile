FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: swaggergen docker docs

docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t payment-listener .

swaggergen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger

docs:
	cd api-definitions \
		&& npm install


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
