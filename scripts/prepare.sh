#!/bin/bash
set -e

echo "=== Сборка и публикация Docker-образа ==="

DOCKER_IMAGE="lo0ken/prices_backend"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_IMAGE="${DOCKER_IMAGE}:${IMAGE_TAG}"

# Проверка Docker
if ! docker info > /dev/null 2>&1; then
    echo "✗ Docker не запущен или недоступен"
    exit 1
fi

echo "Сборка образа ${FULL_IMAGE} для платформы linux/amd64..."
docker build --platform linux/amd64 -t ${FULL_IMAGE} .

echo "✓ Docker-образ успешно собран"

if docker images | grep -q "${DOCKER_IMAGE}"; then
    echo "✓ Образ ${FULL_IMAGE} готов к использованию"
else
    echo "✗ Ошибка при создании образа"
    exit 1
fi

echo "Пуш образа в Docker Registry..."
echo "Примечание: Убедитесь, что вы авторизованы в Docker Hub (docker login)"
docker push ${FULL_IMAGE}
echo "✓ Образ загружен в registry: ${FULL_IMAGE}"

echo "=== Подготовка завершена успешно ==="
