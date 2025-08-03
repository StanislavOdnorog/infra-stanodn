#!ipxe

# Direct auto-boot - no user interaction
echo "Auto-booting ${PXE_SCRIPT}..."
chain http://${PXE_IP_ADDRESS}/ipxe/${PXE_SCRIPT}