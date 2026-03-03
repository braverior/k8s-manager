#!/bin/sh
# Generate config.js from environment variable K8S_MANAGER_CLUSTERS
# This runs at container startup to inject runtime configuration

CONFIG_FILE="/usr/share/nginx/html/config.js"

if [ -n "$K8S_MANAGER_CLUSTERS" ]; then
  echo "window.__K8S_CLUSTERS__ = ${K8S_MANAGER_CLUSTERS};" > "$CONFIG_FILE"
  echo "[entrypoint] Generated config.js with K8S_MANAGER_CLUSTERS"
else
  echo "// No K8S_MANAGER_CLUSTERS configured" > "$CONFIG_FILE"
  echo "[entrypoint] WARNING: K8S_MANAGER_CLUSTERS not set"
fi

# Start nginx
exec nginx -g "daemon off;"
