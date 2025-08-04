#!/bin/sh

set -e

# Parse config
envsubst < /tmp/config.yaml > /etc/config.yaml

CONFIG_FILE=/etc/config.yaml
SHARE_DIR=/share

echo "🔍 Parsing $CONFIG_FILE and downloading missing files..."

IFS=$'\n'

# Получаем список образов
for os in $(yq -r e '.images | keys | .[]' "$CONFIG_FILE"); do
  os_dir="$SHARE_DIR/$os"
  mkdir -p "$os_dir"

  echo "📦 Processing OS: $os"

  # Обработка блока download
  for file in $(yq -r e ".images.$os.download | keys | .[]" "$CONFIG_FILE"); do
    url_template=$(yq -r e -o=json ".images.$os.download.$file" "$CONFIG_FILE" | envsubst)
    filename=$(basename "$url_template")
    target_path="$os_dir/$file"

    if [ -f "$target_path" ]; then
      echo "✅ $target_path already exists, skipping."
    else
      echo "⬇️  Downloading $url_template -> $target_path"
      curl -L "$url_template" -o "$target_path"
    fi
  done

  # Обработка блока script
  script_path="$SHARE_DIR/target.ipxe"
  yq -r e ".images.$os.script" "$CONFIG_FILE" | envsubst > "$script_path"
  echo "💾 Saved script to $script_path"
done

echo "✅ All images and scripts are ready."

# Генерация конфигурации dnsmasq
envsubst < /tmp/dnsmasq.conf.template > /etc/dnsmasq.conf

echo "🚀 Starting dnsmasq..."
dnsmasq -k --conf-file=/etc/dnsmasq.conf -d &

echo "🚀 Starting nginx..."
nginx -t
nginx -g "daemon off;"

