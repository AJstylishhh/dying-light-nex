#!/bin/bash
mkdir -p certs
openssl req -x509 -newkey rsa:2048 \
  -keyout certs/server.key \
  -out certs/server.crt \
  -days 365 -nodes \
  -subj "/CN=g27332e00-lp1.s.n.srv.nintendo.net"
echo "Created certs/server.crt and certs/server.key"
