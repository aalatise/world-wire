
all: dep

dep:
	# go get github.com/golang/dep/cmd/dep
	export PATH=${PATH}:/usr/local/go/bin:${GOPATH}/bin
	dep ensure -v -vendor-only

git-tag:
	@echo $(label)
	git pull \
	&& git tag $(label) \
	&& git push --tags

swaggergen:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	swagger validate api-definitions/participant-api.yaml
	swagger flatten api-definitions/participant-api.yaml  > api-definitions/participant-api.json
	swagger validate api-definitions/admin-api.yaml
	swagger flatten api-definitions/admin-api.yaml  > api-definitions/admin-api.json