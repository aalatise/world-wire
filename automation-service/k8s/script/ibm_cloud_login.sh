#!/bin/bash

IBMCLOUD_API_KEY=$1
IBMIKSID=$2

ibmcloud login --no-region --apikey "$IBMCLOUD_API_KEY"
ibmcloud ks cluster config --cluster "$IBMIKSID"