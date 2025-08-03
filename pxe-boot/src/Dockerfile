FROM alpine:3.19 AS builder

# Accept build arguments
ARG PXE_SCRIPT=parrot-live.ipxe

# Install all dependencies for iPXE building, nginx, and dnsmasq
RUN apk add --no-cache \
    git \
    gcc \
    musl-dev \
    linux-headers \
    libc-dev \
    binutils \
    make \
    perl \
    xz-dev \
    mtools \
    cdrkit \
    syslinux \
    nginx \
    dnsmasq \
    gettext \
    supervisor

# Build iPXE binaries
WORKDIR /build
RUN git clone https://github.com/ipxe/ipxe.git

# Copy our embedded script
COPY ./build-ipxe/ipxe/${PXE_SCRIPT} /build/boot.ipxe

# Build iPXE with embedded script
WORKDIR /build/ipxe/src
RUN make bin/undionly.kpxe EMBED=/build/boot.ipxe
RUN make bin-x86_64-efi/ipxe.efi EMBED=/build/boot.ipxe

# Create directory structure
RUN mkdir -p /assets/tftpboot
RUN cp bin/undionly.kpxe /assets/tftpboot/
RUN cp bin-x86_64-efi/ipxe.efi /assets/tftpboot/

FROM alpine:3.19

COPY --from=builder /assets/tftpboot/ /assets/tftpboot/

# Setup nginx
RUN mkdir -p /run/nginx
COPY ./nginx/nginx.conf /etc/nginx/nginx.conf

# Setup dnsmasq
COPY ./dnsmasq/dnsmasq.conf.template /tmp/dnsmasq.conf.template

# Setup supervisor to manage multiple processes
COPY supervisord.conf /etc/supervisord.conf

# Create entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Expose ports
EXPOSE 53/udp 67/udp 69/udp 80

# Set working directory
WORKDIR /

ENTRYPOINT ["/entrypoint.sh"]
CMD ["supervisord", "-c", "/etc/supervisord.conf", "-n"]