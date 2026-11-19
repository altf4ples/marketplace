# Marketplace — учебный backend на Go

Модульный монолит маркетплейса на Go, PostgreSQL и Redis. Проект демонстрирует проектирование доменных модулей, транзакции, конкурентный доступ, Redis Lua, state machine, HTTP API и контейнеризацию.

> **Статус:** учебный релиз `v1.0.0`. Это не production-магазин: нет аутентификации/авторизации, платежей, складских остатков, доставки и реальной логистики. `user_id` передаётся клиентом и не является механизмом безопасности.

## Возможности

- Пользователи: создание пользователя.
- Каталог: создание магазина, создание товара, просмотр товаров магазина.
- Корзина: Redis Hash `cart:{user_id}`, количество товара, TTL 7 дней, атомарное добавление и получение.
- Заказы: checkout, снимки цены/названий, история состояний, просмотр списка и деталей.
- State machine: `created → assembling → shipped → delivered`; отмена разрешена из `created` и `assembling`.
- Уведомления: сохраняются в PostgreSQL в той же транзакции, что и событие заказа.
- Инфраструктура: JSON-логи, `request_id`, panic recovery, health endpoint, graceful shutdown.

## Стек

- Go 1.27.3
- PostgreSQL 16, pgx/v5
- Redis 7, go-redis/v9
- `shopspring/decimal`
- `net/http` (method-aware ServeMux)
- Docker Compose
- golang-migrate

## Быстрый запуск

Требуется Docker Engine с Compose plugin и Git.

```bash
git clone https://github.com/<your-username>/marketplace.git
cd marketplace
cp .env.example .env
```

Измени пароли в `.env` и синхронно обнови пароль в `DATABASE_URL`, затем:

```bash
docker compose up --build
```

API будет доступно на `http://localhost:8080`. Проверка готовности:

```bash
curl -i http://localhost:8080/healthz
```

Остановка без удаления данных:

```bash
docker compose down
```

Удаление PostgreSQL и Redis volumes (необратимо):

```bash
docker compose down -v
```

## Конфигурация

Все настройки читаются из переменных окружения:

| Переменная | Назначение |
|---|---|
| `HTTP_ADDR` | HTTP bind address, по умолчанию `:8080` |
| `DATABASE_URL` | PostgreSQL DSN |
| `REDIS_ADDR` | Redis address; default `localhost:6379`, in Compose `redis:6379` |
| `REDIS_PASSWORD` | Redis password |
| `LOG_LEVEL` | `debug`, `info`, `warn`, `error` |

Для Compose в `DATABASE_URL` используется hostname `postgres`, для локального запуска приложения вне Compose — `localhost`. При локальном запуске вне Compose также установи `REDIS_ADDR=localhost:6379`.

## API

Все тела запросов и ответы — JSON. Ошибки возвращаются в едином формате:

```json
{"error":{"code":"validation_error","message":"..."}}
```

| Метод | Endpoint | Назначение |
|---|---|---|
| `POST` | `/users` | Создать пользователя |
| `POST` | `/store` | Создать магазин |
| `POST` | `/stores/{storeID}/products` | Создать товар |
| `GET` | `/stores/{storeID}/products` | Список товаров магазина |
| `GET` | `/cart?user_id={uuid}` | Получить корзину |
| `POST` | `/cart/items` | Добавить количество товара |
| `DELETE` | `/cart/items/{productID}?user_id={uuid}` | Удалить позицию |
| `DELETE` | `/cart?user_id={uuid}` | Очистить корзину |
| `POST` | `/orders` | Оформить заказ |
| `GET` | `/orders/{orderID}?user_id={uuid}` | Детали заказа |
| `GET` | `/users/{userID}/orders` | Заказы пользователя |
| `POST` | `/orders/{orderID}/events` | Применить событие state machine |
| `GET` | `/users/{userID}/notifications` | Уведомления пользователя |
| `GET` | `/healthz` | Проверка PostgreSQL и Redis |

### Полный сценарий через curl

1. Создать пользователя:

```bash
curl -s http://localhost:8080/users -H 'Content-Type: application/json' \
  -d '{"name":"Alex","phone_number":"+79990000000"}'
```

Сохрани `id` из ответа как `USER_ID`.

2. Создать магазин:

```bash
curl -s http://localhost:8080/store -H 'Content-Type: application/json' \
  -d '{"name":"Demo Shop","description":"Учебный магазин"}'
```

Сохрани `id` как `STORE_ID`.

3. Создать товар:

```bash
curl -s http://localhost:8080/stores/$STORE_ID/products \
  -H 'Content-Type: application/json' \
  -d '{"name":"Keyboard","description":"Mechanical keyboard","price":"4990.00"}'
```

Сохрани `id` как `PRODUCT_ID`.

4. Добавить товар в корзину:

```bash
curl -s http://localhost:8080/cart/items -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"$USER_ID\",\"product_id\":\"$PRODUCT_ID\",\"quantity\":2}"
```

5. Оформить заказ:

```bash
curl -s http://localhost:8080/orders -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"$USER_ID\"}"
```

