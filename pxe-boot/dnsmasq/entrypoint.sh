#!/bin/sh

# Генерируем dnsmasq.conf из шаблона с переменными
envsubst < /tmp/dnsmasq.conf.template > /etc/dnsmasq.conf

# Генерируем default.ipxe из шаблона, подставляя PXE_IP_ADDRESS и PXE_SCRIPT
envsubst < /assets/ipxe/default.ipxe.tpl > /var/lib/tftpboot/default.ipxe

# Запускаем dnsmasq
exec dnsmasq -k --conf-file=/etc/dnsmasq.conf -d
