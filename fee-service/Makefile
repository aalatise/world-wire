FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

all: docker

docker:
	docker build --build-arg BUILD_VERSION="${VERSION}" -f Dockerfile -t fee-service .

docs:
	cd api-definitions \
		&& npm install \
		&& npm run md-docs-client-api \


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

gitsub:
	# Setting up your project in vscode may require that you add the utility repo as a submodule
	# Reminder: Make sure you have the correct branch compatabile with your api-service branch
	# eg: PS C:\Users\ChaseOliphant\go\src\github.ibm.com\gftn\api-service> git submodule add --force  ../utility
	git submodule add --force  ../utility