#!/bin/bash
export VER="1.4.2"
wget https://releases.hashicorp.com/packer/${VER}/packer_${VER}_linux_amd64.zip
unzip packer_${VER}_linux_amd64.zip
sudo mv packer /usr/local/bin
packer --version
# docker login -u ${DOCKER_USER} -p ${DOCKER_PASSWORD} ${DOCKER_REGISTRY}
docker login -u iamapikey -p ${RES_IBM_SERVICE_ID_API_KEY} ${RES_ICR_URL}
echo 'Starting Packer build'
packer build -machine-readable \
    -var "RES_ICR_URL=${RES_ICR_URL}" \
    packer-node-alpine.json
docker push ${RES_ICR_URL}/gftn/node-alpine:latest