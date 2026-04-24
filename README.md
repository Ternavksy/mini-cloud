# mini-cloud

## 🚀 Общее описание
mini-cloud - это современное веб-приложение для управления файлами, построенное на Go с использованием архитектуры микросервисов и интегрированное с S3-совместимым хранилищем MinIO. Приложение предоставляет API для загрузки, скачивания, списка и удаления файлов, а также веб-интерфейс через консоль MinIO.

---

## 📦 Содержание проекта
```
mini-cloud/
├── Dockerfile
├── main.go
├── go.mod
├── go.sum
├── cmd/
│   └── main.go (для CLI)
├── internal/
│   ├── models/
│   ├── repository/
│   ├── server/
│   │   ├── handler.go
│   │   ├── router.go
│   │   └── service.go
│   └── storage/
│       ├── factory.go
│       ├── local_storage.go
│       ├── s3_storage.go
│       └── storage.go
├── storage_file/
│   ├── first_test.txt
│   └── test.txt
└── .env (пример в S3_SETUP.md)
```

---

## 🛠️ Установка и Установка зависимостей

### 1. Установка проекта
```bash
# Клонируйте репозиторий
git clone https://github.com/seuusername/mini-cloud.git && cd mini-cloud

# Установите зависимости
go mod tidy && go mod download
```

### 2. Настройка конфигурации
Скопируйте `.env.example` в `.env` и измените параметры:
```env
# Для локального хранилища
STORAGE_TYPE=local
LOCAL_STORAGE_PATH=./storage

# Или для MinIO
STORAGE_TYPE=s3
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET_NAME=files
```

> 🔐 Для продакшена рекомендуется сгенерировать новые ключи:
> ```bash
> go run main.go
> # Приложение создаст бакет при первом запуске
> ```

---

## 🐳 Настройка MinIO (для интеграции с S3-совместимым хранилищем)

### 1. Запуск через Docker
```bash
docker-compose up -d # Запускает MinIO и API
```
- **API**: `localhost:8080`
- **MinIO Web UI**: `localhost:9001`

### 2. Ручное создание MinIO
```bash
docker run -d -p 9000:9000 -p 9001:9001 \
   --name miniio minio/minio server /data
```
Создайте веб-интерфейс:
```bash
docker run -d -p 9001:9001 \
   minio/minio cmd admin --console-root /data
```

---

## 🌐 API Endpoints

### 1. Загрузка файла
```bash
curl -X POST -F "file=@/path/to/file.txt" http://localhost:8080/upload
```
**Ответ:**
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "name": "myfile.txt"
}
```

### 2. Скачивание файла
```bash
curl http://localhost:8080/download/123e4567-e89b-12d3-a456-426614174000 -O
```

### 3. Список файлов
```bash
curl http://localhost:8080/list
```
**Ответ:**
```json
{
  "files": [
    "123e4567-e89b-12d3-a456-426614174000",
    "456e7890-e89b-12d3-a456-426614174001"
  ]
}
```

### 4. Удаление файла
```bash
curl -X DELETE http://localhost:8080/delete/123e4567-e89b-12d3-a456-426614174000
```

---

## 🔄 Переключение между хранилищами

### Локальное хранилище
```env
STORAGE_TYPE=local
LOCAL_STORAGE_PATH=./storage # любой путь на файловой системе
```

### S3 (MinIO)
```env
STORAGE_TYPE=s3
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
S3_BUCKET_NAME=files
```

---

## 📦 Пример использования

### 1. Запускаем через docker-compose
```bash
docker-compose up -d

curl -X POST -F "file=@test.txt" http://localhost:8080/upload
curl http://localhost:8080/list
```

### 2. Через веб-консоль MinIO
1. Откройте браузер: `http://localhost:9001`
2. Войдите под учетными данными: `minioadmin` / `minioadmin`
3. Убедитесь, что создан бакет `files`
