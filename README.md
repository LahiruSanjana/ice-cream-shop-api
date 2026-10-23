# Ice Cream Shop API

A REST API for managing an ice cream shop's users, customers, categories, items,
and orders. The application is built with Go, Fiber, GORM, and SQLite.

## Features

- User registration and login with JWT authentication
- Role-based access control for owner, manager, and cashier roles
- Customer management
- Category and inventory item management
- Order creation with stock validation and transaction rollback
- Order filtering by date and customer
- Paginated order listing
- Automatic database migrations and development seed data
- Unit and handler tests, including validation and failure cases

## Technology stack

- **Language:** Go 1.26.4
- **Web framework:** Fiber v2
- **ORM:** GORM
- **Database:** SQLite
- **Authentication:** JWT stored in an HTTP-only cookie
- **Password hashing:** bcrypt

## Project structure

```text
.
├── cmd/server/main.go       # Application entry point
├── src/assets/              # ER Diagram
├── src/database/            # Database connection, migrations, and seed data
├── src/handlers/            # HTTP handlers and tests
├── src/middleware/          # Authentication and role middleware
├── src/models/              # GORM entities
└── src/routes/              # API route registration
```

## Requirements

- Go 1.26.4 or a compatible newer Go version
- Git

No separate database server is required because the application uses SQLite.

## Configuration

The application loads variables from a `.env` file in the project root. Create
one with the following values:

```env
JWT_SECRET=replace-with-a-long-random-secret
DB_NAME=icecream.db
```

`JWT_SECRET` is required for protected routes and must be kept private. Do not
commit real secrets to source control.

The server currently listens on `http://localhost:8000`. Although an existing
`.env` may contain `PORT`, the current server entry point uses port `8000`
directly.

## Run locally

From the repository root:

```bash
go mod download
go run ./cmd/server
```

When the server starts, it:

1. Loads the environment variables.
2. Opens the SQLite database configured by `DB_NAME`.
3. Runs GORM auto-migrations.
4. Seeds the initial development users if the users table is empty.
5. Starts the API on port `8000`.

Health check:

```bash
curl http://localhost:8000/
```

Expected response:

```text
App is running healthy.
```

## Seed users for local development

The first application start creates these users when no users exist:

| Role | Email | Password |
|---|---|---|
| Owner | `owner@icecream.com` | `owner123` |
| Manager | `manager@icecream.com` | `manager123` |
| Cashier | `cashier@icecream.com` | `cashier123` |

These credentials are for local development only. Change or remove them before
deploying the application to a shared or production environment.

## Authentication

Register or log in through the auth endpoints. A successful login sets a `jwt`
cookie. Protected endpoints read the token from that cookie, so clients must
preserve and send cookies with subsequent requests.

Example login:

```bash
curl -i -c cookies.txt \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"owner@icecream.com\",\"password\":\"owner123\"}" \
  http://localhost:8000/api/auth/login
```

Example authenticated request:

```bash
curl -b cookies.txt http://localhost:8000/api/customers
```

## API endpoints

Base URL: `http://localhost:8000/api`

### Authentication

| Method | Endpoint | Authentication | Description |
|---|---|---|---|
| `POST` | `/auth/register` | No | Register a user |
| `POST` | `/auth/login` | No | Log in and receive a JWT cookie |
| `POST` | `/auth/logout` | No | Clear the JWT cookie |

### Customers

All customer endpoints require authentication.

| Method | Endpoint | Permission | Description |
|---|---|---|---|
| `GET` | `/customers` | Any authenticated user | List customers |
| `GET` | `/customers/:id` | Any authenticated user | Get one customer |
| `POST` | `/customers` | Any authenticated user | Create a customer |
| `PUT` | `/customers/:id` | Any authenticated user | Update a customer |
| `DELETE` | `/customers/:id` | Owner | Delete a customer |

### Categories

All category endpoints require authentication.

| Method | Endpoint | Permission | Description |
|---|---|---|---|
| `GET` | `/categories` | Any authenticated user | List categories |
| `GET` | `/categories/:id` | Any authenticated user | Get one category |
| `POST` | `/categories` | Owner, Manager | Create a category |
| `PUT` | `/categories/:id` | Owner, Manager | Update a category |
| `DELETE` | `/categories/:id` | Owner, Manager | Delete a category |

### Items

All item endpoints require authentication.

| Method | Endpoint | Permission | Description |
|---|---|---|---|
| `GET` | `/items` | Any authenticated user | List items |
| `GET` | `/items/:id` | Any authenticated user | Get one item |
| `POST` | `/items` | Owner, Manager | Create an item |
| `PUT` | `/items/:id` | Any authenticated user | Update an item |
| `DELETE` | `/items/:id` | Owner | Delete an item |

### Orders

All order endpoints require authentication.

| Method | Endpoint | Permission | Description |
|---|---|---|---|
| `POST` | `/orders` | Any authenticated user | Create an order |
| `GET` | `/orders` | Any authenticated user | Get paginated orders |
| `GET` | `/orders/latest` | Any authenticated user | Get orders from the last 24 hours |
| `GET` | `/orders/date?date=YYYY-MM-DD` | Any authenticated user | Get orders for a date |
| `GET` | `/orders/customer/:customer_id` | Any authenticated user | Get orders for a customer |

Pagination parameters:

```text
GET /api/orders?page=1&limit=10
```

`page` and `limit` must be positive integers. The maximum limit is `100`.

## Request examples

### Create a customer

```bash
curl -b cookies.txt \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Jane Doe\",\"email\":\"jane@example.com\",\"phone\":\"0771234567\",\"address\":\"Main Street\"}" \
  http://localhost:8000/api/customers
```

### Create a category

```bash
curl -b cookies.txt \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Ice Cream\"}" \
  http://localhost:8000/api/categories
```

### Create an item

```bash
curl -b cookies.txt \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"Chocolate Ice Cream\",\"description\":\"Rich chocolate flavor\",\"price\":5.50,\"quantity\":10,\"category_id\":1}" \
  http://localhost:8000/api/items
```

### Create an order

```bash
curl -b cookies.txt \
  -H "Content-Type: application/json" \
  -d "{\"customer_id\":1,\"items\":[{\"item_id\":1,\"quantity\":2}]}" \
  http://localhost:8000/api/orders
```

An order is rejected when an item does not exist, the customer does not exist,
the quantity is not positive, or there is insufficient stock. Failed order
transactions do not reduce inventory or create a partial order.

## Testing

Run all tests:

```bash
go test ./...
```

Run the handler tests only:

```bash
go test ./src/handlers
```

Run tests with the race detector:

```bash
go test -race ./...
```

The test suite covers successful requests, invalid request bodies, missing
required fields, not-found cases, authorization failures, invalid parameters,
and order stock rollback behavior.

## HTTP status codes

| Status | Meaning |
|---:|---|
| `200 OK` | Request completed successfully |
| `201 Created` | Resource created successfully |
| `400 Bad Request` | Invalid JSON, missing data, or invalid parameters |
| `401 Unauthorized` | Authentication is missing or invalid |
| `403 Forbidden` | Authenticated user lacks the required role |
| `404 Not Found` | Requested resource does not exist |
| `409 Conflict` | Order cannot be completed because stock is insufficient |
| `500 Internal Server Error` | Unexpected server or database failure |

## Notes for production

- Use a strong, unique `JWT_SECRET`.
- Do not use the seeded development passwords in production.
- Store `.env` outside version control.
- Run the application behind HTTPS so the secure JWT cookie can be transmitted.
- Replace development seed data with a controlled migration or provisioning
  process.
