# Wallet Service

REST-сервис для управления балансом кошельков. Стек: Go, chi, PostgreSQL 18, Docker Compose.

## Запуск

```bash
git clone <repo-url>
cd <repo>
docker compose up --build
```

Сервис доступен на `http://localhost:8080`.
Миграции из `migrations/` применяются автоматически при первой инициализации БД.
Чтобы пересоздать БД с нуля: `docker compose down -v`.

## Конфигурация

Переменные окружения читаются из `config.env` (в compose передаются через `env_file`).

| Переменная | Описание | Пример |
|---|---|---|
| `HTTP_ADDRESS` | Адрес HTTP-сервера | `:8080` |
| `HTTP_READ_TIMEOUT` / `HTTP_WRITE_TIMEOUT` / `HTTP_IDLE_TIMEOUT` | Таймауты сервера | `5s` / `10s` / `60s` |
| `HTTP_SHUTDOWN_TIMEOUT` | Таймаут graceful shutdown | `10s` |
| `POSTGRES_DSN` | Строка подключения | `postgres://app:pass@postgres:5432/wallets?sslmode=disable` |
| `POSTGRES_MAX_CONNS` / `POSTGRES_MIN_CONNS` | Размер пула соединений | `30` / `5` |
| `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | Параметры контейнера PostgreSQL | `app`, `pass`, `wallets` |

Внутри сети Docker приложение подключается к `postgres:5432`. Порт `5433` нужен только для доступа с хоста.

## API

Суммы передаются целым числом в минимальных единицах валюты: `10000` = 100 RUB.

### Создать кошелёк
`POST /api/v1/wallet`
```json
{
  "currency": "RUB"
}
```
Ответ `201`:
```json
{
  "walletId": "01a14fe6-1eb6-7a0a-acd9-03bb84c68b30",
  "balance": 0,
  "currency": "RUB"
}
```
### Получить баланс

`GET /api/v1/wallets/{walletId}`

Ответ `200`: 
```json
{
  "walletId": "01a14fe6-1eb6-7a0a-acd9-03bb84c68b30",
  "balance": 0,
  "currency": "RUB"
}
```
### Коды ошибок

| Код | Причина |
|---|---|
| 400 | Невалидный JSON, UUID, тип операции или сумма |
| 404 | Кошелёк не найден |
| 409 | Недостаточно средств при `WITHDRAW` |
| 500 | Внутренняя ошибка |

### Изменить баланс

`POST /api/v1/wallets`

```json
{
  "walletId": "6f1c2c0e-8f6b-4c53-9a58-1a2b3c4d5e6f",
  "operationType": "DEPOSIT",
  "amount": 1000
}
```
`operationType`: `DEPOSIT` или `WITHDRAW`. `amount`: целое число больше 0 (в минимальных единицах).

Ответ `200`:

```json
{
  "operationId": "01a14fe6-68be-7cf7-863d-94cc543aae77",
  "walletId": "01a14fe6-1eb6-7a0a-acd9-03bb84c68b30",
  "status": "APPLIED"
}
```
## Идемпотентность

- Каждая операция регистрируется в таблице `operations` с уникальным `idempotency_key`.
- Повтор запроса с тем же ключом и теми же параметрами не меняет баланс и возвращает сохранённый результат.
- Баланс и запись об операции меняются в одной транзакции. Гонка двух запросов с одним ключом разрешается уникальным индексом.
