
build:
	./gradlew compileJava
	./gradlew compileGroovy
	./gradlew compileTestJava
	./gradlew compileTestGroovy

service-docker:
ifeq ($(build), all)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_DOCKER_REGISTRY}
	docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
	cd .. \
		&& make dep
	cd ../gftn-models \
		&& make
	cd ../api-service \
		&& make docker
	cd ../quotes-service \
		&& make docker
	cd ../global-whitelist-service \
		&& make docker
	cd ../crypto-service \
		&& make docker \
    	&& make docker-prod
	cd ../payment-listener \
		&& make docker
	cd ../participant-registry \
		&& make docker
	cd ../administration-service \
		&& make docker
	cd ../auth-service \
		&& make docker
	cd ../anchor-service \
		&& make docker
	cd ../send-service \
		&& make docker
	cd ../payout-service \
		&& make docker
	cd ../fee-service \
		&& make docker
	cd ../ww-gateway \
		&& make docker			
else ifeq ($(build), gas-service)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_DOCKER_REGISTRY}
	docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
	cd ../gas-service \
		&& make docker
else ifeq ($(build), world-wire-web)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_DOCKER_REGISTRY}
	docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
	cd ../world-wire-web \
		&& docker build -f Dockerfile -t gftn/ww-portal .
else ifeq ($(build), auth-service)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_DOCKER_REGISTRY}
	docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
	cd ../auth-service-go \
		&& make docker
else ifeq ($(build), crypto-service)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_DOCKER_REGISTRY}
	docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
	cd ../crypto-service \
		&& make docker \
		&& make docker-prod
else
		cd .. \
			&& make dep
		cd ../gftn-models \
			&& make
		cd ../$(build) \
		&& make docker || exit 1
endif



service-docs:
	cd ../gftn-models \
		&& make
	cd ../api-service \
		&& make swaggergen && make docs
	cd ../anchor-service \
		&& make swaggergen && make docs
	cd ../quotes-service \
		&& make swaggergen && make docs
	cd ../participant-registry \
		&& make swaggergen && make docs
	cd ../administration-service \
		&& make swaggergen && make docs
	cd ../crypto-service \
		&& make swaggergen && make docs
	cd ../payout-service \
		&& make swaggergen && make docs
	cd ../global-whitelist-service \
		&& make swaggergen && make docs

build-go:
	cd ../gftn-models \
		&& make
	cd ../api-service \
		&& go build
	cd ../participant-registry \
		&& go build
	cd ../payment-listener \
		&& go build
	cd ../administration-service \
		&& go build
	cd ../crypto-service \
		&& go build
	cd ../fee-service \
		&& go build
	cd ../anchor-service \
		&& go build
	cd ../quotes-service \
		&& go build
	cd ../send-service \
		&& go build
	cd ../payout-service \
		&& go build
	cd ../global-whitelist-service/whitelistserver \
    	&& go build
	cd ../ww-gateway \
    	&& go build
docker:
	docker build -f src/main/docker/Dockerfile -t gftn/integration-tests .

push-dockers:
	@echo $(version)
	docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_DOCKER_REGISTRY}
	docker tag gftn/api-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/api-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/api-service:$(version)
	docker tag gftn/crypto-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/crypto-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/crypto-service:$(version)
	docker tag gftn/crypto-service-prod:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/crypto-service-prod:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/crypto-service-prod:$(version)
	docker tag gftn/participant-registry:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/participant-registry:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/participant-registry:$(version)
	docker tag gftn/quotes-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/quotes-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/quotes-service:$(version)
	docker tag gftn/administration-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/administration-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/administration-service:$(version)
	docker tag gftn/payment-listener:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/payment-listener:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/payment-listener:$(version)
	docker tag gftn/anchor-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/anchor-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/anchor-service:$(version)
	docker tag gftn/gas-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/gas-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/gas-service:$(version)
	docker tag gftn/payout-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/payout-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/payout-service:$(version)
	docker tag gftn/send-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/send-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/send-service:$(version)
	docker tag gftn/global-whitelist-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/global-whitelist-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/global-whitelist-service:$(version)
	docker tag gftn/fee-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/fee-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/fee-service:$(version)
	docker tag gftn/ww-gateway:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/ww-gateway:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/ww-gateway:$(version)
	docker tag gftn/automation-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/automation-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/automation-service:$(version)
	docker tag gftn/auth-service:latest ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/auth-service:$(version)
	docker push ${RES_DOCKER_REGISTRY}/${RES_ICP_NAMESPACE}/auth-service:$(version)

push-to-icr:
	@echo "Version is $(version)"
	docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
	# Print number of images in docker daemon
	docker images -a --filter='dangling=false' --filter=reference='gftn/*:*' | wc -l
	# List the images
	docker images -a --filter='dangling=false' --filter=reference='gftn/*:*' --format '{{.Repository}}:{{.Tag}}'
	# Push the images
	docker images -a --filter='dangling=false' --filter=reference='gftn/*:*' --format '{{.Repository}}:{{.Tag}} {{.Repository}}' | xargs -n 2 -t sh -c 'docker tag $$0 ${RES_ICR_URL}/$$1; docker push ${RES_ICR_URL}/$$1:$(version)'

push-administration-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_DOCKER_REGISTRY}
	docker tag gftn/administration-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/administration-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/administration-service:$(version)

push-auth-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/auth-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/auth-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/auth-service:$(version)

push-anchor-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/anchor-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/anchor-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/anchor-service:$(version)

push-api-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/api-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/api-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/api-service:$(version)

push-crypto-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/crypto-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/crypto-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/crypto-service:$(version)
	docker tag gftn/crypto-service-prod:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/crypto-service-prod:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/crypto-service-prod:$(version)

push-fee-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/fee-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/fee-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/fee-service:$(version)

push-gas-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/gas-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/gas-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/gas-service:$(version)

push-global-whitelist-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/global-whitelist-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/global-whitelist-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/global-whitelist-service:$(version)

push-participant-registry-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/participant-registry:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/participant-registry:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/participant-registry:$(version)

push-payment-listener-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/payment-listener:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/payment-listener:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/payment-listener:$(version)

push-payout-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/payout-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/payout-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/payout-service:$(version)

push-quotes-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/quotes-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/quotes-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/quotes-service:$(version)

push-send-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/send-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/send-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/send-service:$(version)


push-ww-gateway-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/ww-gateway:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/ww-gateway:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/ww-gateway:$(version)

push-automation-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/automation-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/automation-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/automation-service:$(version)

push-portal-api-service-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/portal-api-service:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/portal-api-service:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/portal-api-service:$(version)

push-ww-portal-dockers:
	@echo $(version)
	# docker login -u ${DOCKER_USER} -p "${DOCKER_PASSWORD}" ${RES_ICR_URL}
	docker tag gftn/ww-portal:latest ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/ww-portal:$(version)
	docker push ${RES_ICR_URL}/${RES_ICP_NAMESPACE}/ww-portal:$(version)