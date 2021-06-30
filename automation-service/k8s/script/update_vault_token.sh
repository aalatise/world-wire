#!/bin/bash

VAULT_ADDR=$(kubectl get cm env-config-global -o json | jq '.data.VAULT_ADDR')

DEPLOYLIST=$(kubectl get deployment -o go-template --template '{{range .items}}{{.metadata.name}}{{"\n"}}{{end}}')

PARTICIPANTID="empty"

for D in $DEPLOYLIST
do
  IFS='-' # hyphen (-) is set as delimiter
  read -ra NAME <<< "$D"
  IFS=''

  if [ "${NAME[0]}" != $PARTICIPANTID ]; then
    if [ "${NAME[0]}" != "ww" ] && [ "${NAME[0]}" != "automation" ]; then
      PARTICIPANTID="${NAME[0]}"
      ROLENAME="sandbox-ww-${PARTICIPANTID}"

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

      VAULT_CERT=$(kubectl get secret vault-server-tls -n vault -o json | jq '.data."vault.crt"' | sed -e 's/^"//' -e 's/"$//' -e 's/\\n/\n/g')

      echo "${VAULT_CERT}"

      kubectl create secret generic vault-cert-${PARTICIPANTID} --from-literal=TOKEN=${TOKEN} --from-literal=VAULT_CERT=${VAULT_CERT} --dry-run=client -o yaml | kubectl apply -f -
    fi
  fi
done