Сохрани `id` заказа как `ORDER_ID`.

6. Перевести заказ по жизненному циклу:

```bash
curl -s http://localhost:8080/orders/$ORDER_ID/events \
  -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"$USER_ID\",\"event\":\"assemble\"}"

curl -s http://localhost:8080/orders/$ORDER_ID/events \
  -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"$USER_ID\",\"event\":\"ship\"}"

curl -s http://localhost:8080/orders/$ORDER_ID/events \
  -H 'Content-Type: application/json' \
  -d "{\"user_id\":\"$USER_ID\",\"event\":\"deliver\"}"
```

7. Получить уведомления:

```bash
curl -s http://localhost:8080/users/$USER_ID/notifications
```

## Архитектура

```text
cmd/api
  └── composition root: создаёт зависимости и запускает HTTP server

internal/
  ├── users/         пользователь и проверка существования
  ├── catalog/       магазины и товары
  ├── cart/           Redis Hash + Lua scripts
  ├── order/          checkout + state machine + PostgreSQL transactions
  ├── notification/  чтение уведомлений
  └── platform/      config, clients, middleware, HTTP helpers

migrations/           SQL schema migrations
docs/                 OpenAPI и технические заметки
```

Модули зависят от узких интерфейсов. Реализации PostgreSQL/Redis находятся в соответствующих repository, а `cmd/api` связывает их вместе.

## Важные технические решения

### Корзина

Корзина хранится как Redis Hash: поле — `product_id`, значение — quantity. Lua-скрипт атомарно увеличивает количество и обновляет TTL. Второй Lua-скрипт выполняет `HGETALL + DEL` как одну Redis-операцию.

Checkout восстанавливает извлечённую корзину при ошибке валидации или SQL. Восстановление использует `HINCRBY`, поэтому не затирает товары, добавленные параллельно. Это компенсирующая операция, а не распределённая транзакция: аварийное завершение процесса между Redis Take и SQL commit остаётся ограничением этой учебной версии.

### Заказ и цена

Checkout проверяет пользователя, извлекает корзину, загружает актуальные товары и считает сумму через `decimal`. В одной PostgreSQL-транзакции создаются `orders`, `order_items`, начальная запись истории и уведомление. В `order_items` сохраняются цена, название товара и название магазина на момент покупки.

### State machine

Допустимые переходы:

```text
created ──assemble──> assembling ──ship──> shipped ──deliver──> delivered
   │                       │
   └──────── cancel ───────┴──────────────> cancelled
```

Переход блокирует строку заказа через `SELECT ... FOR UPDATE`. Обновление статуса, запись истории и уведомления коммитятся вместе.

### Уведомления

Отдельный брокер не используется. Уведомления — записи PostgreSQL. Order repository вызывает transaction-scoped `NotificationWriter`, реализованный notification repository, поэтому запись уведомления фиксируется в той же транзакции, что и событие заказа. Это надёжно для хранения уведомлений в БД, но не реализует отправку push/email.

## Разработка и проверки

Локально нужны Go 1.27.3 и Docker Compose.

```bash
make fmt
make test
make test-race
make vet
make build
make check
``` 

После `docker compose up --build` можно запустить сквозную проверку:

```bash
make smoke
```

Проверка интеграционного окружения:

```bash
docker compose config
docker compose up --build
```

Миграции запускаются сервисом `migrate` при старте Compose. SQL-файлы имеют парные `.up.sql` и `.down.sql`.

## Ограничения

- Нет аутентификации и авторизации: API принимает `user_id` от клиента.
- Нет управления остатками и резервирования stock.
- Нет платежей, доставки и реальной логистики.
- Нет внешней отправки уведомлений.
- Redis и PostgreSQL не имеют общей транзакции; checkout компенсирует известные ошибки, но не все сценарии внезапного падения процесса.
- Не предназначен для публичного production-развёртывания без security hardening.

## License

MIT — см. [LICENSE](LICENSE).

## Tests and quality checks

Run the local checks before opening a pull request:

```bash
make check
```

The CI pipeline runs:

- `go test -race -cover ./...` — unit tests and race detection;
- `go vet ./...` — Go static analysis;
- `gofmt` verification — fails if Go sources are not formatted;
- `CGO_ENABLED=0 go build ./...` — compiles every package;
- `docker compose config --quiet` — validates the Compose configuration.

Business-logic tests cover user/store/product validation, order checkout behavior,
price snapshots, cart restoration after checkout failures, and allowed/forbidden
state-machine transitions. HTTP tests use `net/http/httptest` and verify successful
user creation and malformed JSON handling.

PostgreSQL and Redis repository integration tests are kept behind the `integration`
build tag. They run in GitHub Actions against PostgreSQL 16 and Redis 7 services:

```bash
go test -tags=integration -race -v ./internal/integration/...
```

To run them locally, set `TEST_DATABASE_URL` and `TEST_REDIS_ADDR` to reachable
PostgreSQL and Redis instances. The integration tests create the minimal `store.users`
table if it does not exist and clean up their test user/cart data.
