#!/bin/sh

# Generate dnsmasq.conf from template with variables
envsubst < /tmp/dnsmasq.conf.template > /etc/dnsmasq.conf

# Start dnsmasq
exec dnsmasq -k --conf-file=/etc/dnsmasq.conf -d