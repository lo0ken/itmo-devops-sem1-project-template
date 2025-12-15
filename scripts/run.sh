#!/bin/bash
set -e

echo "=== Создание и настройка сервера в Yandex Cloud ==="

VM_NAME="prices-api-vm-$(date +%s)"
ZONE="ru-central1-a"
NETWORK_NAME="prices-api-network"
SUBNET_NAME="prices-api-subnet-a"
SUBNET_RANGE="10.128.0.0/24"
IMAGE_FAMILY="ubuntu-2204-lts"
SSH_USER="ubuntu"
SSH_KEY_PATH="${HOME}/.ssh/id_rsa"

if [ ! -f "${SSH_KEY_PATH}.pub" ]; then
    echo "✗ SSH ключ не найден: ${SSH_KEY_PATH}.pub"
    echo "Создайте SSH ключ с помощью команды:"
    echo "  ssh-keygen -t rsa -b 4096 -f ${SSH_KEY_PATH}"
    exit 1
fi

echo "Проверка наличия сети..."
NETWORK_ID=$(yc vpc network list --format json | jq -r ".[] | select(.name==\"${NETWORK_NAME}\") | .id")

if [ -z "$NETWORK_ID" ]; then
    echo "Сеть не найдена, создаём новую сеть ${NETWORK_NAME}..."
    NETWORK_ID=$(yc vpc network create \
        --name ${NETWORK_NAME} \
        --description "Network for prices API service" \
        --format json | jq -r '.id')
    echo "✓ Сеть создана: ${NETWORK_ID}"
else
    echo "✓ Сеть уже существует: ${NETWORK_ID}"
fi

echo "Проверка наличия подсети..."
SUBNET_ID=$(yc vpc subnet list --format json | jq -r ".[] | select(.name==\"${SUBNET_NAME}\") | .id")

if [ -z "$SUBNET_ID" ]; then
    echo "Подсеть не найдена, создаём новую подсеть ${SUBNET_NAME}..."
    SUBNET_ID=$(yc vpc subnet create \
        --name ${SUBNET_NAME} \
        --zone ${ZONE} \
        --network-id ${NETWORK_ID} \
        --range ${SUBNET_RANGE} \
        --format json | jq -r '.id')
    echo "✓ Подсеть создана: ${SUBNET_ID}"
else
    echo "✓ Подсеть уже существует: ${SUBNET_ID}"
fi

echo "Создание виртуальной машины ${VM_NAME}..."

# Попытка создания VM с retry при перегрузке
MAX_RETRIES=5
RETRY_DELAY=10
VM_ID=""

for retry in $(seq 1 $MAX_RETRIES); do
    echo "Попытка создания VM $retry/$MAX_RETRIES..."

    VM_CREATE_OUTPUT=$(yc compute instance create \
        --name ${VM_NAME} \
        --zone ${ZONE} \
        --network-interface subnet-name=${SUBNET_NAME},nat-ip-version=ipv4 \
        --create-boot-disk image-folder-id=standard-images,image-family=${IMAGE_FAMILY},size=20 \
        --memory 2G \
        --cores 2 \
        --ssh-key "${SSH_KEY_PATH}.pub" \
        --format json 2>&1)

    VM_ID=$(echo "$VM_CREATE_OUTPUT" | jq -r '.id' 2>/dev/null)

    if [ -n "$VM_ID" ] && [ "$VM_ID" != "null" ]; then
        echo "✓ Виртуальная машина создана: ${VM_ID}"
        break
    fi

    if echo "$VM_CREATE_OUTPUT" | grep -q "ResourceExhausted"; then
        if [ $retry -lt $MAX_RETRIES ]; then
            echo "⚠ Сервис перегружен, повторная попытка через ${RETRY_DELAY} секунд..."
            sleep $RETRY_DELAY
        else
            echo "✗ Не удалось создать VM после $MAX_RETRIES попыток"
            echo "Ошибка: $VM_CREATE_OUTPUT"
            exit 1
        fi
    else
        echo "✗ Ошибка создания виртуальной машины:"
        echo "$VM_CREATE_OUTPUT"
        exit 1
    fi
done

echo "Получение IP-адреса..."
sleep 10

VM_IP=$(yc compute instance get ${VM_ID} --format json | jq -r '.network_interfaces[0].primary_v4_address.one_to_one_nat.address')

if [ -z "$VM_IP" ] || [ "$VM_IP" = "null" ]; then
    echo "✗ Не удалось получить IP-адрес"
    exit 1
fi

echo "✓ IP-адрес получен: ${VM_IP}"

echo "Ожидание готовности SSH..."
max_attempts=60
attempt=0

while ! ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 ${SSH_USER}@${VM_IP} "echo 'SSH ready'" 2>/dev/null; do
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
        echo "✗ SSH не стал доступен"
        exit 1
    fi
    echo "SSH недоступен, ожидание... (попытка $attempt/$max_attempts)"
    sleep 5
done

echo "✓ SSH подключение готово"

echo "Установка Docker и Docker Compose..."
ssh -o StrictHostKeyChecking=no ${SSH_USER}@${VM_IP} << 'ENDSSH'
    set -e

    sudo apt-get update

    sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common

    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

    sudo apt-get update
    sudo apt-get install -y docker-ce docker-ce-cli containerd.io

    sudo usermod -aG docker $USER

    sudo curl -L "https://github.com/docker/compose/releases/download/v2.20.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
    sudo chmod +x /usr/local/bin/docker-compose

    sudo systemctl start docker
    sudo systemctl enable docker
ENDSSH

echo "✓ Docker и Docker Compose установлены"

echo "Копирование файлов проекта..."
ssh -o StrictHostKeyChecking=no ${SSH_USER}@${VM_IP} "mkdir -p ~/app"
scp -o StrictHostKeyChecking=no -r ./* ${SSH_USER}@${VM_IP}:~/app/

echo "✓ Файлы проекта скопированы"

echo "Запуск приложения через Docker Compose..."
ssh -o StrictHostKeyChecking=no ${SSH_USER}@${VM_IP} << 'ENDSSH'
    set -e
    cd ~/app

    # Запускаем Docker Compose (нужно использовать sudo для новой сессии)
    sudo docker-compose up -d --build

    # Проверяем статус контейнеров
    sudo docker-compose ps
ENDSSH

echo "✓ Приложение запущено"

echo "Ожидание готовности API..."
max_attempts=60
attempt=0

while ! curl -s -o /dev/null http://${VM_IP}:8080/api/v0/prices 2>/dev/null; do
    attempt=$((attempt + 1))
    if [ $attempt -ge $max_attempts ]; then
        echo "✗ API не стал доступен"
        exit 1
    fi
    echo "API недоступен, ожидание... (попытка $attempt/$max_attempts)"
    sleep 5
done

echo "✓ API готов к работе"

echo "${VM_IP}" > .vm_ip

echo ""
echo "=== Развёртывание завершено успешно ==="
echo "IP-адрес сервера: ${VM_IP}"
echo "API доступен по адресу: http://${VM_IP}:8080"
echo ""
