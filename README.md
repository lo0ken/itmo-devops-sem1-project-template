# Финальный проект 1 семестра

REST API сервис для загрузки и выгрузки данных о ценах. Сервис принимает сжатые CSV файлы (ZIP/TAR), сохраняет данные в PostgreSQL и предоставляет возможность экспорта с фильтрацией.

## Используемые технологии

- **Язык программирования**: Go 1.23.3
- **База данных**: PostgreSQL 13.10
- **Контейнеризация**: Docker, Docker Compose
- **Облачная платформа**: Yandex Cloud
- **CI/CD**: GitHub Actions
- **Container Registry**: Docker Hub
- **Библиотеки**:
  - `github.com/gorilla/mux` - HTTP маршрутизация
  - `github.com/lib/pq` - PostgreSQL драйвер

## Требования к системе

- **ОС**: Linux, macOS или Windows с WSL2
- **Для локального запуска**:
  - Docker 20.10+
  - Docker Compose 2.0+
- **Для развертывания в облаке**:
  - Yandex Cloud CLI
  - Доступ к Docker Hub
  - SSH ключи для доступа к VM

## API эндпоинты

### POST /api/v0/prices

Загрузка данных о ценах из сжатого CSV файла.

**Query параметры:**
- `type` (optional) - тип архива: `zip` или `tar`. По умолчанию: `zip`

**Body:**
- `multipart/form-data` с полем `file` содержащим архив

**Формат CSV:**
```csv
id,name,category,price,create_date
1,iPhone 13,Electronics,799.99,2024-01-01
```

**Ответ:**
```json
{
  "total_count": 100,
  "duplicates_count": 5,
  "total_items": 95,
  "total_categories": 10,
  "total_price": 15000.50
}
```

**Пример запроса:**
```bash
curl -X POST "http://localhost:8080/api/v0/prices?type=zip" \
  -F "file=@sample_data.zip"
```

### GET /api/v0/prices

Выгрузка данных о ценах в виде ZIP архива с CSV файлом.

**Query параметры (все опциональные):**
- `start` - начальная дата фильтрации (формат: YYYY-MM-DD)
- `end` - конечная дата фильтрации (формат: YYYY-MM-DD)
- `min` - минимальная цена
- `max` - максимальная цена

**Ответ:**
- ZIP архив содержащий `data.csv` с отфильтрованными данными

**Примеры запросов:**
```bash
# Получить все данные
curl -o prices.zip http://localhost:8080/api/v0/prices

# Получить данные за период
curl -o prices.zip "http://localhost:8080/api/v0/prices?start=2024-01-01&end=2024-12-31"

# Получить данные в ценовом диапазоне
curl -o prices.zip "http://localhost:8080/api/v0/prices?min=100&max=1000"

# Комбинированные фильтры
curl -o prices.zip "http://localhost:8080/api/v0/prices?start=2024-01-01&min=500&max=2000"
```

## Bash скрипты

### 1. prepare.sh - Подготовка Docker образа

Собирает Docker образ и публикует его в Docker Hub.

```bash
# Стандартный запуск (тег latest)
./scripts/prepare.sh

# С кастомным тегом
IMAGE_TAG=v1.0.0 ./scripts/prepare.sh
```

**Что делает:**
- Проверяет наличие Docker
- Собирает образ для платформы linux/amd64
- Публикует образ в Docker Hub (lo0ken/prices_backend)

**Требования:**
- Авторизация в Docker Hub: `docker login`

### 2. run.sh - Развертывание в Yandex Cloud

Создает виртуальную машину в Yandex Cloud и запускает приложение.

```bash
# Стандартный запуск
./scripts/run.sh

# С кастомным тегом образа
IMAGE_TAG=v1.0.0 ./scripts/run.sh
```

**Что делает:**
1. Проверяет наличие SSH ключей и .env файла
2. Создает/использует сеть и подсеть в Yandex Cloud
3. Создает виртуальную машину (Ubuntu 22.04)
4. Устанавливает Docker и Docker Compose на VM
5. Копирует docker-compose.production.yml и .env на VM
6. Загружает образ из Docker Hub
7. Запускает контейнеры
8. Проверяет доступность API
9. Сохраняет IP адрес в файл `.vm_ip`

