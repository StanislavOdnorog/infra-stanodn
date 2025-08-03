FROM alpine:3.19

RUN apk add --no-cache dnsmasq gettext

COPY dnsmasq.conf.template /tmp/dnsmasq.conf.template
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]