FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: swaggergen docker docs lint test

deploy: docker push

docker:
	rsync -av --exclude='node_modules' --exclude='checkmarx' --exclude='integration-tests' --exclude='*/*.txt' --exclude='api-definitions' \
	 --exclude='anchor-service'  --exclude='automation-service' --exclude='*/*.toml' --exclude='*/*.yml'  --exclude='*/*.github'  --exclude='*/*.ts'  --exclude='*/*.json' --exclude='*/*.yaml'\
	  --exclude='*/*.md' --exclude='*/*.gitignore' --exclude='rdo-service' --exclude='rdo-client' --exclude='payout-service' --exclude='participant-registry' \
	  --exclude='send-service' --exclude='callback-service' --exclude='administration-service' --exclude='api-service' --exclude='quotes-service'\
	  --exclude='crypto-service' --exclude='fee-service' --exclude='gas-service' --exclude='payment-listener' --exclude='world-wire-web' --exclude="kubernetes" --exclude="infrastructure" --exclude="travis"\
	 --exclude='.idea' --exclude='global-whitelist-service'\
	  --exclude='.git' --exclude='ww-gateway' --exclude='auth-service' $(subst /auth-service-go,,$(shell pwd)) deployment/docker/vendor >> /dev/null

	docker build --build-arg BUILD_VERSION="${VERSION}" -f deployment/docker/Dockerfile -t gftn/auth-service .
	rm -rf deployment/docker/vendor

dep:
	go get github.com/golang/dep/cmd/dep
	export PATH=$PATH:/usr/local/go/bin:$GOPATH/bin
	dep ensure -vendor-only

build:
	go build

lint:
	gometalinter --vendor ./... --disable-all --enable=golint

test:
	go test -run TestAuthForExternalEndpoint

install:
	go install

push:
	docker tag gftn/auth-service:latest jp.icr.io/dapww/auth-service:latest
	docker push jp.icr.io/dapww/auth-service:latest