**Требования:**
- Yandex Cloud CLI настроен и авторизован
- SSH ключи: `~/.ssh/id_rsa` и `~/.ssh/id_rsa.pub`
- Файл `.env` с переменными окружения
- Образ уже собран и опубликован (запустите prepare.sh)

### 3. tests.sh - Запуск тестов

Запускает тесты указанного уровня сложности.

```bash
# Уровень 1 - базовые тесты
./scripts/tests.sh 1

# Уровень 2 - продвинутые тесты
./scripts/tests.sh 2

# Уровень 3 - сложные тесты
./scripts/tests.sh 3
```

**Уровни тестирования:**
- **Уровень 1**: Базовая проверка POST/GET эндпоинтов, подключение к БД
- **Уровень 2**: Поддержка ZIP и TAR архивов, агрегатные запросы
- **Уровень 3**: Валидация данных, обработка дубликатов, фильтры по датам и ценам

**Требования:**
- Сервер запущен (локально или в облаке)
- Файл `.vm_ip` с IP адресом сервера (создается скриптом run.sh)

## Установка и запуск

### Локальный запуск

1. **Создайте файл .env:**
```bash
cat > .env << EOF
POSTGRES_DB=project-sem-1
POSTGRES_USER=validator
POSTGRES_PASSWORD=val1dat0r
POSTGRES_HOST=postgres
POSTGRES_PORT=5432
SERVER_PORT=8080
EOF
```

2. **Запустите через Docker Compose:**
```bash
docker-compose up -d
```

3. **Проверьте доступность:**
```bash
curl http://localhost:8080/api/v0/prices
```

4. **Остановка:**
```bash
docker-compose down
```

### Развертывание в Yandex Cloud

1. **Настройте Yandex Cloud CLI:**
```bash
yc init
```

2. **Создайте SSH ключи (если нет):**
```bash
ssh-keygen -t rsa -b 4096 -f ~/.ssh/id_rsa
```

3. **Авторизуйтесь в Docker Hub:**
```bash
docker login
```

4. **Создайте файл .env** (см. выше)

5. **Соберите и опубликуйте образ:**
```bash
./scripts/prepare.sh
```

6. **Разверните в облаке:**
```bash
./scripts/run.sh
```

7. **Проверьте статус:**
```bash
VM_IP=$(cat .vm_ip)
curl http://${VM_IP}:8080/api/v0/prices
```

## Конфигурация базы данных

PostgreSQL должен быть настроен со следующими параметрами:

- **Host**: localhost (или postgres в Docker)
- **Port**: 5432
- **Database**: project-sem-1
- **User**: validator
- **Password**: val1dat0r
- **Table**: prices

**Структура таблицы:**
```sql
CREATE TABLE prices (
    id INTEGER,
    name TEXT,
    category TEXT,
    price NUMERIC,
    create_date DATE
);
```

## CI/CD

Проект использует GitHub Actions для автоматического развертывания и тестирования.

**Workflow файл**: `.github/workflows/go_check.yaml`

**Необходимые GitHub Secrets:**
- `YC_SERVICE_ACCOUNT_KEY` - ключ сервисного аккаунта Yandex Cloud
- `YC_CLOUD_ID` - ID облака
- `YC_FOLDER_ID` - ID папки
- `POSTGRES_USER` - пользователь БД
- `POSTGRES_PASSWORD` - пароль БД
- `DOCKERHUB_USERNAME` - имя пользователя Docker Hub
- `DOCKERHUB_TOKEN` - токен доступа Docker Hub

**Процесс:**
1. Сборка и публикация образа в Docker Hub
2. Создание VM в Yandex Cloud
3. Развертывание приложения
4. Запуск тестов трех уровней
5. Удаление VM

## Тестовые данные

Директория `sample_data` содержит примеры CSV файлов:
```
sample_data/
└── data.csv
```

Пример архива: `sample_data.zip`

## Контакт

При возникновении вопросов обращайтесь к разработчику проекта.
