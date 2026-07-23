# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o vulnapp .

# ── Run stage ─────────────────────────────────────────────────────────────────
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/vulnapp .
COPY templates/ templates/

EXPOSE 8080

CMD ["./vulnapp"]
