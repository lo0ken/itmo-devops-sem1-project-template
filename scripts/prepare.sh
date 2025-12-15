#!/bin/bash
set -e

echo "=== Подготовка окружения ==="

echo "Скачивание Go зависимостей..."
go mod download
echo "✓ Зависимости загружены"

echo "Проверка подключения к PostgreSQL..."

export PGPASSWORD=${POSTGRES_PASSWORD}
POSTGRES_HOST=${POSTGRES_HOST}
POSTGRES_USER=${POSTGRES_USER}
POSTGRES_DB=${POSTGRES_DB}

attempt=0
max_attempts=30

until psql -h "$POSTGRES_HOST" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -c '\q' 2>/dev/null; do
  attempt=$((attempt + 1))
  if [ $attempt -ge $max_attempts ]; then
    echo "✗ PostgreSQL недоступен после $max_attempts попыток"
    exit 1
  fi
  echo "PostgreSQL недоступен, ожидание... (попытка $attempt/$max_attempts)"
  sleep 1
done

echo "✓ PostgreSQL готов к работе"

echo "=== Подготовка завершена успешно ==="
