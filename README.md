# Subscription Service

REST API сервис для управления подписками пользователей.

## 📋 Описание

Сервис предоставляет REST API для создания, чтения, обновления и удаления (CRUD) записей о подписках пользователей на различные сервисы. Также поддерживает расчет суммарной стоимости подписок за выбранный период с возможностью фильтрации.

## 🚀 Возможности

- ✅ CRUD операции над подписками
- ✅ Расчет суммарной стоимости подписок за период
- ✅ Фильтрация по ID пользователя и названию сервиса
- ✅ Пагинация списка подписок
- ✅ PostgreSQL с миграциями
- ✅ Структурированное логирование
- ✅ Swagger документация
- ✅ Docker поддержка
- ✅ Graceful shutdown

## 🛠 Технологии

- **Go** 1.21+
- **PostgreSQL** 16+
- **Docker** & Docker Compose
- **Chi** - HTTP роутер
- **pgx/v5** - PostgreSQL драйвер
- **slog** - структурированное логирование
- **Swagger** - API документация

## 📦 Установка и запуск

### Требования

- Go 1.21 или выше
- Docker & Docker Compose
- Make (опционально)

### Быстрый старт с Docker Compose

1. **Клонируйте репозиторий:**
```bash
git clone https://github.com/yourusername/subscription-service.git
cd subscription-service
```

2. **Создайте .env файл:**
```bash
cp .env.example .env
```

3. **Запустите сервисы:**
```bash
docker-compose up -d
```

4. **Проверьте статус:**
```bash
docker-compose ps
```

Сервис будет доступен по адресу: `http://localhost:8080`

### Запуск без Docker (локальная разработка)

1. **Установите зависимости:**
```bash
go mod download
```

2. **Запустите PostgreSQL:**
```bash
docker run -d \
  --name postgres \
  -e POSTGRES_USER=subscriptions \
  -e POSTGRES_PASSWORD=subscriptions_password \
  -e POSTGRES_DB=subscriptions_db \
  -p 5432:5432 \
  postgres:15-alpine
```

3. **Создайте .env файл и настройте параметры:**
```bash
cp .env.example .env
# Отредактируйте .env если нужно
```

4. **Запустите приложение:**
```bash
go run cmd/api/main.go
```

## 🎯 API Endpoints

### Subscriptions

| Метод | Endpoint | Описание |
|-------|----------|----------|
| POST | `/api/v1/subscriptions` | Создать подписку |
| GET | `/api/v1/subscriptions` | Получить список подписок |
| GET | `/api/v1/subscriptions/{id}` | Получить подписку по ID |
| PUT | `/api/v1/subscriptions/{id}` | Обновить подписку |
| DELETE | `/api/v1/subscriptions/{id}` | Удалить подписку |
| GET | `/api/v1/subscriptions/cost` | Рассчитать стоимость подписок |

### Другие endpoints

| Метод | Endpoint | Описание |
|-------|----------|----------|
| GET | `/health` | Health check |
| GET | `/swagger/*` | Swagger UI |

## 📖 Примеры использования

### Создание подписки
```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "cost": 299,
    "start_date": "07-2025"
  }'
```

**Ответ:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "service_name": "Yandex Plus",
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "cost": 299,
  "start_date": "07-2025",
  "end_date": null,
  "created_at": "2025-02-16T10:30:00Z",
  "updated_at": "2025-02-16T10:30:00Z"
}
```

### Получение списка подписок
```bash
curl http://localhost:8080/api/v1/subscriptions?page=1&page_size=10
```

### Получение подписки по ID
```bash
curl http://localhost:8080/api/v1/subscriptions/550e8400-e29b-41d4-a716-446655440000
```

### Обновление подписки
```bash
curl -X PUT http://localhost:8080/api/v1/subscriptions/550e8400-e29b-41d4-a716-446655440000 \
  -H "Content-Type: application/json" \
  -d '{
    "cost": 399,
    "end_date": "12-2025"
  }'
```

### Удаление подписки
```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/550e8400-e29b-41d4-a716-446655440000
```

### Расчет стоимости подписок
```bash
# Все подписки за период
curl "http://localhost:8080/api/v1/subscriptions/cost?start_date=01-2025&end_date=12-2025"

# С фильтром по пользователю
curl "http://localhost:8080/api/v1/subscriptions/cost?start_date=01-2025&end_date=12-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"

# С фильтром по сервису
curl "http://localhost:8080/api/v1/subscriptions/cost?start_date=01-2025&end_date=12-2025&service_name=Yandex%20Plus"
```

**Ответ:**
```json
{
  "total_cost": 5980,
  "period": "01-2025 to 12-2025",
  "count": 20
}
```

## 📚 Swagger документация

После запуска сервиса Swagger UI доступен по адресу:
```
http://localhost:8080/swagger/index.html
```

Также доступна YAML спецификация:
```
http://localhost:8080/swagger/doc.json
```

## 🗂 Структура проекта
```
subscription-service/
├── cmd/
│   └── api/
│       └── main.go              # Точка входа приложения
├── internal/
│   ├── config/
│   │   └── config.go           # Конфигурация
│   ├── domain/
│   │   └── subscription.go     # Доменные модели
│   ├── handler/
│   │   ├── router.go           # HTTP роутер
│   │   └── subscription.go     # HTTP handlers
│   ├── repository/
│   │   └── subscription.go     # Репозиторий БД
│   └── service/
│       └── subscription.go     # Бизнес-логика
├── pkg/
│   ├── logger/
│   │   └── logger.go           # Логгер
│   └── validator/
│       └── validator.go        # Валидация
├── migrations/
│   └── 001_init.sql            # SQL миграции
├── docs/
│   ├── docs.go                 # Swagger docs
│   └── swagger.yaml            # OpenAPI спецификация
├── .env.example                # Пример конфигурации
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
└── README.md
```

## 🔧 Makefile команды
```bash
make help           # Показать доступные команды
make build          # Собрать приложение
make run            # Запустить локально
make test           # Запустить тесты
make docker-up      # Запустить через docker-compose
make docker-down    # Остановить docker-compose
make docker-logs    # Показать логи
make swagger        # Сгенерировать Swagger документацию
make clean          # Очистить артефакты сборки
```

## 🔐 Переменные окружения

Основные переменные окружения (см. `.env.example`):

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| SERVER_HOST | Хост сервера | 0.0.0.0 |
| SERVER_PORT | Порт сервера | 8080 |
| DB_HOST | Хост PostgreSQL | postgres |
| DB_PORT | Порт PostgreSQL | 5432 |
| DB_USER | Пользователь БД | subscriptions |
| DB_PASSWORD | Пароль БД | subscriptions_password |
| DB_NAME | Имя БД | subscriptions_db |
| LOG_LEVEL | Уровень логирования | info |

## 🧪 Тестирование
```bash
# Запустить тесты
make test

# Запустить тесты с покрытием
make coverage
```

## 📝 Логирование

Сервис использует структурированное логирование (slog) с JSON форматом:
```json
{
  "time": "2025-02-16T10:30:00Z",
  "level": "INFO",
  "msg": "subscription created successfully",
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Уровни логирования: `debug`, `info`, `warn`, `error`

## 🐳 Docker

### Сборка образа
```bash
docker build -t subscription-service:latest .
```

### Запуск контейнера
```bash
docker run -p 8080:8080 subscription-service:latest
```

## 🔍 Health Check
```bash
curl http://localhost:8080/health
```

**Ответ:**
```json
{
  "status": "ok"
}
```


