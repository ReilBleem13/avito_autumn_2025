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

## Нагрузочное тестирование (k6)

`Running (1m54.3s), 00/12 VUs, 575 complete and 0 interrupted iterations`

#### Thresholds
```shell
    checks
    ✓ 'rate>0.999' rate=100.00%

    http_req_duration
    ✓ 'p(90)<300' p(90)=23.37ms

    http_req_failed
    ✓ 'rate<0.001' rate=0.00%
```

#### Total Results 

```shell
    checks_total.......: 4097    35.839119/s
    checks_succeeded...: 100.00% 4097 out of 4097
    checks_failed......: 0.00%   0 out of 4097
```

#### HTTP

```shell
    http_req_duration..............: avg=13.88ms min=1.97ms med=8.98ms max=435.79ms p(90)=23.37ms p(95)=32.37ms
    expected_response:true ...: avg=13.88ms min=1.97ms med=8.98ms max=435.79ms p(90)=23.37ms p(95)=32.37ms
    http_req_failed................: 0.00%  0 out of 1267
    http_reqs......................: 1267   11.083272
```

### **Дополнительные задания**
- Реализовал интеграционное тестирование с помощью testcontainers.