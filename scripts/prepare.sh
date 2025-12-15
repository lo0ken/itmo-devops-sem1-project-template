#!/bin/bash
set -e

echo "=== Сборка Docker-образа ==="

IMAGE_NAME="prices-api"
IMAGE_TAG="latest"

echo "Сборка образа ${IMAGE_NAME}:${IMAGE_TAG}..."
docker build -t ${IMAGE_NAME}:${IMAGE_TAG} .

echo "✓ Docker-образ успешно собран"

if docker images | grep -q "${IMAGE_NAME}"; then
    echo "✓ Образ ${IMAGE_NAME}:${IMAGE_TAG} готов к использованию"
else
    echo "✗ Ошибка при создании образа"
    exit 1
fi

echo "=== Подготовка завершена успешно ==="
