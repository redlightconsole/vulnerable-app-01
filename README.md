# vulnerable-app-01

A deliberately vulnerable Go web application built for security research and education. It demonstrates a classic **SQL injection** vulnerability in a simple login form backed by SQLite.

> ⚠️ **Warning:** This application contains intentional security vulnerabilities. Run it only in isolated, local environments. Never expose it to the public internet.

---

## Features

- Landing page (`/`) served via Go HTML templates
- Login page (`/login`) with a form that is vulnerable to SQL injection
- SQLite database with a `users` table seeded on first run
- Dockerfile for easy containerised execution

## Prerequisites

**Native Go**
- [Go 1.25+](https://go.dev/dl/)

**Docker**
- [Docker](https://docs.docker.com/get-docker/)

## Running the App

### With Go

```bash
git clone https://github.com/redlightconsole/vulnerable-app-01.git
cd vulnerable-app-01
go run .
```

### With Docker

```bash
docker build -t vulnapp .
docker run -p 8080:8080 vulnapp
```

The app listens on **http://localhost:8080**.

## Project Structure

```
.
├── main.go              # HTTP server, route handlers, DB initialisation
├── templates/
│   ├── index.html       # Landing page
│   └── login.html       # Login form
├── Dockerfile           # Two-stage Docker build
├── go.mod
└── go.sum
```

The SQLite database (`users.db`) is created automatically on first run in the working directory.

## Seeded Users

| Username | Password      | Email                |
|----------|---------------|----------------------|
| admin    | S3cr3tP@ss!   | admin@example.com    |
| alice    | alice1234     | alice@example.com    |
| bob      | b0bpass       | bob@example.com      |

## SQL Injection Vulnerability

The login handler in `main.go` builds its SQL query by concatenating raw user input directly into the query string without parameterisation:

```go
query := fmt.Sprintf(
    "SELECT id, username, email FROM users WHERE username = '%s' AND password = '%s'",
    username, password,
)
```

### Authentication bypass

Enter the following in the **Username** field (leave the password blank or anything):

```
' OR '1'='1
```

This transforms the query into:

```sql
SELECT id, username, email FROM users WHERE username = '' OR '1'='1' AND password = ''
```

`'1'='1'` is always true, so the query returns the first row in the table and grants access.

### Union-based data extraction

Retrieve all usernames and passwords in a single request:

```
' UNION SELECT 1, username, password FROM users --
```

The response will display the first matched row's `username` and `password` columns.

## Remediation (for reference)

Replace the vulnerable query with a parameterised statement:

```go
rows, err := db.Query(
    "SELECT id, username, email FROM users WHERE username = ? AND password = ?",
    username, password,
)
```

This ensures user-supplied values are treated as data, not SQL syntax.

## License

For educational purposes only.
