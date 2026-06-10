# mini-cloud

## 🚀 Быстрая демонстрация API
Ниже набор команд для демонстрации API от начала до конца.

### 1. Запустить сервер с локальным хранилищем
```bash
cd /Users/larisa/Downloads/mini-cloud
make local
```

Метаданные коллекций хранятся в SQLite. По умолчанию файл БД создаётся как `metadata.db`; путь можно изменить:
```bash
METADATA_DB_PATH=./metadata.db make local
```

### 2. Проверить список файлов
```bash
curl http://localhost:8080/files
```

### 3. Загрузить файл
```bash
curl -X POST -F "file=@README.md" http://localhost:8080/upload
```

### 4. Снова посмотреть список файлов
```bash
curl http://localhost:8080/files
```

Дальше возьмите id из ответа upload. Например:

### 5. Скачать файл по id
```bash
curl http://localhost:8080/download/fb105076-1416-4be2-a043-580a29e7977b -o downloaded_README.md
```

### 6. Удалить файл по id
```bash
curl -X DELETE http://localhost:8080/files/fb105076-1416-4be2-a043-580a29e7977b
```

### 7. Проверить, что файл удалился
```bash
curl http://localhost:8080/files
```

## 📁 Коллекции файлов
Коллекции работают как альбомы: в них хранится список id файлов.

У JSON-ручек единый формат ответа:
```json
{"data": {"id": "collection-id"}}
```

Ошибки возвращаются так:
```json
{"error": {"message": "collection not found"}}
```

### Создать коллекцию
```bash
curl -X POST http://localhost:8080/collections \
  -H "Content-Type: application/json" \
  -d '{"name":"album","file_ids":["fb105076-1416-4be2-a043-580a29e7977b"]}'
```

### Получить список коллекций
```bash
curl http://localhost:8080/collections
```

### Получить коллекцию по id
```bash
curl http://localhost:8080/collections/collection-id
```

### Добавить файл в коллекцию
```bash
curl -X POST http://localhost:8080/collections/collection-id/files/file-id
```

### Удалить файл из коллекции
```bash
curl -X DELETE http://localhost:8080/collections/collection-id/files/file-id
```

### Удалить коллекцию
```bash
curl -X DELETE http://localhost:8080/collections/collection-id
```

## 🧬 Генерация кода
DTO для request/response генерируются из OpenAPI-спеки `api/openapi.yaml`.
Моки для тестов генерируются через `go generate`.

```bash
make generate
go test ./...
```

## 🧪 Демонстрация через MinIO
Если хотите запустить демонстрацию через MinIO, сначала запустите:

```bash
docker-compose up -d
```

Потом сервер:

```bash
S3_ENDPOINT=localhost:9000 \
S3_ACCESS_KEY=minioadmin \
S3_SECRET_KEY=minioadmin \
S3_BUCKET_NAME=files \
STORAGE_TYPE=s3 \
go run main.go server
```

### MinIO UI
- URL: `http://localhost:9001/login`
- login: `minioadmin`
- password: `minioadmin`
