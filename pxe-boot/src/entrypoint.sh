#!/bin/sh

set -e

# =============================================================================
# Global Variables
# =============================================================================
readonly CONFIG_FILE=/etc/config.yaml
readonly SHARE_DIR=/share
readonly SCRIPT_NAME="$(basename "$0")"

# =============================================================================
# Utility Functions
# =============================================================================

log_info() {
    echo "ℹ️  [INFO] $*"
}

log_success() {
    echo "✅ [SUCCESS] $*"
}

log_error() {
    echo "❌ [ERROR] $*" >&2
}

log_warn() {
    echo "⚠️  [WARN] $*"
}

log_download() {
    echo "⬇️  [DOWNLOAD] $*"
}

log_write() {
    echo "📝 [WRITE] $*"
}

log_script() {
    echo "🎯 [SCRIPT] $*"
}

log_service() {
    echo "🚀 [SERVICE] $*"
}

# =============================================================================
# Configuration Functions
# =============================================================================

init_config() {
    log_info "Initializing configuration..."
    
    # Parse environment variables into config
    envsubst < /tmp/config.yaml > "$CONFIG_FILE"
    
    # Set IFS for safe iteration
    IFS=$'\n'
    
    log_success "Configuration initialized at $CONFIG_FILE"
}

validate_os_selection() {
    local os="$1"
    
    log_info "Validating OS selection: $os"
    
    if [ -z "$os" ]; then
        log_error "PXE_AUTO_OS environment variable not set!"
        exit 1
    fi
    
    if ! yq -e ".images.$os" "$CONFIG_FILE" > /dev/null 2>&1; then
        log_error "OS '$os' not found in configuration!"
        log_info "Available OS options:"
        yq -r e '.images | keys | .[]' "$CONFIG_FILE" | while read available_os; do
            echo "  - $available_os"
        done
        exit 1
    fi
    
    log_success "OS '$os' validated successfully"
}

# =============================================================================
# File Processing Functions
# =============================================================================

setup_os_directory() {
    local os="$1"
    local os_dir="$SHARE_DIR/$os"
    
    mkdir -p "$os_dir"
    echo "$os_dir"
}

download_files() {
    local os="$1"
    local os_dir="$2"
    
    log_info "Processing downloads for $os..."
    
    if ! yq -e ".images.$os.download" "$CONFIG_FILE" > /dev/null 2>&1; then
        log_info "No downloads configured for $os"
        return 0
    fi
    
    local download_count=0
    local skip_count=0
    
    # Create temporary file for download list to avoid subshell variable issues
    local temp_file=$(mktemp)
    yq eval ".images.\"$os\".download | to_entries | .[] | [.key, .value] | @csv" "$CONFIG_FILE" | tr -d '"' > "$temp_file"
    
    # Process each download
    while IFS=',' read -r file url_template; do
        [ -z "$file" ] && continue  # Skip empty lines
        
        url=$(echo "$url_template" | envsubst)
        target_path="$os_dir/$file"

        if [ -f "$target_path" ]; then
            log_info "$target_path already exists, skipping"
            skip_count=$((skip_count + 1))
        else
            log_download "$url -> $target_path"
            if curl -L --fail "$url" -o "$target_path"; then
                download_count=$((download_count + 1))
                log_success "Downloaded $file"
            else
                log_error "Failed to download $file from $url"
                rm -f "$target_path"
                rm -f "$temp_file"
                exit 1
            fi
        fi
    done < "$temp_file"
    
    rm -f "$temp_file"
    log_success "Downloads completed: $download_count new, $skip_count skipped"
}

