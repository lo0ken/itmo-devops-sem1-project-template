#!/bin/bash
set -e

DOCKER_IMAGE="lo0ken/prices_backend"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_IMAGE="${DOCKER_IMAGE}:${IMAGE_TAG}"

if ! docker info > /dev/null 2>&1; then
    echo "Docker not running"
    exit 1
fi

echo "Building ${FULL_IMAGE}..."
docker build --platform linux/amd64 -t ${FULL_IMAGE} .

if ! docker images | grep -q "${DOCKER_IMAGE}"; then
    echo "Build failed"
    exit 1
fi

echo "Pushing to registry..."
docker push ${FULL_IMAGE}
echo "Done: ${FULL_IMAGE}"
