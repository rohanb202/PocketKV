# ---------- Build Stage ----------

FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o router ./cmd/router
RUN CGO_ENABLED=0 GOOS=linux go build -o node ./cmd/node


# ---------- Runtime Stage ----------

FROM alpine:latest

RUN apk add --no-cache wget

WORKDIR /app

COPY --from=builder /app/router .
COPY --from=builder /app/node .

EXPOSE 8080
EXPOSE 8081
EXPOSE 8082
EXPOSE 8083