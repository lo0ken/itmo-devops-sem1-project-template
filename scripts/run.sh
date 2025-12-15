#!/bin/bash
set -e

echo "=== Запуск приложения ==="

echo "Компиляция приложения..."
go build -o app main.go
echo "✓ Компиляция завершена"

echo "Запуск приложения в фоновом режиме..."
./app &
APP_PID=$!
echo "✓ Приложение запущено с PID: $APP_PID"

echo "Ожидание готовности API..."

attempt=0
max_attempts=30

for i in $(seq 1 $max_attempts); do
  if curl -s -o /dev/null http://localhost:8080/api/v0/prices 2>/dev/null; then
    echo "✓ API готов к работе!"
    echo "=== Приложение успешно запущено ==="
    exit 0
  fi

  if ! kill -0 $APP_PID 2>/dev/null; then
    echo "✗ Приложение завершилось с ошибкой"
    exit 1
  fi

  echo "API недоступен, ожидание... (попытка $i/$max_attempts)"
  sleep 1
done

echo "✗ API не стал доступен за $max_attempts секунд"
echo "Остановка приложения..."
kill $APP_PID 2>/dev/null || true
exit 1
