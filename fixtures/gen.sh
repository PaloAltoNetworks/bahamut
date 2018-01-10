#!/bin/bash

cd "$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )" || exit 1
mkdir -p ./certs
tg cert --out ./certs --name "ca" --is-ca --org "Aporeto" --common-name "localhost" --force
tg cert --out ./certs --name "client" --auth-client  --common-name "localhost" --signing-cert ./certs/ca-cert.pem --signing-cert-key ./certs/ca-key.pem --force
tg cert --out ./certs --name "server" --auth-server  --common-name "localhost" --signing-cert ./certs/ca-cert.pem --signing-cert-key ./certs/ca-key.pem --force
