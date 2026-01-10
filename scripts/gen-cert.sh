#!/bin/bash

CERT_DIR="./certs"
mkdir -p $CERT_DIR

echo "Generating TLS certificates..."


openssl genrsa -out $CERT_DIR/ca-key.pem 4096

openssl req -new -x509 -days 3650 -key $CERT_DIR/ca-key.pem \
    -out $CERT_DIR/ca-cert.pem \
    -subj "/C=US/ST=State/L=City/O=GoTunnel/OU=CA/CN=GoTunnel CA"


openssl genrsa -out $CERT_DIR/server-key.pem 4096

openssl req -new -key $CERT_DIR/server-key.pem \
    -out $CERT_DIR/server-csr.pem \
    -subj "/C=US/ST=State/L=City/O=GoTunnel/OU=Server/CN=localhost"

cat > $CERT_DIR/server-ext.cnf << EOF
subjectAltName = DNS:localhost,DNS:*.localhost,IP:127.0.0.1
extendedKeyUsage = serverAuth
EOF

openssl x509 -req -days 3650 \
    -in $CERT_DIR/server-csr.pem \
    -CA $CERT_DIR/ca-cert.pem \
    -CAkey $CERT_DIR/ca-key.pem \
    -CAcreateserial \
    -out $CERT_DIR/server-cert.pem \
    -extfile $CERT_DIR/server-ext.cnf

rm $CERT_DIR/server-csr.pem $CERT_DIR/server-ext.cnf $CERT_DIR/ca-cert.srl

echo "✓ Certificates generated in $CERT_DIR/"
echo ""
echo "Files created:"
echo "  - ca-cert.pem       (CA certificate - for client)"
echo "  - server-cert.pem   (Server certificate)"
echo "  - server-key.pem    (Server private key)"
echo ""
echo "Server usage: --tls-cert=certs/server-cert.pem --tls-key=certs/server-key.pem"
echo "Client usage: --tls-ca=certs/ca-cert.pem"