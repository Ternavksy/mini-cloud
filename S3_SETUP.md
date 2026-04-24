# Интеграция S3 с MinIO

Это руководство описывает как настроить и использовать MinIO (S3-совместимое хранилище) для вашего приложения.

## Что такое MinIO?

**MinIO** — это бесплатное, открытое S3-совместимое хранилище объектов. Оно позволяет вам использовать S3 API локально без необходимости платить за AWS.

## Быстрый старт

### 1. Запустить MinIO в Docker

```bash
docker-compose up -d
```

Это запустит MinIO на `localhost:9000` (API) и `localhost:9001` (Web UI).

### 2. Доступ к MinIO Console

Откройте браузер и перейдите на: http://localhost:9001

**Учетные данные:**
- Username: `minioadmin`
- Password: `minioadmin`

### 3. Создать .env файл

Скопируйте `.env.example` в `.env`:

```bash
cp .env.example .env
```

Или вручную создайте `.env` с переменными:

```env
STORAGE_TYPE=s3
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET_NAME=files
```

### 4. Загрузить dependencies

```bash
go mod download
go mod tidy
```

### 5. Запустить приложение

```bash
go run main.go
```

Приложение автоматически создаст бакет `files` при первом запуске.

## API Endpoints

### Загрузить файл
```bash
curl -X POST -F "file=@путь/к/файлу" http://localhost:8080/upload
```

Ответ:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "myfile.txt"
}
```

### Скачать файл
```bash
curl http://localhost:8080/download/123e4567-e89b-12d3-a456-426614174000 -O
```

### Список файлов
```bash
curl http://localhost:8080/list
```

Ответ:
```json
{
  "files": [
    "123e4567-e89b-12d3-a456-426614174000",
    "456e7890-e89b-12d3-a456-426614174001"
  ]
}
```

### Удалить файл
```bash
curl -X DELETE http://localhost:8080/delete/123e4567-e89b-12d3-a456-426614174000
```

## Переключение между локальным и S3 хранилищем

Измените переменную `STORAGE_TYPE` в `.env`:

### Использование локального хранилища
```env
STORAGE_TYPE=local
LOCAL_STORAGE_PATH=./storage
```

### Использование S3 (MinIO)
```env
STORAGE_TYPE=s3
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET_NAME=files
```

## Развертывание в продакшене

### С собственным MinIO сервером

```bash
docker-compose -f docker-compose.prod.yml up -d
```

### С AWS S3

Просто обновите переменные окружения:

```env
STORAGE_TYPE=s3
S3_ENDPOINT=s3.amazonaws.com
S3_ACCESS_KEY=ваш_aws_access_key
S3_SECRET_KEY=ваш_aws_secret_key
S3_BUCKET_NAME=ваш_bucket_name
```

## Управление данными в MinIO Console

1. Откройте http://localhost:9001
2. Войдите с `minioadmin` / `minioadmin`
3. Перейдите в раздел "Object Browser"
4. Управляйте файлами (загрузка, скачивание, удаление)

## Остановить MinIO

```bash
docker-compose down
```

Для удаления данных также:
```bash
docker-compose down -v
```

## Решение проблем

### MinIO не запускается
```bash
# Проверить логи
docker-compose logs minio

# Убедиться, что порты свободны
lsof -i :9000
lsof -i :9001
```

### Подключение не работает
- Проверить, запущен ли MinIO: `docker ps | grep minio`
- Проверить .env переменные
- Проверить, доступен ли `localhost:9000` из приложения

### Бакет не создается автоматически
Создайте вручную через MinIO Console или используйте AWS CLI:

```bash
aws s3api create-bucket --bucket files --endpoint-url http://localhost:9000 \
  --region us-east-1 \
  --access-key minioadmin \
  --secret-key minioadmin
```
