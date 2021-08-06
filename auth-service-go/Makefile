FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: docker

docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t auth-service .

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