write_config_files() {
    local os="$1"
    local os_dir="$2"
    
    log_info "Processing config files for $os..."
    
    if ! yq eval ".images.\"$os\".write" "$CONFIG_FILE" > /dev/null 2>&1; then
        log_info "No config files to write for $os"
        return 0
    fi
    
    local write_count=0
    
    for file in $(yq eval ".images.\"$os\".write | keys | .[]" "$CONFIG_FILE"); do
        local target_path="$os_dir/$file"
        log_write "Creating config file: $target_path"
        
        if yq eval ".images.\"$os\".write.\"$file\"" "$CONFIG_FILE" | envsubst > "$target_path"; then
            write_count=$((write_count + 1))
            log_success "Created $file"
        else
            log_error "Failed to create $file"
            exit 1
        fi
    done
    
    log_success "Config files created: $write_count"
}

generate_boot_script() {
    local os="$1"
    local os_dir="$2"
    
    log_info "Generating boot script for $os..."
    
    if ! yq -e ".images.$os.script" "$CONFIG_FILE" > /dev/null 2>&1; then
        log_warn "No boot script configured for $os"
        return 0
    fi
    
    local script_path="$os_dir/$os.ipxe"
    
    if yq -r e ".images.$os.script" "$CONFIG_FILE" | envsubst > "$script_path"; then
        log_success "Boot script saved: $script_path"
    else
        log_error "Failed to generate boot script for $os"
        exit 1
    fi
}

# =============================================================================
# PXE Setup Functions
# =============================================================================

generate_autoboot_script() {
    local os="$1"
    
    log_script "Generating autoboot script for OS: $os"
    
    cat > "$SHARE_DIR/target.ipxe" << EOF
#!ipxe

echo "Auto-booting ${os}..."
set base-url http://${PXE_IP_ADDRESS}
chain \${base-url}/${os}/${os}.ipxe || goto error

:error
echo "Failed to boot ${os}"
echo "Check container logs for details"
echo "Rebooting in 10 seconds..."
sleep 10
reboot
EOF
    
    log_success "Autoboot script saved: $SHARE_DIR/target.ipxe"
}

setup_dnsmasq() {
    log_service "Configuring dnsmasq..."
    
    if envsubst < /tmp/dnsmasq.conf.template > /etc/dnsmasq.conf; then
        log_success "dnsmasq configuration generated"
    else
        log_error "Failed to generate dnsmasq configuration"
        exit 1
    fi
}

# =============================================================================
# Service Management Functions
# =============================================================================

start_services() {
    log_service "Starting PXE services..."
    
    # Start dnsmasq in background
    log_service "Starting dnsmasq (DHCP/TFTP)..."
    if dnsmasq -k --conf-file=/etc/dnsmasq.conf -d ; then
        log_success "dnsmasq started in background"
        dnsmasq -k --conf-file=/etc/dnsmasq.conf -d &
        sleep 2  # Give dnsmasq time to start
    else
        log_error "Failed to start dnsmasq"
        exit 1
    fi
    
    # Test nginx configuration
    log_service "Testing nginx configuration..."
    if nginx -t; then
        log_success "nginx configuration is valid"
    else
        log_error "nginx configuration test failed"
        exit 1
    fi
    
    # Start nginx in foreground (main process)
    log_service "Starting nginx (HTTP server)..."
    exec nginx -g "daemon off;"
}

# =============================================================================
# Main Function
# =============================================================================

main() {
    local os="$PXE_AUTO_OS"
    
    log_info "Starting PXE Boot Server initialization..."
    log_info "Selected OS: $os"
    
    # Initialize configuration
    init_config
    
    # Validate OS selection
    validate_os_selection "$os"
    
    # Setup OS directory
    local os_dir
    os_dir=$(setup_os_directory "$os")
    
    # Process files for selected OS
    download_files "$os" "$os_dir"
    write_config_files "$os" "$os_dir"
    generate_boot_script "$os" "$os_dir"
    
    # Generate PXE boot scripts
    generate_autoboot_script "$os"
    
    # Setup services
    setup_dnsmasq
    
    log_success "All files and scripts ready for $os"
    
    # Start services (this will exec nginx and not return)
    start_services
}

# =============================================================================
# Entry Point
# =============================================================================

# Execute main function with all arguments
main "$@"

