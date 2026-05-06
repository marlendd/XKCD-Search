# XKCD Search Service

Микросервисная система для поиска по комиксам XKCD с управлением нагрузкой, авторизацией, мониторингом и web-интерфейсом.

## Архитектура

Система состоит из пяти Go-сервисов, взаимодействующих через gRPC и HTTP:

```
client → frontend (BFF) → api (REST) → search (gRPC)
                                     → update (gRPC)
                                     → words  (gRPC)
                                            ↓
                                         postgres
                                            ↑
                                      nats (pub/sub)
```

- `frontend` — BFF, web-интерфейс для браузера (html/template, cookie-авторизация)
- `api` — REST gateway, точка входа для клиентов
- `search` — поиск комиксов по базе данных, перестройка индекса по событию NATS
- `update` — загрузка и обновление комиксов с xkcd.com, публикует событие в NATS
- `words` — нормализация слов (Snowball-стемминг)

## Middleware

### Concurrency limiter (`GET /api/search`)
Ограничивает количество одновременных запросов. При превышении лимита возвращает `503 Service Unavailable`.
Настраивается через `SEARCH_CONCURRENCY`.

### Rate limiter (`GET /api/isearch`)
Ограничивает скорость запросов (RPS) без отклонения — задерживает соединения.
Реализован как leaky bucket. Настраивается через `SEARCH_RATE`.

### Auth middleware
Проверяет JWT токен из заголовка `Authorization: Token <token>`.
Защищает `POST /api/db/update` и `DELETE /api/db`.

### Metrics middleware
Перехватывает все запросы, измеряет время ответа и HTTP статус.
Экспортирует histogram `http_request_duration_seconds` с метками `status` и `url`.

## API

| Метод | Endpoint | Описание | Авторизация |
|-------|----------|----------|-------------|
| POST | `/api/login` | Получить JWT токен | — |
| GET | `/api/ping` | Проверка доступности сервисов | — |
| GET | `/api/search` | Поиск комиксов (с concurrency limit) | — |
| GET | `/api/isearch` | Индексный поиск комиксов (с rate limit) | — |
| POST | `/api/db/update` | Обновить базу комиксов (async) | JWT |
| DELETE | `/api/db` | Удалить базу комиксов | JWT |
| GET | `/api/db/stats` | Статистика базы | — |
| GET | `/api/db/status` | Статус обновления | — |
| GET | `/metrics` | Метрики Prometheus/VictoriaMetrics | — |

## Web-интерфейс

| URL | Описание |
|-----|----------|
| `http://localhost:28084/` | Поиск комиксов |
| `http://localhost:28084/login` | Вход в систему |
| `http://localhost:28084/admin` | Панель администратора |

## Авторизация

Через web-интерфейс: `http://localhost:28084/login`. Имя и пароль задаются в `.env`.

Через API напрямую:
```bash
TOKEN=$(curl -s -X POST \
  -d "{\"name\": \"$ADMIN_USER\", \"password\": \"$ADMIN_PASSWORD\"}" \
  localhost:28080/api/login)

curl -X POST -H "Authorization: Token $TOKEN" localhost:28080/api/db/update
```

Токен выдаётся на `TOKEN_TTL` (по умолчанию 2 минуты) и подписывается HS256 с ключом `JWT_SECRET`.

## Запуск

```bash
cp .env.example .env
# Замените демонстрационные значения в .env перед запуском.
make test container_runtime=/usr/local/bin/docker
```

Или пошагово:

```bash
docker compose up --build -d
sleep 10
docker run --rm --network=host --env-file .env tests:latest
```

## Конфигурация (переменные среды)

| Переменная | Описание | По умолчанию |
|------------|----------|--------------|
| `API_ADDRESS` | Адрес API сервера | `:8080` |
| `ADMIN_USER` | Имя администратора | — |
| `ADMIN_PASSWORD` | Пароль администратора | — |
| `JWT_SECRET` | Ключ подписи JWT | — |
| `TOKEN_TTL` | Время жизни JWT токена | `2m` |
| `SEARCH_CONCURRENCY` | Лимит одновременных запросов к /api/search | `10` |
| `SEARCH_RATE` | Лимит RPS для /api/isearch | `100` |
| `WORDS_ADDRESS` | Адрес words сервиса | — |
| `UPDATE_ADDRESS` | Адрес update сервиса | — |
| `SEARCH_ADDRESS` | Адрес search сервиса | — |
| `FRONTEND_ADDRESS` | Адрес web-интерфейса | `:8080` |

## Мониторинг

- VictoriaMetrics: http://localhost:8428
- Grafana: http://localhost:3000 (при первом входе стандартно `admin / admin`)
- pgAdmin: http://localhost:18888 (данные входа задаются в `.env`)

Файл дашборда находится в `metrics/dashboard.json`. После запуска добавьте VictoriaMetrics
(`http://victoriametrics:8428`) как Prometheus datasource в Grafana и импортируйте дашборд вручную.
Он показывает RPS по URL, статусы ответов и latency в ms.

## Локальные проверки

В репозитории три отдельных Go-модуля. Проверить их можно командами:

```bash
(cd search-services && go test ./... && go vet ./...)
(cd examples && go test ./... && go vet ./...)
(cd tests && go test -run '^$' ./... && go vet ./...)
```

Последняя команда только компилирует интеграционные тесты. Для полного прогона используйте `make test`.

## Стек

- Go 1.25+
- gRPC / protobuf
- PostgreSQL 18
- NATS (pub/sub)
- JWT (HS256) — `golang-jwt/jwt/v5`
- html/template + cookie-авторизация
- VictoriaMetrics + Grafana
- Docker Compose
