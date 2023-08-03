#!/usr/bin/env sh

mkdir -p rekor-key
cd rekor-key
openssl ecparam -name prime256v1 -genkey -noout -out key.pem
cd ../
