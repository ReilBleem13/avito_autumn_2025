# Avito-Autumn-2025

Сервис для управления pull requests.

### Запуск тестов
`make run-test` или `go test -tags=integration ./... -v`

### Запуск приложения
1. Скопируйте файл конфигурации:

```shell
cp .env-test .env
```
2. Запустите сервис через Docker:

```shell
make docker-up или docker compose up -d
```

Сервис будет доступен по адресу `localhost:8080`
___

### Стек приложения:
- Go 1.25
- PostgreSQL 16
- Docker 

### **Дополнительные задания**
- Реализовал интеграционное тестирование с помощью testcontainers.