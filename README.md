# Avito тестовое задание

Этот проект — бэкенд-сервис для работы с пользователями, балансами, операциями и покупками мерча. В нём реализована базовая логика с использованием транзакций, in-memory кэша и HTTP API. Для удобства проверки сервис полностью упакован в Docker, и база данных (PostgreSQL) также поднимается в контейнере с помощью docker-compose.

## Структура проекта

- **cmd/app/main.go** – Точка входа приложения.
- **internal/model/** – Модели данных (User, Merch, Purchase, Operation, UserInfo).
- **internal/repository/** – Логика работы с базой (PostgreSQL) и транзакционные операции.
- **internal/cache/** – In-memory кэш с воркером для периодического обновления.
- **internal/handler/** – HTTP-эндпоинты для API (создание пользователя, перевод, покупка и получение данных).
- **migrations/** – SQL-скрипты для создания схемы (например, `01_schema.sql`) и заполнения тестовыми данными (seed).
- **Dockerfile** – Сборка Go-приложения в контейнер.
- **docker-compose.yml** – Поднятие контейнеров с приложением и PostgreSQL.
- **README.md** – Эта инструкция.

## Требования

- Docker Desktop (или Docker Engine) установлен.
- Docker Compose версии 2 или выше.

## Быстрый старт

1. **Клонируйте репозиторий и перейдите в его директорию:**

   ```bash
   git clone <URL_репозитория>
   cd avito
2.  **Запустите контейнеры с помощью Docker Compose**
    ```bash
    docker compose up --build
    
3. **Проверьте, что контейнеры запущены**
    ```bash
    docker compose ps
Вы должны увидеть два контейнера: например, avito_db и avito_app.

4.  **Тестирование**

    4.1 Создание пользователя (POST /users): 
    $headers = @{ "Content-Type" = "application/json" }
    $body = '{"name":"Vasya","password":"secret","balance":1000}'
    Invoke-WebRequest -Uri "http://localhost:8080/users" -Method POST -Headers $headers -Body $body
    
    4.2 Перевод средств между пользователями (POST /transfer)
    $headers = @{ "Content-Type" = "application/json" }
    $body = '{"from_user":1,"to_user":2,"amount":300}'
    Invoke-WebRequest -Uri "http://localhost:8080/transfer" -Method POST -Headers $headers -Body $body

    4.3 Покупка мерча (POST /purchase)
    $headers = @{ "Content-Type" = "application/json" }
    $body = '{"user_id":1,"merch_id":1,"amount":2}'
    Invoke-WebRequest -Uri "http://localhost:8080/purchase" -Method POST -Headers $headers -Body $body

    4.4 Получение данных пользователя (GET /get_user)
    Invoke-WebRequest -Uri "http://localhost:8080/get_user?id=1" -Method GET

    проверка например у пользователя 1: "http://localhost:8080/get_user?id=1"

5. **Остановка контейнеров**
    ```bash
   docker compose down -v