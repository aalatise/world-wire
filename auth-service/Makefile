FILE=../VERSION
VARIABLE=`cat $(FILE)`
export VERSION=${VARIABLE} created on: `date +'%y-%m-%d::%H:%M:%S'`

docker:
	
	# install dependencies
	bash install_dependencies.sh --gcloud

	# run unit tests
	# sh run-tests.sh

	docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
	docker pull ${RES_ICR_URL}/gftn/node-alpine
	docker tag ${RES_ICR_URL}/gftn/node-alpine gftn/node-alpine:latest

	# create images
	bash ./deployment/docker/create-images.sh --version ${VARIABLE}