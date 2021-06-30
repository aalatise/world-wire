


docker:
	cd ../utility/aws/javascript/ && npm install && npm run build
	mkdir -p ./utility/aws/
	rsync -rupE ../utility/aws/javascript ./utility/aws/ >> /dev/null
	docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker pull ${RES_ICR_URL}/gftn/node-alpine:latest
	docker tag ${RES_ICR_URL}/gftn/node-alpine:latest gftn/node-alpine:latest
	docker images
	docker build -f deployment/Dockerfile -t gftn/gas-service .
install:
	npm install
test:
	npm run test \
	npm run stress-test \

