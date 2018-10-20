#!/bin/bash
set -u -e -o pipefail -o xtrace

# This script re-generates self-signing certs in certs directory
# run as: ./scripts/mkcert.sh

# on mac must use password https://github.com/dotnet/corefx/issues/24225
# password hard-coded here and in NewSecuredServiceLocator() see --Security.Certificate.Passwor
RAVEN_Security_Certificate_Password=pwd1234

rm -rf ./certs
mkdir -p ./certs
cd ./certs

openssl genrsa -out ca.key 2048

openssl req -new -x509 -key ca.key -out ca.crt -subj "/C=US/ST=Arizona/L=Nevada/O=RavenDB Test CA/OU=RavenDB test CA/CN=a.javatest11.development.run/emailAddress=ravendbca@example.com"

openssl genrsa -out localhost.key 2048

openssl req -new  -key localhost.key -out localhost.csr -subj "/C=US/ST=Arizona/L=Nevada/O=RavenDB Test/OU=RavenDB test/CN=a.javatest11.development.run/emailAddress=ravendb@example.com"

openssl x509 -req -days 365 -extensions ext -extfile ../scripts/test_cert.conf -in localhost.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out localhost.crt

cat localhost.key localhost.crt > cert.pem

openssl pkcs12 -passout pass:${RAVEN_Security_Certificate_Password} -export -out server.pfx -inkey localhost.key -in localhost.crt

