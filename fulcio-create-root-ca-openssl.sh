#!/usr/bin/env sh

mkdir -p fulcio-root
cd fulcio-root

openssl ecparam -genkey -name prime256v1 -noout -out unenc.key
openssl ec -in unenc.key -out file_ca_key.pem -des3
openssl ec -in file_ca_key.pem -pubout -out file_ca_pub.pem
openssl req -new -x509 -days 365 -extensions v3_ca -key file_ca_key.pem -out fulcio-root.pem

rm unenc.key
cd ../
