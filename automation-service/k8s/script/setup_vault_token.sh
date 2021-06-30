#!/bin/bash

PARTICIPANTID=$1
VAULT_TOKEN=$2
VAULT_ADDR=$3

ROLENAME="sandbox-ww-${PARTICIPANTID}"

PAYLOAD=$(jq -n \
--arg saname "$ROLENAME" \
--arg pname "$ROLENAME" \
'{"bound_service_account_names": $saname, "bound_service_account_namespaces": "default", "policies": [$pname], "max_ttl": 0, "token_ttl": 0}')

echo "${PAYLOAD}"

# Create a role in the Vault
RESULT=$(curl -k --header "X-Vault-Token: ${VAULT_TOKEN}" --request POST --data "${PAYLOAD}" "${VAULT_ADDR}/v1/auth/kubernetes/role/${ROLENAME}")

echo "${RESULT}"

sed "s/role_name_variable/$ROLENAME/g" /var/k8s/script/participant-service-account-template.yaml \
> "/var/k8s/script/participant-service-account-${PARTICIPANTID}.yaml"

# Create the participant service account
kubectl apply -f "/var/k8s/script/participant-service-account-${PARTICIPANTID}.yaml"

# Get the JWT token from the participant service account token secret
ETOKEN=$(kubectl get secret -o json | jq -r --arg name "$ROLENAME" '.items[] | select(.metadata.annotations["kubernetes.io/service-account.name"]==$name) | .data.token')
JWT=$(echo $ETOKEN | base64 -d)

LOGINPAYLOAD=$(jq -n \
--arg jwt "$JWT" \
--arg role "$ROLENAME" \
'{"jwt": $jwt, "role": $role}')

echo "${LOGINPAYLOAD}"

# Get the Vault client token
TOKEN=$(curl -k --request POST --data "${LOGINPAYLOAD}" "${VAULT_ADDR}/v1/auth/kubernetes/login" | jq '.auth.client_token' | sed -e 's/^"//' -e 's/"$//' -e 's/\\n/\n/g')

echo "${TOKEN}"
# Get the Vault CA certificate
VAULT_CERT=$(kubectl get secret vault-server-tls -n vault -o json | jq '.data."vault.crt"' | sed -e 's/^"//' -e 's/"$//' -e 's/\\n/\n/g')

echo "${VAULT_CERT}"

kubectl create secret generic vault-cert-${PARTICIPANTID} --from-literal=TOKEN=${TOKEN} --from-literal=VAULT_CERT=${VAULT_CERT}

POLICYPATH="ww/data/sandbox/${PARTICIPANTID}/*"
# Creat Vault policy
POLICY="path \\\"$POLICYPATH\\\" {capabilities = [\\\"create\\\", \\\"update\\\", \\\"read\\\", \\\"list\\\"]} path \\\"ww/data/sandbox/ww/*\\\" {capabilities = [\\\"create\\\", \\\"update\\\", \\\"read\\\", \\\"list\\\"]}"
POLICYPAYLOAD=$(cat <<EOF
{
  "policy": "$POLICY"
}
EOF
)

curl -k --header "X-Vault-Token: ${VAULT_TOKEN}" --request PUT --data ''"$POLICYPAYLOAD"'' "${VAULT_ADDR}/v1/sys/policy/${ROLENAME}"