#!/bin/bash

container_id="$(cat /etc/hostname)"

# Check if the "k3d-openchoreo-local" network exists and connect the container to it
if docker network inspect k3d-openchoreo-local &>/dev/null; then
  # Check if the container is already connected
  if [ "$(docker inspect -f '{{json .NetworkSettings.Networks.k3d-openchoreo-local}}' "${container_id}")" = "null" ]; then
    docker network connect "k3d-openchoreo-local" "${container_id}"
    echo "Connected container ${container_id} to k3d-openchoreo-local network."
  else
    echo "Container ${container_id} is already connected to k3d-openchoreo-local network."
  fi
fi

# Fix kubeconfig to use k3d cluster's internal network IP instead of 127.0.0.1
if k3d cluster list 2>/dev/null | grep -q "openchoreo-local"; then
  CONTROL_PLANE_IP=$(docker inspect k3d-openchoreo-local-server-0 --format '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' 2>/dev/null | head -1)
  if [ -n "$CONTROL_PLANE_IP" ]; then
    echo "Configuring kubectl to connect to k3d cluster at ${CONTROL_PLANE_IP}..."
    mkdir -p /state/kube
    k3d kubeconfig get openchoreo-local 2>/dev/null | sed "s|server: https://[^:]*:[0-9]*|server: https://${CONTROL_PLANE_IP}:6443|" > /state/kube/config-internal.yaml
    export KUBECONFIG=/state/kube/config-internal.yaml
    echo "✓ kubectl configured successfully"
  fi
fi

exec /bin/bash -l
