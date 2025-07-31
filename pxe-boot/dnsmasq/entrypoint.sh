#!/bin/sh

envsubst < /tmp/dnsmasq.conf.template > /etc/dnsmasq.conf

exec dnsmasq -k --conf-file=/etc/dnsmasq.conf -d
