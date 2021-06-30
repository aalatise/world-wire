#!/bin/bash
PARTICIPANT_ID=$1
IBMCLOUD_ES_NAME=$2
SERVICE="ES-SERVICE-KEY-$PARTICIPANT_ID"

# KEY_PASSWORD=$2
# ENVIRONMENT=$3
# BASEPATH=$4
# DOCKERREGISTRYURL=$5
# ORGID=$6
# VERSION=$7

ibmcloud es init

# Create service ids for microservices to connect to ES
ibmcloud resource service-key-create "$SERVICE" "Manager" --instance-name "$IBMCLOUD_ES_NAME"
SERVICE_ID_APIKEY=$(ibmcloud resource service-key $SERVICE --output json | jq -r '.[] | .credentials.apikey')
# Store their api_keys in k8s secrets
kubectl create secret generic kafka-secret-${PARTICIPANT_ID} --from-literal=KAFKA_KEY_SASL_PASSWORD=$SERVICE_ID_APIKEY

G2_TOPICS="${PARTICIPANT_ID}_TRANSACTIONS"

for TOPIC in $G2_TOPICS
do
ibmcloud es topic-create --name $TOPIC --partitions 3
done

# get zookeeper connect
# ZOOKEEPER=$(aws kafka list-clusters --cluster-name-filter $MSKNAME | jq '.ClusterInfoList[].ZookeeperConnectString' | sed -e 's/^"//' -e 's/"$//')

# create participate certs, push and save in k8s secret
# CA_ARN=$(aws acm-pca list-certificate-authorities | jq '.CertificateAuthorities[].Arn' | sed -e 's/^"//' -e 's/"$//')
# CERT_DN="$PARTICIPANT_ID.$ENVIRONMENT.io"
# CERT_ARN=$(aws acm request-certificate --domain-name $CERT_DN --certificate-authority-arn $CA_ARN | jq '.CertificateArn' | sed -e 's/^"//' -e 's/"$//' )

# MSKPATH="/var/files/msk/$PARTICIPANT_ID"
# mkdir -p $MSKPATH

# wait 10 seconds for the cert to be issued
# sleep 10s

# get the private certificate
# aws acm get-certificate --certificate-arn $CERT_ARN | jq '.Certificate' | sed -e 's/^"//' -e 's/"$//' -e 's/\\n/\n/g' >> "$MSKPATH/kafka_cert.crt"
# 
# key password from secret manager
# KAFKA_KEY_PASSWORD="$PARTICIPANT_ID-$KEY_PASSWORD"
# get the private key
# aws acm export-certificate --certificate-arn $CERT_ARN --passphrase $KAFKA_KEY_PASSWORD | jq '.PrivateKey' | sed -e 's/^"//' -e 's/"$//' -e 's/\\n/\n/g' >> "$MSKPATH/kafka_key.key"

# store the cert and private key into the k8s secret
# kubectl create secret generic -n default kafka-secret-$PARTICIPANT_ID --from-file=$MSKPATH/kafka_cert.crt --from-file=$MSKPATH/kafka_key.key --from-literal=kafka_key_password=$KAFKA_KEY_PASSWORD

# rm $MSKPATH/kafka_cert.crt $MSKPATH/kafka_key.key

# create kafka topics and grant acl for producer and consumer to use those topics
# sed "s/{{ PARTICIPANT_ID }}/$PARTICIPANT_ID/g" $BASEPATH/create_topic_anchor.template.yaml \
# | sed "s/{{ DOCKER_REGISTRY_URL }}/$DOCKERREGISTRYURL/g" \
# | sed "s/{{ VERSION }}/$VERSION/g" \
# | sed "s/{{ ZOOKEEPER }}/$ZOOKEEPER/g" \
# | sed "s/{{ DN }}/$CERT_DN/g" \
# > $MSKPATH/create_topic_anchor.$PARTICIPANT_ID.yaml

# kubectl label namespace kafka-topics istio-injection=disabled --overwrite

# kubectl create -n kafka-topics -f $MSKPATH/create_topic_anchor.$PARTICIPANT_ID.yaml
# kubectl wait -n kafka-topics --timeout=300s --for=condition=complete job/$PARTICIPANT_ID-create-topic-anchor