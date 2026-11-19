# Технический аудит — Marketplace v1.0.0

## Что реализовано

- `users`: создание пользователя, проверка существования, обработка unique violation.
- `catalog`: создание магазина и товара, получение списка товаров магазина, чтение товара вместе с названием магазина для checkout.
- `cart`: Redis Hash `cart:{user_id}`, количество, TTL 7 дней, Lua для атомарного increment и Take, компенсирующий Restore.
- `order`: checkout, ценовые и текстовые snapshots, список/детали заказов, state machine, история состояний.
- `notification`: чтение уведомлений пользователя и transaction-scoped writer; order repository передаёт ему текущий `pgx.Tx`, чтобы запись уведомления была атомарна с заказом/переходом.
- `platform`: конфигурация, PostgreSQL/Redis clients, JSON helpers, request ID logging, panic recovery, health check, HTTP timeouts.
- Docker Compose, миграции, OpenAPI, unit tests, end-to-end smoke script, CI и GitHub Release workflow.

## Checkout

1. Проверяется существование пользователя.
2. Redis Lua атомарно читает и удаляет корзину.
3. Сервис валидирует quantity и загружает текущие данные товаров.
4. Сумма считается через `shopspring/decimal`.
5. PostgreSQL transaction создаёт заказ, позиции со snapshot полями, начальную запись истории и уведомление.
6. При известной ошибке после Take сервис пытается вернуть позиции через Lua/HINCRBY с сохранением параллельных добавлений.

### Ограничение распределённой согласованности

Redis и PostgreSQL не участвуют в общей транзакции. Компенсация покрывает ошибки, которые процесс успел обработать, но не падение процесса между Redis Take и PostgreSQL commit. Также нет idempotency key для checkout. Это явно задокументированное ограничение учебного v1.0.0; для production понадобятся lease/claim workflow, идемпотентность и механизм reconciliation.

## State machine

- `created -> assembling` по `assemble`
- `created -> cancelled` по `cancel`
- `assembling -> shipped` по `ship`
- `assembling -> cancelled` по `cancel`
- `shipped -> delivered` по `deliver`

Переход выполняется под `SELECT ... FOR UPDATE`. Обновление заказа, запись истории и запись уведомления выполняются одной PostgreSQL-транзакцией. Поэтому параллельные конфликтующие события сериализуются на строке заказа.

## Осознанные ограничения

- Нет authentication/authorization. `user_id` поступает от клиента и не является доказательством личности.
- Нет stock/inventory, резервирования товара и проверки доступного количества.
- Нет оплаты, курьеров, доставки и внешней логистики.
- Уведомления хранятся в БД, но не отправляются по email/push.
- Нет checkout idempotency key.
- Нет управления магазином/товаром (update/delete), пагинации и поиска.
- Redis cart и PostgreSQL orders не имеют общей транзакции.
- Физическое удаление пользователей не реализовано; FK заказов сохраняет целостность.
- Интеграционная проверка предоставлена как `make smoke`; CI запускает unit tests, race detector, vet, build и Compose config validation.

## Перед публичным production-развёртыванием

1. Добавить аутентификацию и проверку владельца ресурсов.
2. Добавить idempotency key и durable cart claim/recovery.
3. Добавить inventory reservation и правила изменения цены.
4. Добавить rate limiting, CORS policy, security headers и аудит действий.
5. Добавить integration tests в CI с реальными PostgreSQL и Redis.
6. Добавить метрики, трассировку, секреты через secret manager и backup/restore.
7. Добавить нагрузочные тесты и политику retention для уведомлений/истории.

## Приёмка релиза

- [x] Все бизнес-модули присутствуют.
- [x] SQL migrations up/down присутствуют.
- [x] Checkout и state transitions транзакционные в PostgreSQL.
- [x] Redis cart uses Lua for atomic operations.
- [x] Unit tests для state machine, checkout и catalog validation.
- [x] End-to-end smoke script.
- [x] OpenAPI.
- [x] Docker Compose.
- [x] GitHub Actions CI and release workflows.
- [ ] `go test -race ./...` на Go 1.27.3.
- [ ] `docker compose up --build` и `make smoke` на Docker host.
- [ ] Создать GitHub repository, push `main`, push tag `v1.0.0`.
