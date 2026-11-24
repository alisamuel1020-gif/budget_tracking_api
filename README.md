# Budget Tracker API

A JSON API for recording personal income and expenses. It stores each transaction with an amount, type, category, and date, and it reports totals and spending by category.

Built with Go's standard library only, with no third-party dependencies.

## Features

- Create, list, view, and delete transactions
- Money stored as whole cents to avoid floating-point rounding errors
- Input validation with clear error messages
- Summary of total income, total expenses, balance, and expenses by category
- Thread-safe in-memory storage
- Tests for the HTTP handlers, validation, and summary logic

## Project structure

```
.
├── main.go       Starts the HTTP server
├── server.go     Routes and request handlers
├── model.go      Transaction type and validation rules
├── store.go      In-memory storage and summary calculation
├── api_test.go   Tests
├── go.mod
└── README.md
```

## Requirements

- [Go](https://go.dev/dl/) 1.22 or newer

## Run it

From the project folder:

```
go run .
```

The server listens on port 8080. To use another port, set `PORT`:

```
PORT=3000 go run .
```

Data is kept in memory, so it is cleared each time the server stops.

## Test it

```
go test ./...
```

Add `-v` to see each test.

## API

All request and response bodies are JSON.

| Method   | Path                     | Description                     | Success |
| -------- | ------------------------ | ------------------------------- | ------- |
| `GET`    | `/api/health`            | Status check                    | 200     |
| `POST`   | `/api/transactions`      | Create a transaction            | 201     |
| `GET`    | `/api/transactions`      | List transactions, newest first | 200     |
| `GET`    | `/api/transactions/{id}` | Get one transaction             | 200     |
| `DELETE` | `/api/transactions/{id}` | Delete a transaction            | 204     |
| `GET`    | `/api/summary`           | Totals and expenses by category | 200     |

### Transaction fields

| Field         | Type   | Rules                                      |
| ------------- | ------ | ------------------------------------------ |
| `description` | string | Required, up to 200 characters             |
| `amountCents` | int    | Required, greater than 0 (`4550` = $45.50) |
| `type`        | string | `"income"` or `"expense"`                  |
| `category`    | string | Required                                   |
| `date`        | string | `YYYY-MM-DD`                               |

### Errors

| Status | Meaning                                      |
| ------ | -------------------------------------------- |
| 400    | Malformed JSON, unknown field, or invalid ID |
| 404    | Transaction not found                        |
| 422    | A field failed validation                    |

Error responses look like `{"error": "amountCents must be greater than 0"}`.

## Examples

Add income:

```
curl -X POST localhost:8080/api/transactions \
  -d '{"description":"Paycheck","amountCents":300000,"type":"income","category":"Salary","date":"2026-11-01"}'
```

Add an expense:

```
curl -X POST localhost:8080/api/transactions \
  -d '{"description":"Groceries","amountCents":4550,"type":"expense","category":"Food","date":"2026-11-03"}'
```

Get the summary:

```
curl localhost:8080/api/summary
```

Example response:

```json
{
  "incomeCents": 300000,
  "expenseCents": 4550,
  "balanceCents": 295450,
  "expenseByCategory": {"Food": 4550}
}
```

## Design notes

- **Cents, not dollars.** Floating-point numbers can't represent values like 0.10 exactly, so money is stored as integers.
- **Mutex-protected store.** HTTP handlers run concurrently, so the in-memory map is locked during reads and writes.
- **Strict JSON decoding.** Unknown fields are rejected, which catches typos in client requests.

## Planned improvements

- Persist transactions in SQLite
- Update transactions (`PUT`)
- Filter by date range or category
- User accounts and authentication
