#!/usr/bin/env sh

mkdir -p keys-cert
cd keys-cert
openssl ecparam -name prime256v1 -genkey -noout -out rekor_key.pem
cd ../
