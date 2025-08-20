FROM alpine:3.19

# Install all dependencies for iPXE building, nginx, and dnsmasq
RUN apk add --no-cache \
    curl \
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
    xorriso \
    syslinux \
    nginx \
    dnsmasq \
    gettext \
    yq

# Build iPXE binaries
WORKDIR /build
RUN git clone https://github.com/ipxe/ipxe.git

# Copy our embedded script
COPY ./ipxe/default.ipxe /tmp/boot.ipxe
ARG PXE_IP_ADDRESS
RUN envsubst < /tmp/boot.ipxe > /build/boot.ipxe

# Build iPXE with embedded script
WORKDIR /build/ipxe/src
RUN make bin/undionly.kpxe EMBED=/build/boot.ipxe
RUN make bin-x86_64-efi/ipxe.efi EMBED=/build/boot.ipxe

# Create directory structure
RUN mkdir -p /assets/tftpboot
RUN cp bin/undionly.kpxe /assets/tftpboot/
RUN cp bin-x86_64-efi/ipxe.efi /assets/tftpboot/

# Setup nginx
RUN mkdir -p /run/nginx
COPY ./nginx/nginx.conf /etc/nginx/nginx.conf

# Setup dnsmasq
COPY ./dnsmasq/dnsmasq.conf.template /tmp/dnsmasq.conf.template

# Create entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Copy custom ipxe scripts
RUN mkdir /share -p

# Expose ports
EXPOSE 53/udp 67/udp 69/udp 80

# Set working directory
WORKDIR /

# Copy config file
COPY ./config.yaml /tmp/config.yaml

ENTRYPOINT ["/entrypoint.sh"]
