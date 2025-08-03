#!/bin/sh

# Generate dnsmasq.conf from template with variables
envsubst < /tmp/dnsmasq.conf.template > /etc/dnsmasq.conf

# Generate default.ipxe from template
envsubst < /assets/ipxe/default.ipxe.tpl > /assets/tftpboot/default.ipxe

# Start dnsmasq
exec dnsmasq -k --conf-file=/etc/dnsmasq.conf -d