#!/bin/bash

PARTICIPANT=$1
DOCKERTAG=$2
REPLICAS=$3
IBMIKSID=$4

ibmcloud ks cluster config --cluster "$IBMIKSID"

BASE_PATH="/var/k8s"
PARTICIPANTS_PATH="$BASE_PATH/anchor"

OLDPARTICIPANT=participant_id_variable
OLDDOCKERTAG=docker_tag_variable
OLDREPLICAS=replica_variable

mkdir -p $PARTICIPANTS_PATH/$PARTICIPANT

NEW_PARTICIPANT_PATH="$PARTICIPANTS_PATH/$PARTICIPANT"

# copy files into new files
cp -r $BASE_PATH/anchor/template/* $NEW_PARTICIPANT_PATH/

cd $NEW_PARTICIPANT_PATH

# replace
find . -type f | xargs sed -i "s/$OLDPARTICIPANT/$PARTICIPANT/g"
find . -type f | xargs sed -i "s/$OLDDOCKERTAG/$DOCKERTAG/g"
find . -type f | xargs sed -i "s/$OLDREPLICAS/$REPLICAS/g"

mkdir -p "/var/files/anchor"

cp -r $NEW_PARTICIPANT_PATH "/var/files/anchor"

printf "Ready to deploy %s on k8s cluster using %s image\n" "$PARTICIPANT" "$DOCKERTAG"

kubectl apply -f $NEW_PARTICIPANT_PATH/ -n default