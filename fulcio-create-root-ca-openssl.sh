#!/usr/bin/env sh

mkdir -p keys-cert

pushd keys-cert > /dev/null

cat << EOF > ca.cfg
[req]
default_bits = 2048
prompt = yes
x509_extensions = v3_ca
distinguished_name      = req_distinguished_name

[ req_distinguished_name ]
countryName                     = Country Name (2 letter code)
countryName_min                 = 2
countryName_max                 = 2
stateOrProvinceName             = State or Province Name (full name)
localityName                    = Locality Name (eg, city)
0.organizationName              = Organization Name (eg, company)
organizationalUnitName          = Organizational Unit Name (eg, section)
commonName                      = Common Name (eg, fully qualified host name)
commonName_max                  = 64
emailAddress                    = Email Address
emailAddress_max                = 64

[v3_ca]
basicConstraints = critical, CA:TRUE
keyUsage = keyCertSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always,issuer:always
EOF


openssl ecparam -genkey -name prime256v1 -noout -out unenc.key
openssl ec -in unenc.key -out file_ca_key.pem -des3
openssl ec -in file_ca_key.pem -pubout -out file_ca_pub.pem
openssl req -new -x509 -days 365 -key file_ca_key.pem -out fulcio-root.pem -config ca.cfg

rm unenc.key

popd > /dev/null
