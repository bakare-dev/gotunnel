#!/usr/bin/env bash

mkdir -p certs

openssl req -x509 -newkey rsa:4096 \
  -keyout certs/server.key \
  -out certs/server.crt \
  -days 365 \
  -nodes \
  -subj "/CN=gotunnel"
