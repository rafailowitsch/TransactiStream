# Запуск и использование приложения

Для запуска приложения необходимо использовать Docker Compose. Выполните следующую команду:

```sh
docker-compose up --build -d
```

В случае, если один из контейнеров не запуститься, выполните следующую команду:

```sh
docker-compose up -d
```

## URL приложения

### POST: /transaction

Добавляет новую транзакцию.

**Пример запроса:**

```sh
curl -X POST 0.0.0.0:8009/transaction -H "Content-Type: application/json" -d '{"user_id": "user123", "amount": 100.5, "currency": "USD"}'
```

### GET: /transactions

Получает список всех транзакций.

**Пример запроса:**

```sh
curl 0.0.0.0:8009/transactions
```

**Пример ответа:**

```json
{
  "id": "5b51fb04-c74d-48ed-bb3e-16b906f2a285",
  "user_id": "123",
  "amount": 99,
  "currency": "usdt",
  "done": true,
  "timestamp": "2024-07-31T20:04:33.828556Z"
}
```

### GET: /statistics

Получает статистику по транзакциям.

**Пример запроса:**

```sh
curl 0.0.0.0:8009/statistics
```

**Пример ответа:**

```json
{
  "total_transactions": 9,
  "failed_transactions": 2,
  "total_users": 1,
  "average_processing_time": 25.124,
  "currencies": ["usdt"]
}
```