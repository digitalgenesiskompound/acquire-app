#!/bin/bash

# Create certificates directory
mkdir -p certs

# Generate private key
openssl genrsa -out certs/server.key 2048

# Generate certificate signing request
openssl req -new -key certs/server.key -out certs/server.csr -subj "/C=US/ST=State/L=City/O=Organization/CN=10.0.20.10"

# Generate self-signed certificate valid for 365 days
openssl x509 -req -in certs/server.csr -signkey certs/server.key -out certs/server.crt -days 365

# Create a certificate bundle that includes Subject Alternative Names for multiple addresses
cat > certs/openssl.conf << EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = US
ST = State
L = City
O = Organization
CN = localhost

[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = 10.0.20.10
IP.3 = 0.0.0.0
EOF

# Generate new key and certificate with SAN
openssl genrsa -out certs/server.key 2048
openssl req -new -key certs/server.key -out certs/server.csr -config certs/openssl.conf
openssl x509 -req -in certs/server.csr -signkey certs/server.key -out certs/server.crt -days 365 -extensions v3_req -extfile certs/openssl.conf

# Clean up CSR
rm certs/server.csr certs/openssl.conf

echo "SSL certificates generated in ./certs/"
echo "server.key - Private key"
echo "server.crt - Certificate"
echo ""
echo "To use HTTPS, you'll need to:"
echo "1. Update your server to use these certificates"
echo "2. Trust the certificate in your browser (since it's self-signed)"
echo "3. Access via https://10.0.20.10:8443/ instead of http://10.0.20.10:8080/"
